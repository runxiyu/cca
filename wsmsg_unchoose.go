/*
 * Handle the "N" message for unchoosing a course
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
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
	mar []string,
	userID string,
	yeargroup string,
	userCourseGroups *userCourseGroupsT,
	userCourseTypes *userCourseTypesT,
) error {
	_state, ok := states[yeargroup]
	if !ok {
		return errNoSuchYearGroup
	}
	if atomic.LoadUint32(_state) != 2 {
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
			errWsHandlerContextCanceled,
			ctx.Err(),
		)
	default:
	}

	if len(mar) != 2 {
		return errBadNumberOfArguments
	}
	_courseID, err := strconv.ParseInt(mar[1], 10, strconv.IntSize)
	if err != nil {
		return errNoSuchCourse
	}
	courseID := int(_courseID)

	_course, ok := courses.Load(courseID)
	if !ok {
		return errNoSuchCourse
	}
	course, ok := _course.(*courseT)
	if !ok {
		panic("courses map has non-\"*courseT\" items")
	}
	if course == nil {
		return errNoSuchCourse
	}

	ct, err := db.Exec(
		ctx,
		"DELETE FROM choices WHERE userid = $1 AND courseid = $2",
		userID,
		courseID,
	)
	if err != nil {
		return wrapError(errUnexpectedDBError, err)
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
			return errNoSuchCourse
		}
		course, ok := _course.(*courseT)
		if !ok {
			panic("courses map has non-\"*courseT\" items")
		}
		if course == nil {
			return errNoSuchCourse
		}

		if _, ok := (*userCourseGroups)[course.Group]; !ok {
			return errCourseGroupHandlingError
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
