/*
 * Session checking functions
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
)

func getUserInfoFromRequest(req *http.Request) (userID,
	username string,
	department string,
	retErr error,
) {
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
