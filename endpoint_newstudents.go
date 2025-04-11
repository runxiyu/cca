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

func handleNewStudents(w http.ResponseWriter, req *http.Request) (string, int, error) {
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

	file, fileHeader, err := req.FormFile("studentscsv")
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
	if len(titleLine) != 3 {
		return "", -1, wrapAny(
			errBadCSVFormat,
			"expecting 3 fields on the first line (Name, ID, Legal Sex)",
		)
	}
	var nameIndex, idIndex, legalSexIndex int = -1, -1, -1
	for i, v := range titleLine {
		switch v {
		case "Name":
			nameIndex = i
		case "ID":
			idIndex = i
		case "Legal Sex":
			legalSexIndex = i
		}
	}

	if nameIndex == -1 {
		return "", http.StatusBadRequest, wrapAny(
			errMissingCSVColumn,
			"Name",
		)
	}
	if idIndex == -1 {
		return "", http.StatusBadRequest, wrapAny(
			errMissingCSVColumn,
			"ID",
		)
	}
	if legalSexIndex == -1 {
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
			"DELETE FROM expected_students",
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
			if len(line) != 3 {
				return false, -1, wrapAny(
					errInsufficientFields,
					fmt.Sprintf(
						"line %d has a wrong number of items",
						lineNumber,
					),
				)
			}

			id, err := strconv.ParseInt(line[idIndex], 10, 64)
			if err != nil {
				return false, -1, wrapAny(
					errBadCSVFormat,
					fmt.Sprintf(
						"line %d, ID is not a number; make sure that you only submit clean numbers e.g. 12345 as the student ID, don't use s12345/S12345",
						lineNumber,
					),
				)
			}

			name := line[nameIndex]
			legalSex := line[legalSexIndex]

			_, err = tx.Exec(
				ctx,
				"INSERT INTO expected_students(name, id, legal_sex) VALUES ($1, $2, $3)",
				name, id, legalSex,
			)
			if err != nil {
				return false, -1, wrapError(
					errUnexpectedDBError,
					err,
				)
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

func queryNameID(ctx context.Context, query string, args ...any) (result map[int64]string, err error) {
	result = make(map[int64]string)
	var rows pgx.Rows

	if rows, err = db.Query(ctx, query, args...); err != nil {
		return nil, wrapError(errUnexpectedDBError, err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		var id int64
		if err = rows.Scan(&name, &id); err != nil {
			return nil, wrapError(errUnexpectedDBError, err)
		}
		result[id] = name
	}
	return result, wrapError(errUnexpectedDBError, rows.Err())
}
