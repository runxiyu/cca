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
)

var (
	errCannotSetupJwks            = errors.New("cannot set up jwks")
	errInsufficientFields         = errors.New("insufficient fields")
	errCannotGetDepartment        = errors.New("cannot get department")
	errCannotFetchAccessToken     = errors.New("cannot fetch access token")
	errTokenEndpointReturnedError = errors.New("token endpoint returned error")
	errCannotProcessConfig        = errors.New("cannot process configuration file")
	errCannotOpenConfig           = errors.New("cannot open configuration file")
	errCannotDecodeConfig         = errors.New("cannot decode configuration file")
	errMissingConfigValue         = errors.New("missing configuration value")
	errIllegalConfig              = errors.New("illegal configuration")
	errInvalidCourseType          = errors.New("invalid course type")
	errInvalidCourseGroup         = errors.New("invalid course group")
	errMultipleChoicesInOneGroup  = errors.New("multiple choices per group per user")
	errUnsupportedDatabaseType    = errors.New("unsupported db type")
	errUnexpectedDBError          = errors.New("unexpected database error")
	errCannotSend                 = errors.New("cannot send")
	errCannotGenerateRandomString = errors.New("cannot generate random string")
	errContextCancelled           = errors.New("context cancelled")
	errCannotReceiveMessage       = errors.New("cannot receive message")
	errNoSuchCourse               = errors.New("no such course")
	errInvalidState               = errors.New("invalid state")
)
