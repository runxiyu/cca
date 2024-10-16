/*
 * Error definitions
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
	"errors"
	"fmt"
)

var (
	errCannotSetupJwks                  = errors.New("cannot set up jwks")
	errInsufficientFields               = errors.New("insufficient fields")
	errUnknownDepartment                = errors.New("unknown department")
	errCannotProcessConfig              = errors.New("cannot process configuration file")
	errCannotOpenConfig                 = errors.New("cannot open configuration file")
	errCannotDecodeConfig               = errors.New("cannot decode configuration file")
	errMissingConfigValue               = errors.New("missing configuration value")
	errInvalidCourseType                = errors.New("invalid course type")
	errInvalidCourseGroup               = errors.New("invalid course group")
	errMultipleChoicesInOneGroup        = errors.New("multiple choices per group per user")
	errCourseGroupHandlingError         = errors.New("error handling course group")
	errUnsupportedDatabaseType          = errors.New("unsupported db type")
	errUnexpectedDBError                = errors.New("unexpected database error")
	errCannotSend                       = errors.New("cannot send")
	errCannotGenerateRandomString       = errors.New("cannot generate random string")
	errContextCanceled                  = errors.New("context canceled")
	errCannotReceiveMessage             = errors.New("cannot receive message")
	errNoSuchCourse                     = errors.New("reference to non-existent course")
	errInvalidState                     = errors.New("invalid state")
	errCannotSetState                   = errors.New("cannot set state")
	errWebSocketWrite                   = errors.New("error writing to websocket")
	errHTTPWrite                        = errors.New("error writing to http writer")
	errCannotCheckCookie                = errors.New("error checking cookie")
	errNoCookie                         = errors.New("no cookie found")
	errNoSuchUser                       = errors.New("no such user")
	errNoSuchYearGroup                  = errors.New("no such year group")
	errPostOnly                         = errors.New("only post is supported on this endpoint")
	errMalformedForm                    = errors.New("malformed form")
	errAuthorizeEndpointError           = errors.New("authorize endpoint returned error")
	errCannotParseClaims                = errors.New("cannot parse claims")
	errCannotUnpackClaims               = errors.New("cannot unpack claims")
	errJWTMalformed                     = errors.New("jwt token is malformed")
	errJWTSignatureInvalid              = errors.New("jwt token has invalid signature")
	errJWTExpired                       = errors.New("jwt token has expired or is not yet valid")
	errJWTInvalid                       = errors.New("jwt token is somehow invalid")
	errStaffOnly                        = errors.New("this page is only available to staff")
	errDisableStudentAccessFirst        = errors.New("you must disable student access before performing this operation")
	errFormNoFile                       = errors.New("you need to select a file before submitting the form")
	errNotACSV                          = errors.New("the file you uploaded is not a csv file")
	errCannotReadCSV                    = errors.New("cannot read csv")
	errBadCSVFormat                     = errors.New("bad csv format")
	errMissingCSVColumn                 = errors.New("missing csv column")
	errUnexpectedNilCSVLine             = errors.New("unexpected nil csv line")
	errWhileSetttingUpCourseTablesAgain = errors.New("error while setting up course tables again")
	errCannotWriteTemplate              = errors.New("cannot write template")
	errUnknownCommand                   = errors.New("unknown command")
	errBadNumberOfArguments             = errors.New("bad number of arguments")
	errInvalidYearGroupOrCourseType     = errors.New("invalid year group or course type (something is broken)")
	// errInvalidCourseID                  = errors.New("invalid course id")
)

func wrapError(a, b error) error {
	if a == nil && b == nil {
		return nil
	}
	return fmt.Errorf("%w: %w", a, b)
}

func wrapAny(a error, b any) error {
	if a == nil && b == nil {
		return nil
	}
	return fmt.Errorf("%w: %v", a, b)
}
