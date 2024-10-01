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
	"time"

	"github.com/coder/websocket"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func messageHello(ctx context.Context, c *websocket.Conn, mar []string, userID string, session string) error {
	_, _, _ = mar, userID, session
	err := writeText(ctx, c, "HI")
	if err != nil {
		return fmt.Errorf("error replying to HELLO: %w", err)
	}
	return nil
}

func messageChooseCourse(ctx context.Context, c *websocket.Conn, mar []string, userID string, session string) error {
	_ = session
	if len(mar) != 2 {
		return protocolError(ctx, c, "Invalid number of arguments for Y")
	}
	_courseID, err := strconv.ParseInt(mar[1], 10, strconv.IntSize)
	if err != nil {
		return protocolError(ctx, c, "Course ID must be an integer")
	}
	courseID := int(_courseID)
	course := func() *courseT {
		coursesLock.RLock()
		defer coursesLock.RUnlock()
		return courses[courseID]
	}()

	err = func() (returnedError error) { /* Named returns so I could modify them in defer */
		tx, err := db.Begin(ctx)
		if err != nil {
			return protocolError(ctx, c, "Database error while beginning transaction")
		}
		defer func() {
			err := tx.Rollback(ctx)
			if err != nil && (!errors.Is(err, pgx.ErrTxClosed)) {
				returnedError = protocolError(ctx, c, "Database error while rolling back transaction in defer block")
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
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				err := writeText(ctx, c, "Y "+mar[1])
				if err != nil {
					return fmt.Errorf("error reaffirming course choice: %w", err)
				}
				return nil
			}
			return protocolError(ctx, c, "Database error while inserting course choice")
		}

		ok := func() bool {
			course.SelectedLock.Lock()
			defer course.SelectedLock.Unlock()
			if course.Selected < course.Max {
				course.Selected++
				go propagateIgnoreFailures(fmt.Sprintf("N %d %d", courseID, course.Selected))
				return true
			}
			return false
		}()

		if ok {
			err := tx.Commit(ctx)
			if err != nil {
				go func() { /* Separate goroutine because we don't need a response from this operation */
					course.SelectedLock.Lock()
					defer course.SelectedLock.Unlock()
					course.Selected--
					propagateIgnoreFailures(fmt.Sprintf("N %d %d", courseID, course.Selected))
				}()
				return protocolError(ctx, c, "Database error while committing transaction")
			}
			err = writeText(ctx, c, "Y "+mar[1])
			if err != nil {
				return fmt.Errorf("error affirming course choice: %w", err)
			}
		} else {
			err := tx.Rollback(ctx)
			if err != nil {
				return protocolError(ctx, c, "Database error while rolling back transaction due to course limit")
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

func messageUnchooseCourse(ctx context.Context, c *websocket.Conn, mar []string, userID string, session string) error {
	_ = session
	if len(mar) != 2 {
		return protocolError(ctx, c, "Invalid number of arguments for N")
	}
	_courseID, err := strconv.ParseInt(mar[1], 10, strconv.IntSize)
	if err != nil {
		return protocolError(ctx, c, "Course ID must be an integer")
	}
	courseID := int(_courseID)
	course := func() *courseT {
		coursesLock.RLock()
		defer coursesLock.RUnlock()
		return courses[courseID]
	}()

	ct, err := db.Exec(
		ctx,
		"DELETE FROM choices WHERE userid = $1 AND courseid = $2",
		userID,
		courseID,
	)
	if err != nil {
		return protocolError(ctx, c, "Database error while deleting course choice")
	}

	if ct.RowsAffected() != 0 {
		go func() { /* Separate goroutine because we don't need a response from this operation */
			course.SelectedLock.Lock()
			defer course.SelectedLock.Unlock()
			course.Selected--
			propagateIgnoreFailures(fmt.Sprintf("N %d %d", courseID, course.Selected))
		}()
	}

	err = writeText(ctx, c, "N "+mar[1])
	if err != nil {
		return fmt.Errorf("error replying that course has been deselected: %w", err)
	}

	return nil
}
