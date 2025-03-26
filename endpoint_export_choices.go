/*
 * Staff page
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func handleExportChoices(
	w http.ResponseWriter,
	req *http.Request,
) (string, int, error) {
	_, _, department, err := getUserInfoFromRequest(req)
	if err != nil {
		return "", -1, err
	}
	if department != staffDepartment {
		return "", http.StatusForbidden, errors.New("staff only")
	}

	type userCacheT struct {
		Name       string
		StudentID  string
		Department string
	}
	userCacheMap := make(map[string]userCacheT)

	rows, err := db.Query(req.Context(), "SELECT userid, courseid FROM choices")
	if err != nil {
		return "", -1, fmt.Errorf("query choices: %w", err)
	}
	output := make([][]string, 0)
	for {
		if !rows.Next() {
			err := rows.Err()
			if err != nil {
				return "", -1, fmt.Errorf("read next choice: %w", err)
			}
			break
		}
		var currentUserID,
			currentUserName,
			currentStudentID,
			currentDepartment string
		var currentCourseID int
		err := rows.Scan(&currentUserID, &currentCourseID)
		if err != nil {
			return "", -1, fmt.Errorf("scan choice: %w", err)
		}
		currentUserCache, ok := userCacheMap[currentUserID]
		if ok {
			currentUserName = currentUserCache.Name
			currentDepartment = currentUserCache.Department
			currentStudentID = currentUserCache.StudentID
		} else {
			var currentUserEmail string
			err := db.QueryRow(
				req.Context(),
				"SELECT name, email, department FROM users WHERE id = $1",
				currentUserID,
			).Scan(
				&currentUserName,
				&currentUserEmail,
				&currentDepartment,
			)
			if err != nil {
				return "", -1, fmt.Errorf("scan user info: %w")
			}
			before, _, found := strings.Cut(currentUserEmail, "@")
			if found {
				currentStudentID, _ = strings.CutPrefix(returnFirst(strings.CutPrefix(before, "s")), "S")
			} else {
				currentStudentID = currentUserEmail
			}
			userCacheMap[currentUserID] = userCacheT{
				Name:       currentUserName,
				StudentID:  currentStudentID,
				Department: currentDepartment,
			}
		}

		_course, ok := courses.Load(currentCourseID)
		if !ok {
			return "", -1, fmt.Errorf("no such course")
		}
		course := _course.(*courseT)
		if course == nil {
			return "", -1, fmt.Errorf("no such course")
		}
		output = append(
			output,
			[]string{
				currentUserName,
				currentStudentID,
				currentDepartment,
				course.Title,
				course.Group,
				course.SectionID,
				course.CourseID,
			},
		)
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment;filename=cca_choices.csv")
	_, err = w.Write([]byte{0xEF, 0xBB, 0xBF})
	if err != nil {
		return "", -1, fmt.Errorf("write http stream: %w", err)
	}
	csvWriter := csv.NewWriter(w)
	err = csvWriter.Write([]string{
		"Student Name",
		"Student ID",
		"Grade/Year",
		"Group/Activity",
		"Container",
		"Section ID",
		"Course ID",
	})
	if err != nil {
		return "", -1, fmt.Errorf("write http stream: %w", err)
	}
	err = csvWriter.WriteAll(output)
	if err != nil {
		return "", -1, fmt.Errorf("write http stream: %w", err)
	}
	csvWriter.Flush()
	if csvWriter.Error() != nil {
		return "", -1, fmt.Errorf("write http stream: %w", err)
	}
	return "", -1, nil
}
