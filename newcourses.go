/*
 * Overwrite courses with uploaded CSV
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
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"

	"github.com/jackc/pgx/v5"
)

func handleNewCourses(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		wstr(
			w,
			http.StatusMethodNotAllowed,
			"Only POST is allowed here",
		)
		return
	}

	sessionCookie, err := req.Cookie("session")
	if errors.Is(err, http.ErrNoCookie) {
		wstr(
			w,
			http.StatusUnauthorized,
			"No session cookie, which is required for this endpoint",
		)
		return
	} else if err != nil {
		wstr(w, http.StatusBadRequest, "Error: Unable to check cookie.")
		return
	}

	var userDepartment string
	err = db.QueryRow(
		req.Context(),
		"SELECT department FROM users WHERE session = $1",
		sessionCookie.Value,
	).Scan(&userDepartment)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			wstr(
				w,
				http.StatusForbidden,
				"Invalid session cookie",
			)
			return
		}
		wstr(
			w,
			http.StatusInternalServerError,
			fmt.Sprintf(
				"Error: Unexpected database error: %s",
				err,
			),
		)
		return
	}

	if userDepartment != staffDepartment {
		wstr(
			w,
			http.StatusForbidden,
			"You are not authorized to view this page",
		)
		return
	}

	if atomic.LoadUint32(&state) != 0 {
		wstr(
			w,
			http.StatusBadRequest,
			"Uploading the course table is only supported when student-access is disabled",
		)
		return
	}
	/* TODO: Potential race. The global state may need to be write-locked. */

	file, fileHeader, err := req.FormFile("coursecsv")
	if err != nil {
		wstr(
			w,
			http.StatusBadRequest,
			"Failed loading file from request... did you select a file before hitting that red button?",
		)
		return
	}

	if fileHeader.Header.Get("Content-Type") != "text/csv" {
		wstr(
			w,
			http.StatusBadRequest,
			"Does not look like a proper CSV file",
		)
		return
	}

	csvReader := csv.NewReader(file)
	titleLine, err := csvReader.Read()
	if err != nil {
		wstr(
			w,
			http.StatusBadRequest,
			"Error reading CSV",
		)
		return
	}
	if titleLine == nil {
		wstr(
			w,
			http.StatusInternalServerError,
			"Unexpected nil titleLine slice",
		)
		return
	}
	if len(titleLine) != 6 {
		wstr(
			w,
			http.StatusBadRequest,
			"First line has more than 6 elements",
		)
		return
	}
	var titleIndex, maxIndex, teacherIndex, locationIndex,
		typeIndex, groupIndex int = -1, -1, -1, -1, -1, -1
	for i, v := range titleLine {
		switch v {
		case "Title":
			titleIndex = i
		case "Max":
			maxIndex = i
		case "Teacher":
			teacherIndex = i
		case "Location":
			locationIndex = i
		case "Type":
			typeIndex = i
		case "Group":
			groupIndex = i
		}
	}

	{
		check := func(indexName string, indexNum int) bool {
			if indexNum == -1 {
				wstr(
					w,
					http.StatusBadRequest,
					fmt.Sprintf(
						"Missing column \"%s\"",
						indexName,
					),
				)
				return true
			}
			return false
		}

		if check("Title", titleIndex) {
			return
		}
		if check("Max", maxIndex) {
			return
		}
		if check("Teacher", teacherIndex) {
			return
		}
		if check("Location", locationIndex) {
			return
		}
		if check("Type", typeIndex) {
			return
		}
		if check("Group", groupIndex) {
			return
		}
	}

	lineNumber := 1
	ok := func(ctx context.Context) bool {
		tx, err := db.Begin(ctx)
		if err != nil {
			wstr(
				w,
				http.StatusInternalServerError,
				"Unexpected database error",
			)
		}
		defer func() {
			err := tx.Rollback(ctx)
			if err != nil && (!errors.Is(err, pgx.ErrTxClosed)) {
				wstr(
					w,
					http.StatusInternalServerError,
					"Unexpected database error",
				)
				return
			}
		}()
		_, err = tx.Exec(
			ctx,
			"DELETE FROM choices",
		)
		if err != nil {
			wstr(
				w,
				http.StatusInternalServerError,
				"Unexpected database error",
			)
		}
		_, err = tx.Exec(
			ctx,
			"DELETE FROM courses",
		)
		if err != nil {
			wstr(
				w,
				http.StatusInternalServerError,
				"Unexpected database error",
			)
		}

		for {
			lineNumber++
			line, err := csvReader.Read()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				wstr(
					w,
					http.StatusInternalServerError,
					"Error reading CSV",
				)
				return false
			}
			if line == nil {
				wstr(
					w,
					http.StatusInternalServerError,
					"Unexpected nil line",
				)
				return false
			}
			if len(line) != 6 {
				wstr(
					w,
					http.StatusBadRequest,
					fmt.Sprintf(
						"Line %d has insufficient items",
						lineNumber,
					),
				)
				return false
			}
			_, err = tx.Exec(
				ctx,
				"INSERT INTO courses(nmax, title, teacher, location, ctype, cgroup) VALUES ($1, $2, $3, $4, $5, $6)",
				line[maxIndex],
				line[titleIndex],
				line[teacherIndex],
				line[locationIndex],
				line[typeIndex],
				line[groupIndex],
			)
			if err != nil {
				wstr(
					w,
					http.StatusInternalServerError,
					"Unexpected database error",
				)
				return false
			}
		}
		courses.Clear()
		err = setupCourses(ctx)
		if err != nil {
			wstr(
				w,
				http.StatusInternalServerError,
				"Error setting up course table again",
			)
			return false
		}
		err = tx.Commit(ctx)
		if err != nil {
			wstr(
				w,
				http.StatusInternalServerError,
				"Unexpected database error",
			)
			return false
		}
		return true
	}(req.Context())
	if !ok {
		return
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)
}
