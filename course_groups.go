/*
 * Course groups
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
)

type userCourseGroupsT map[courseGroupT]struct{}

type courseGroupT string

func checkCourseGroup(cg courseGroupT) bool {
	_, ok := courseGroups[cg]
	return ok
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
		return wrapError(
			errUnexpectedDBError,
			err,
		)
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
		var thisCourseID int
		err := rows.Scan(&thisCourseID)
		if err != nil {
			return wrapError(
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
