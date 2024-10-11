/*
 * Course data structures and locking
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
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/coder/websocket"
)

type (
	courseTypeT  string
	courseGroupT string
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
	Type         courseTypeT
	Group        courseGroupT
	Teacher      string
	Location     string
	Usems        sync.Map /* string, *usemT */
}

const (
	sport    courseTypeT = "Sport"
	nonSport courseTypeT = "Non-sport"
)

var courseTypes = map[courseTypeT]struct{}{
	sport:    {},
	nonSport: {},
}

const (
	mw1 courseGroupT = "MW1"
	mw2 courseGroupT = "MW2"
	mw3 courseGroupT = "MW3"
	tt1 courseGroupT = "TT1"
	tt2 courseGroupT = "TT2"
	tt3 courseGroupT = "TT3"
)

var courseGroups = map[courseGroupT]string{
	mw1: "Monday/Wednesday CCA1",
	mw2: "Monday/Wednesday CCA2",
	mw3: "Monday/Wednesday CCA3",
	tt1: "Tuesday/Thursday CCA1",
	tt2: "Tuesday/Thursday CCA2",
	tt3: "Tuesday/Thursday CCA3",
}

func checkCourseType(ct courseTypeT) bool {
	_, ok := courseTypes[ct]
	return ok
}

func checkCourseGroup(cg courseGroupT) bool {
	_, ok := courseGroups[cg]
	return ok
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
		"SELECT id, nmax, title, ctype, cgroup, teacher, location FROM courses",
	)
	if err != nil {
		return fmt.Errorf("%w: %w", errUnexpectedDBError, err)
	}

	for {
		if !rows.Next() {
			err := rows.Err()
			if err != nil {
				return fmt.Errorf(
					"%w: %w",
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
		)
		if err != nil {
			return fmt.Errorf("%w: %w", errUnexpectedDBError, err)
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
			return fmt.Errorf(
				"%w: %w",
				errUnexpectedDBError,
				err,
			)
		}
		courses.Store(currentCourse.ID, &currentCourse)
		atomic.AddUint32(&numCourses, 1)
	}

	return nil
}

type userCourseGroupsT map[courseGroupT]struct{}

func populateUserCourseGroups(
	ctx context.Context,
	userCourseGroups *userCourseGroupsT,
	userID string,
) error {
	rows, err := db.Query(
		ctx,
		"SELECT courseid FROM choices WHERE userid = $1",
		userID,
	)
	if err != nil {
		return fmt.Errorf(
			"%w: %w",
			errUnexpectedDBError,
			err,
		)
	}
	for {
		if !rows.Next() {
			err := rows.Err()
			if err != nil {
				return fmt.Errorf(
					"%w: %w",
					errUnexpectedDBError,
					err,
				)
			}
			break
		}
		var thisCourseID int
		err := rows.Scan(&thisCourseID)
		if err != nil {
			return fmt.Errorf(
				"%w: %w",
				errUnexpectedDBError,
				err,
			)
		}
		var thisGroupName courseGroupT
		_course, ok := courses.Load(thisCourseID)
		if !ok {
			return fmt.Errorf(
				"%w: %d",
				errNoSuchCourse,
				thisCourseID,
			)
		}
		course, ok := _course.(*courseT)
		if !ok {
			panic("courses map has non-\"*courseT\" items")
		}
		thisGroupName = course.Group
		if _, ok := (*userCourseGroups)[thisGroupName]; ok {
			return fmt.Errorf(
				"%w: user %v, group %v",
				errMultipleChoicesInOneGroup,
				userID,
				thisGroupName,
			)
		}
		(*userCourseGroups)[thisGroupName] = struct{}{}
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
	go propagateSelectedUpdate(course)
	err := sendSelectedUpdate(ctx, conn, course.ID)
	if err != nil {
		return fmt.Errorf(
			"%w: %w",
			errCannotSend,
			err,
		)
	}
	return nil
}
