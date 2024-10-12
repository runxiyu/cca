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
	"strconv"
)

func handleExportStudents(w http.ResponseWriter, req *http.Request) {
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

	rows, err := db.Query(req.Context(), "SELECT name, email, department, confirmed FROM users")
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
		var currentUserName, currentEmail, currentDepartment string
		var currentConfirmed bool
		err := rows.Scan(
			&currentUserName,
			&currentEmail,
			&currentDepartment,
			&currentConfirmed,
		)
		if err != nil {
			wstr(
				w,
				http.StatusInternalServerError,
				"Unexpected database error",
			)
			return
		}

		if currentDepartment == staffDepartment {
			continue
		}

		output = append(
			output,
			[]string{
				currentUserName,
				currentEmail,
				currentDepartment,
				strconv.FormatBool(currentConfirmed),
			},
		)
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment;filename=cca_students.csv")
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
