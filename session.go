/*
 * Session checking functions
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
	"net/http"

	"github.com/jackc/pgx/v5"
)

func getUserInfoFromRequest(req *http.Request) (userID, username, department string, retErr error) {
	sessionCookie, err := req.Cookie("session")
	if errors.Is(err, http.ErrNoCookie) {
		retErr = wrapError(errNoCookie, err)
		return
	} else if err != nil {
		retErr = wrapError(errCannotCheckCookie, err)
		return
	}

	err = db.QueryRow(
		req.Context(),
		"SELECT id, name, department FROM users WHERE session = $1",
		sessionCookie.Value,
	).Scan(&userID, &username, &department)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			retErr = errNoSuchUser
			return
		}
		retErr = wrapError(errUnexpectedDBError, err)
		return
	}
	return
}
