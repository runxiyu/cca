/*
 * Handle the "N" message for unchoosing a course
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
	"strconv"
	"sync/atomic"

	"github.com/coder/websocket"
)

func messageUnchooseCourse(
	ctx context.Context,
	c *websocket.Conn,
	reportError reportErrorT,
	mar []string,
	userID string,
	userCourseGroups *userCourseGroupsT,
	userCourseTypes *userCourseTypesT,
) error {
	if atomic.LoadUint32(&state) != 2 {
		err := writeText(ctx, c, "E :Course selections are not open")
		if err != nil {
			return wrapError(
				errCannotSend,
				err,
			)
		}
		return nil
	}

	select {
	case <-ctx.Done():
		return wrapError(
			errContextCanceled,
			ctx.Err(),
		)
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

	_course, ok := courses.Load(courseID)
	if !ok {
		return reportError("no such course")
	}
	course, ok := _course.(*courseT)
	if !ok {
		panic("courses map has non-\"*courseT\" items")
	}
	if course == nil {
		return reportError("couse is nil")
	}

	ct, err := db.Exec(
		ctx,
		"DELETE FROM choices WHERE userid = $1 AND courseid = $2",
		userID,
		courseID,
	)
	if err != nil {
		return reportError(
			"Database error while deleting course choice",
		)
	}

	if ct.RowsAffected() != 0 {
		err := course.decrementSelectedAndPropagate(ctx, c)
		if err != nil {
			return wrapError(
				errCannotSend,
				err,
			)
		}

		_course, ok := courses.Load(courseID)
		if !ok {
			return reportError("no such course")
		}
		course, ok := _course.(*courseT)
		if !ok {
			panic("courses map has non-\"*courseT\" items")
		}
		if course == nil {
			return reportError("couse is nil")
		}

		if _, ok := (*userCourseGroups)[course.Group]; !ok {
			return reportError("inconsistent user course groups")
		}
		delete(*userCourseGroups, course.Group)
		(*userCourseTypes)[course.Type]--
	}

	err = writeText(ctx, c, "N "+mar[1])
	if err != nil {
		return wrapError(
			errCannotSend,
			err,
		)
	}

	return nil
}
