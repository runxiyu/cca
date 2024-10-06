/*
 * WebSocket message handlers
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/coder/websocket"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func messageHello(ctx context.Context, c *websocket.Conn, reportError reportErrorT, mar []string, userID string, session string) error {
	_, _ = mar, session

	select {
	case <-ctx.Done():
		return fmt.Errorf("context done when handling hello: %w", ctx.Err())
	default:
	}

	rows, err := db.Query(
		ctx,
		"SELECT courseid FROM choices WHERE userid = $1",
		userID,
	)
	if err != nil {
		return reportError("error fetching choices")
	}
	courseIDs, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return reportError("error collecting choices")
	}

	err = writeText(ctx, c, "HI :"+strings.Join(courseIDs, ","))
	if err != nil {
		return fmt.Errorf("error replying to HELLO: %w", err)
	}

	return nil
}

func messageChooseCourse(ctx context.Context, c *websocket.Conn, reportError reportErrorT, mar []string, userID string, session string, userCourseGroups *userCourseGroupsT) error {
	_ = session

	select {
	case <-ctx.Done():
		return fmt.Errorf("context done when handling choose: %w", ctx.Err())
	default:
	}

	if len(mar) != 2 {
		return reportError("Invalid number of arguments for Y")
	}
	_courseID, err := strconv.ParseInt(mar[1], 10, strconv.IntSize)
	if err != nil {
		return reportError("Course ID must be an integer")
	}
	courseID := int(_courseID)
	course := getCourseByID(courseID)

	if course == nil {
		return reportError("nil course")
	}

	err = func() (returnedError error) { /* Named returns so I could modify them in defer */
		tx, err := db.Begin(ctx)
		if err != nil {
			return reportError("Database error while beginning transaction")
		}
		defer func() {
			err := tx.Rollback(ctx)
			if err != nil && (!errors.Is(err, pgx.ErrTxClosed)) {
				returnedError = reportError("Database error while rolling back transaction in defer block")
				return
			}
		}()

		_, err = tx.Exec(
			ctx,
			"INSERT INTO choices (seltime, userid, courseid) VALUES ($1, $2, $3)",
			time.Now().UnixMicro(),
			userID,
			courseID,
		)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgErrUniqueViolation {
				err := writeText(ctx, c, "Y "+mar[1])
				if err != nil {
					return fmt.Errorf("error reaffirming course choice: %w", err)
				}
				return nil
			}
			return reportError("Database error while inserting course choice")
		}

		ok := func() bool {
			course.SelectedLock.Lock()
			defer course.SelectedLock.Unlock()
			if course.Selected < course.Max {
				course.Selected++
				return true
			}
			return false
		}()

		if ok {
			go propagateSelectedUpdate(courseID)
			err := tx.Commit(ctx)
			if err != nil {
				err := course.decrementSelectedAndPropagate(ctx, c)
				if err != nil {
					return fmt.Errorf("error decrementing and notifying: %w", err)
				}
				return reportError("Database error while committing transaction")
			}
			var thisCourseGroup courseGroupT
			func() {
				coursesLock.RLock()
				defer coursesLock.RUnlock()
				thisCourseGroup = courses[courseID].Group
			}()
			if (*userCourseGroups)[thisCourseGroup] {
				err := course.decrementSelectedAndPropagate(ctx, c)
				if err != nil {
					return fmt.Errorf("error decrementing and notifying: %w", err)
				}
				return reportError("inconsistent user course groups")
			}
			(*userCourseGroups)[thisCourseGroup] = true
			err = writeText(ctx, c, "Y "+mar[1])
			if err != nil {
				return fmt.Errorf("error affirming course choice: %w", err)
			}
			err = sendSelectedUpdate(ctx, c, courseID)
			if err != nil {
				return fmt.Errorf("error notifying after increment: %w", err)
			}
		} else {
			err := tx.Rollback(ctx)
			if err != nil {
				return reportError("Database error while rolling back transaction due to course limit")
			}
			err = writeText(ctx, c, "R "+mar[1]+" :Full")
			if err != nil {
				return fmt.Errorf("error rejecting course choice: %w", err)
			}
		}
		return nil
	}()
	if err != nil {
		return err
	}
	return nil
}

func messageUnchooseCourse(ctx context.Context, c *websocket.Conn, reportError reportErrorT, mar []string, userID string, session string, userCourseGroups *userCourseGroupsT) error {
	_ = session

	select {
	case <-ctx.Done():
		return fmt.Errorf("context done when handling unchoose: %w", ctx.Err())
	default:
	}

	if len(mar) != 2 {
		return reportError("Invalid number of arguments for N")
	}
	_courseID, err := strconv.ParseInt(mar[1], 10, strconv.IntSize)
	if err != nil {
		return reportError("Course ID must be an integer")
	}
	courseID := int(_courseID)
	course := getCourseByID(courseID)

	if course == nil {
		return reportError("nil course")
	}

	ct, err := db.Exec(
		ctx,
		"DELETE FROM choices WHERE userid = $1 AND courseid = $2",
		userID,
		courseID,
	)
	if err != nil {
		return reportError("Database error while deleting course choice")
	}

	if ct.RowsAffected() != 0 {
		err := course.decrementSelectedAndPropagate(ctx, c)
		if err != nil {
			return fmt.Errorf("error decrementing and notifying: %w", err)
		}
		var thisCourseGroup courseGroupT
		func() {
			coursesLock.RLock()
			defer coursesLock.RUnlock()
			thisCourseGroup = courses[courseID].Group
		}()
		if !(*userCourseGroups)[thisCourseGroup] {
			return reportError("inconsistent user course groups")
		}
		(*userCourseGroups)[thisCourseGroup] = false
	}

	err = writeText(ctx, c, "N "+mar[1])
	if err != nil {
		return fmt.Errorf("error replying that course has been deselected: %w", err)
	}

	return nil
}
