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
	"net/http"
	"strconv"
)

func handleExportStudents(w http.ResponseWriter, req *http.Request) (string, int, error) {
	_, _, department, err := getUserInfoFromRequest(req)
	if err != nil {
		return "", -1, err
	}
	if department != staffDepartment {
		return "", -1, errStaffOnly
	}

	rows, err := db.Query(req.Context(), "SELECT name, email, department, confirmed FROM users")
	if err != nil {
		return "", -1, wrapError(errUnexpectedDBError, err)
	}
	output := make([][]string, 0)
	for {
		if !rows.Next() {
			err := rows.Err()
			if err != nil {
				return "", -1, wrapError(errUnexpectedDBError, err)
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
			return "", -1, wrapError(errUnexpectedDBError, err)
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
		return "", -1, errHTTPWrite
	}
	err = csvWriter.WriteAll(output)
	if err != nil {
		return "", -1, errHTTPWrite
	}
	csvWriter.Flush()
	if csvWriter.Error() != nil {
		return "", -1, errHTTPWrite
	}

	return "", -1, nil
}
