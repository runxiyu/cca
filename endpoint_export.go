/*
 * Staff page
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
	"encoding/csv"
	"fmt"
	"net/http"
	"strings"
)

func handleExport(w http.ResponseWriter, req *http.Request) {
	_, _, department, err := getUserInfoFromRequest(req)
	if err != nil {
		wstr(
			w,
			http.StatusInternalServerError,
			fmt.Sprintf("Error: %v", err),
		)
	}
	if department != staffDepartment {
		wstr(
			w,
			http.StatusForbidden,
			"You are not authorized to view this page",
		)
		return
	}

	type userCacheT struct {
		Name       string
		StudentID  string
		Department string
	}
	userCacheMap := make(map[string]userCacheT)

	rows, err := db.Query(req.Context(), "SELECT userid, courseid FROM choices")
	if err != nil {
		wstr(
			w,
			http.StatusInternalServerError,
			"Unexpected database error",
		)
		return
	}
	output := make([][]string, 0)
	for {
		if !rows.Next() {
			err := rows.Err()
			if err != nil {
				wstr(
					w,
					http.StatusInternalServerError,
					"Unexpected database error",
				)
				return
			}
			break
		}
		var currentUserID, currentUserName, currentStudentID, currentDepartment string
		var currentCourseID int
		err := rows.Scan(&currentUserID, &currentCourseID)
		if err != nil {
			wstr(
				w,
				http.StatusInternalServerError,
				"Unexpected database error",
			)
			return
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
				wstr(
					w,
					http.StatusInternalServerError,
					"Unexpected database error",
				)
				return
			}
			before, _, found := strings.Cut(currentUserEmail, "@")
			if found {
				currentStudentID = before
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
			wstr(
				w,
				http.StatusInternalServerError,
				"Reference to non-existent course",
			)
			return
		}
		course, ok := _course.(*courseT)
		if !ok {
			panic("courses map has non-\"*courseT\" items")
		}
		if course == nil {
			wstr(
				w,
				http.StatusInternalServerError,
				"Course is nil",
			)
			return
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
	w.Header().Set("Content-Disposition", "attachment;filename=cca.csv")
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
		wstr(
			w,
			http.StatusInternalServerError,
			"Error writing output",
		)
		return
	}
	err = csvWriter.WriteAll(output)
	if err != nil {
		wstr(
			w,
			http.StatusInternalServerError,
			"Error writing output",
		)
		return
	}
	csvWriter.Flush()
	if csvWriter.Error() != nil {
		wstr(
			w,
			http.StatusInternalServerError,
			"Error occurred flushing output",
		)
		return
	}
}
