/*
 * Add students
 *
 * Copyright (C) 2024, 2025  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

func handleNewForcedChoices(w http.ResponseWriter, req *http.Request) (string, int, error) {
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

	req.ParseMultipartForm(0)
	file, fileHeader, err := req.FormFile("forcedchoicescsv")
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
	if len(titleLine) != 2 {
		return "", -1, wrapAny(
			errBadCSVFormat,
			"expecting 2 fields on the first line (Name, ID, Legal Sex)",
		)
	}
	var studentIDIndex, sectionIDIndex int = -1, -1
	for i, v := range titleLine {
		switch v {
		case "Student ID":
			studentIDIndex = i
		case "Section ID":
			sectionIDIndex = i
		}
	}

	if studentIDIndex == -1 {
		return "", http.StatusBadRequest, wrapAny(
			errMissingCSVColumn,
			"ID",
		)
	}
	if sectionIDIndex == -1 {
		return "", http.StatusBadRequest, wrapAny(
			errMissingCSVColumn,
			"Legal Sex",
		)
	}

	lineNumber := 1
	ok, statusCode, err := func(ctx context.Context) (
		retBool bool,
		retStatus int,
		retErr error,
	) {
		tx, err := db.Begin(ctx)
		if err != nil {
			return false, -1, wrapError(errUnexpectedDBError, err)
		}
		defer func() {
			err := tx.Rollback(ctx)
			if err != nil && (!errors.Is(err, pgx.ErrTxClosed)) {
				retBool, retStatus, retErr = false, -1, wrapError(
					errUnexpectedDBError,
					err,
				)
				return
			}
		}()
		_, err = tx.Exec(
			ctx,
			"DELETE FROM pre_selected",
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
				return false, -1, wrapError(
					errCannotReadCSV,
					err,
				)
			}
			if line == nil {
				return false, -1, wrapError(
					errCannotReadCSV,
					errUnexpectedNilCSVLine,
				)
			}
			if len(line) != 2 {
				return false, -1, wrapAny(
					errInsufficientFields,
					fmt.Sprintf(
						"line %d has a wrong number of items",
						lineNumber,
					),
				)
			}

			studentID, err := strconv.ParseInt(line[studentIDIndex], 10, 64)
			if err != nil {
				return false, -1, wrapAny(
					errBadCSVFormat,
					fmt.Sprintf(
						"line %d, ID is not a number; make sure that you only submit clean numbers e.g. 12345 as the student ID, don't use s12345/S12345",
						lineNumber,
					),
				)
			}

			sectionID := line[sectionIDIndex]

			_, err = tx.Exec(
				ctx,
				"INSERT INTO pre_selected(student_id, course_id) VALUES ($1, (SELECT id FROM courses WHERE section_id = $2))",
				studentID, sectionID,
			)
			if err != nil {
				return false, -1, fmt.Errorf("while inserting line %d: %w", lineNumber, err)
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

	http.Redirect(w, req, "/", http.StatusSeeOther)

	return "", -1, nil
}
