/*
 * Course types and groups
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"context"
	"fmt"
)

/* Course types, e.g. Sport */

const (
	sport    string = "Sport"
	nonSport string = "Non-sport"
)

var courseTypes = map[string]struct{}{
	sport:    {},
	nonSport: {},
}

func checkCourseType(ct string) bool {
	_, ok := courseTypes[ct]
	return ok
}

type userCourseTypesT map[string]int

func getCourseTypeMinimumForYearGroup(yearGroup, courseType string) (int, error) {
	switch yearGroup {
	case "Y9":
		switch courseType {
		case sport:
			return config.Req.Y9.Sport, nil
		case nonSport:
			return config.Req.Y9.NonSport, nil
		default:
			return 0, fmt.Errorf("invalid course type: %v", courseType)
		}
	case "Y10":
		switch courseType {
		case sport:
			return config.Req.Y10.Sport, nil
		case nonSport:
			return config.Req.Y10.NonSport, nil
		default:
			return 0, fmt.Errorf("invalid course type: %v", courseType)
		}
	case "Y11":
		switch courseType {
		case sport:
			return config.Req.Y11.Sport, nil
		case nonSport:
			return config.Req.Y11.NonSport, nil
		default:
			return 0, fmt.Errorf("invalid course type: %v", courseType)
		}
	case "Y12":
		switch courseType {
		case sport:
			return config.Req.Y12.Sport, nil
		case nonSport:
			return config.Req.Y12.NonSport, nil
		default:
			return 0, fmt.Errorf("invalid course type: %v", courseType)
		}
	default:
		return 0, fmt.Errorf("invalid year group: %v", yearGroup)
	}
}

/* Course groups, e.g. MW1 */

type userCourseGroupsT map[string]struct{}

func checkCourseGroup(cg string) bool {
	_, ok := courseGroups[cg]
	return ok
}

const (
	mw1 string = "MW1"
	mw2 string = "MW2"
	mw3 string = "MW3"
	tt1 string = "TT1"
	tt2 string = "TT2"
	tt3 string = "TT3"
)

var courseGroups = map[string]string{
	mw1: "Monday/Wednesday CCA1",
	mw2: "Monday/Wednesday CCA2",
	mw3: "Monday/Wednesday CCA3",
	tt1: "Tuesday/Thursday CCA1",
	tt2: "Tuesday/Thursday CCA2",
	tt3: "Tuesday/Thursday CCA3",
}

/* Populate both */

func populateUserCourseTypesAndGroups(
	ctx context.Context,
	userCourseTypes *userCourseTypesT,
	userCourseGroups *userCourseGroupsT,
	userID string,
) error {
	rows, err := db.Query(
		ctx,
		"SELECT courseid FROM choices WHERE userid = $1",
		userID,
	)
	if err != nil {
		return fmt.Errorf("get user choices: %w", err)
	}
	for {
		if !rows.Next() {
			err := rows.Err()
			if err != nil {
				return fmt.Errorf("read next user choice: %w", err)
			}
			break
		}
		var thisCourseID int
		err := rows.Scan(&thisCourseID)
		if err != nil {
			return fmt.Errorf("scan user choice: %w", err)
		}
		var thisGroupName, thisTypeName string
		_course, ok := courses.Load(thisCourseID)
		if !ok {
			return fmt.Errorf("unknown course in user choice: %v", thisCourseID)
		}
		course := _course.(*courseT)
		thisGroupName = course.Group
		thisTypeName = course.Type
		if _, ok := (*userCourseGroups)[thisGroupName]; ok {
			return fmt.Errorf("duplicate group in user choices: user %v", userID)
		}
		(*userCourseGroups)[thisGroupName] = struct{}{}
		(*userCourseTypes)[thisTypeName]++
	}
	return nil
}
