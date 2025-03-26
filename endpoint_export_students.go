/*
 * Staff page
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"encoding/csv"
	"fmt"
	"net/http"
	"strconv"
)

func handleExportStudents(
	w http.ResponseWriter,
	req *http.Request,
) (string, int, error) {
	_, _, department, err := getUserInfoFromRequest(req)
	if err != nil {
		return "", -1, err
	}
	if department != staffDepartment {
		return "", -1, errStaffOnly
	}

	rows, err := db.Query(
		req.Context(),
		"SELECT name, email, department, confirmed FROM users",
	)
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

	w.Header().Set(
		"Content-Type",
		"text/csv; charset=utf-8",
	)
	w.Header().Set(
		"Content-Disposition",
		"attachment;filename=cca_students.csv",
	)
	_, err = w.Write([]byte{0xEF, 0xBB, 0xBF})
	if err != nil {
		return "", -1, fmt.Errorf("write http stream: %w", err)
	}
	csvWriter := csv.NewWriter(w)
	err = csvWriter.Write([]string{
		"Student Name",
		"Student ID",
		"Grade/Year",
		"Confirmed",
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
