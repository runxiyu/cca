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
	"strings"
	"sync/atomic"

	"github.com/jackc/pgx/v5"
)

func handleNewCourses(w http.ResponseWriter, req *http.Request) (string, int, error) {
	if req.Method != http.MethodPost {
		return "", http.StatusMethodNotAllowed, errPostOnly
	}

	_, _, department, err := getUserInfoFromRequest(req)
	if err != nil {
		return "", -1, err
	}
	if department != staffDepartment {
		return "", http.StatusForbidden, errStaffOnly
	}

	if atomic.LoadUint32(&state) != 0 {
		return "", http.StatusBadRequest, errDisableStudentAccessFirst
	}

	/* TODO: Potential race. The global state may need to be write-locked. */

	file, fileHeader, err := req.FormFile("coursecsv")
	if err != nil {
		return "", http.StatusBadRequest, wrapError(errFormNoFile, err)
	}

	if fileHeader.Header.Get("Content-Type") != "text/csv" {
		return "", http.StatusBadRequest, errNotACSV
	}

	csvReader := csv.NewReader(file)
	titleLine, err := csvReader.Read()
	if err != nil {
		return "", http.StatusBadRequest, wrapError(errCannotReadCSV, err)
	}
	if titleLine == nil {
		return "", -1, errUnexpectedNilCSVLine
	}
	if len(titleLine) != 8 {
		return "", -1, wrapAny(errBadCSVFormat, "expecting 8 fields on the first line")
	}
	var titleIndex, maxIndex, teacherIndex, locationIndex,
		typeIndex, groupIndex, sectionIDIndex,
		courseIDIndex int = -1, -1, -1, -1, -1, -1, -1, -1
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
		case "Section ID":
			sectionIDIndex = i
		case "Course ID":
			courseIDIndex = i
		}
	}

	if titleIndex == -1 {
		return "", http.StatusBadRequest, wrapAny(errMissingCSVColumn, "Title")
	}
	if maxIndex == -1 {
		return "", http.StatusBadRequest, wrapAny(errMissingCSVColumn, "Max")
	}
	if teacherIndex == -1 {
		return "", http.StatusBadRequest, wrapAny(errMissingCSVColumn, "Teacher")
	}
	if locationIndex == -1 {
		return "", http.StatusBadRequest, wrapAny(errMissingCSVColumn, "Location")
	}
	if typeIndex == -1 {
		return "", http.StatusBadRequest, wrapAny(errMissingCSVColumn, "Type")
	}
	if groupIndex == -1 {
		return "", http.StatusBadRequest, wrapAny(errMissingCSVColumn, "Group")
	}
	if courseIDIndex == -1 {
		return "", http.StatusBadRequest, wrapAny(errMissingCSVColumn, "Course ID")
	}
	if sectionIDIndex == -1 {
		return "", http.StatusBadRequest, wrapAny(errMissingCSVColumn, "Section ID")
	}

	lineNumber := 1
	ok, statusCode, err := func(ctx context.Context) (retBool bool, retStatus int, retErr error) {
		tx, err := db.Begin(ctx)
		if err != nil {
			return false, -1, wrapError(errUnexpectedDBError, err)
		}
		defer func() {
			err := tx.Rollback(ctx)
			if err != nil && (!errors.Is(err, pgx.ErrTxClosed)) {
				retBool, retStatus, retErr = false, -1, wrapError(errUnexpectedDBError, err)
				return
			}
		}()
		_, err = tx.Exec(
			ctx,
			"DELETE FROM choices",
		)
		if err != nil {
			return false, -1, wrapError(errUnexpectedDBError, err)
		}
		_, err = tx.Exec(
			ctx,
			"DELETE FROM courses",
		)
		if err != nil {
			return false, -1, wrapError(errUnexpectedDBError, err)
		}

		for {
			lineNumber++
			line, err := csvReader.Read()
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				return false, -1, wrapError(errCannotReadCSV, err)
			}
			if line == nil {
				return false, -1, wrapError(errCannotReadCSV, errUnexpectedNilCSVLine)
			}
			if len(line) != 8 {
				return false, -1, wrapAny(errInsufficientFields, fmt.Sprintf(
					"line %d has insufficient items",
					lineNumber,
				))
			}
			if !checkCourseType(line[typeIndex]) {
				return false, -1, wrapAny(errInvalidCourseType,
					fmt.Sprintf(
						"line %d has invalid course type \"%s\"\nallowed course types: %s",
						lineNumber,
						line[typeIndex],
						strings.Join(getKeysOfMap(courseTypes), ", "),
					),
				)
			}
			if !checkCourseGroup(line[groupIndex]) {
				return false, -1, wrapAny(errInvalidCourseGroup,
					fmt.Sprintf(
						"line %d has invalid course group \"%s\"\nallowed course groups: %s",
						lineNumber,
						line[groupIndex],
						strings.Join(getKeysOfMap(courseGroups), ", "),
					),
				)
			}
			_, err = tx.Exec(
				ctx,
				"INSERT INTO courses(nmax, title, teacher, location, ctype, cgroup, section_id, course_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)",
				line[maxIndex],
				line[titleIndex],
				line[teacherIndex],
				line[locationIndex],
				line[typeIndex],
				line[groupIndex],
				line[sectionIDIndex],
				line[courseIDIndex],
			)
			if err != nil {
				return false, -1, wrapError(errUnexpectedDBError, err)
			}
		}
		err = tx.Commit(ctx)
		if err != nil {
			return false, -1, wrapError(errUnexpectedDBError, err)
		}
		return true, -1, nil
	}(req.Context())
	if !ok {
		return "", statusCode, err
	}

	courses.Clear()
	err = setupCourses(req.Context())
	if err != nil {
		return "", -1, wrapError(errWhileSetttingUpCourseTablesAgain, err)
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)

	return "", -1, nil
}
