/*
 * Course data structures and locking
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/coder/websocket"
)

type courseT struct {
	/*
	 * Selected is usually accessed atomically, but a lock is still
	 * necessary as we need to sequentialize compare-with-Max-and-increment
	 * operations.
	 * We put Selected before other values to ensure 64-bit alignment on
	 * all systems, because it needs to be accessed atomically. See the
	 * "Bugs" section of sync/atomic.
	 */
	Selected     uint32 /* atomic */
	SelectedLock sync.Mutex
	ID           int
	Max          uint32
	Title        string
	Type         string
	Group        string
	Teacher      string
	Location     string
	CourseID     string
	SectionID    string
	YearGroups   uint8
	Usems        sync.Map /* string, *usemT */
}

var courses sync.Map /* int, *courseT */

var numCourses uint32 /* atomic */

const staffDepartment = "Staff"

/*
 * Read course information from the database. This should be called during
 * setup.
 */
func setupCourses(ctx context.Context) error {
	rows, err := db.Query(
		ctx,
		"SELECT id, nmax, title, ctype, cgroup, teacher, location, course_id, section_id, year_groups FROM courses",
	)
	if err != nil {
		return wrapError(errUnexpectedDBError, err)
	}

	for {
		if !rows.Next() {
			err := rows.Err()
			if err != nil {
				return wrapError(
					errUnexpectedDBError,
					err,
				)
			}
			break
		}
		currentCourse := courseT{} //exhaustruct:ignore
		err = rows.Scan(
			&currentCourse.ID,
			&currentCourse.Max,
			&currentCourse.Title,
			&currentCourse.Type,
			&currentCourse.Group,
			&currentCourse.Teacher,
			&currentCourse.Location,
			&currentCourse.CourseID,
			&currentCourse.SectionID,
			&currentCourse.YearGroups,
		)
		if err != nil {
			return wrapError(errUnexpectedDBError, err)
		}
		if !checkCourseType(currentCourse.Type) {
			return fmt.Errorf(
				"%w: %d %s",
				errInvalidCourseType,
				currentCourse.ID,
				currentCourse.Type,
			)
		}
		if !checkCourseGroup(currentCourse.Group) {
			return fmt.Errorf(
				"%w: %d %s",
				errInvalidCourseGroup,
				currentCourse.ID,
				currentCourse.Group,
			)
		}
		err := db.QueryRow(
			ctx,
			"SELECT COUNT (*) FROM choices WHERE courseid = $1",
			currentCourse.ID,
		).Scan(&currentCourse.Selected)
		if err != nil {
			return wrapError(
				errUnexpectedDBError,
				err,
			)
		}
		courses.Store(currentCourse.ID, &currentCourse)
		atomic.AddUint32(&numCourses, 1)
	}

	return nil
}

func (course *courseT) decrementSelectedAndPropagate(
	ctx context.Context,
	conn *websocket.Conn,
) error {
	func() {
		course.SelectedLock.Lock()
		defer course.SelectedLock.Unlock()
		atomic.AddUint32(&course.Selected, ^uint32(0))
	}()
	go func() {
		defer func() {
			if e := recover(); e != nil {
				slog.Error("panic", "arg", e)
			}
		}()
		propagateSelectedUpdate(course)
	}()
	err := sendSelectedUpdate(ctx, conn, course.ID)
	if err != nil {
		return wrapError(
			errCannotSend,
			err,
		)
	}
	return nil
}

var yearGroupsNumberBits = map[string]uint8{"Y9": 1, "Y10": 2, "Y11": 4, "Y12": 8}

func yearGroupsStringToNumber(s string) (uint8, error) {
	var spec uint8
	if s == "" {
		for _, v := range yearGroupsNumberBits {
			spec |= v
		}
		return spec, nil
	}
	ss := strings.Split(s, " ")
	for _, yg := range ss {
		v, ok := yearGroupsNumberBits[yg]
		if !ok {
			return spec, wrapAny(errYearGroupSpecString, s)
		}
		spec |= v
	}
	return spec, nil
}
