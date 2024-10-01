/*
 * Index page
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
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
)

/*
 * Serve the index page. Also handles the login page in case the user doesn't
 * have any valid login cookies.
 */
func handleIndex(w http.ResponseWriter, req *http.Request) {
	sessionCookie, err := req.Cookie("session")
	if errors.Is(err, http.ErrNoCookie) {
		authURL, err := generateAuthorizationURL()
		if err != nil {
			wstr(w, http.StatusInternalServerError, "Cannot generate authorization URL")
			return
		}
		err = tmpl.ExecuteTemplate(
			w,
			"index_login",
			map[string]string{
				"authURL": authURL,
				"source":  config.Source,
				/*
				 * We directly generate the login URL here
				 * instead of doing so in a redirect to save
				 * requests.
				 */
			},
		)
		if err != nil {
			log.Println(err)
			return
		}
		return
	} else if err != nil {
		wstr(w, http.StatusBadRequest, "Error: Unable to check cookie.")
		return
	}

	var userID, userName, userDepartment string
	err = db.QueryRow(
		req.Context(),
		"SELECT id, name, department FROM users WHERE session = $1",
		sessionCookie.Value,
	).Scan(&userID, &userName, &userDepartment)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			authURL, err := generateAuthorizationURL()
			if err != nil {
				wstr(w, http.StatusInternalServerError, "Cannot generate authorization URL")
				return
			}
			err = tmpl.ExecuteTemplate(
				w,
				"index_login",
				map[string]interface{}{
					"authURL": authURL,
					"notes":   "You sent an invalid session cookie.",
					"source":  config.Source,
				},
			)
			if err != nil {
				log.Println(err)
				return
			}
			return
		}
		wstr(w, http.StatusInternalServerError, fmt.Sprintf("Error: Unexpected database error: %s", err))
		return
	}

	err = func() error {
		coursesLock.RLock()
		defer coursesLock.RUnlock()
		return tmpl.ExecuteTemplate(
			w,
			"index",
			map[string]interface{}{
				"open": true,
				"user": map[string]interface{}{
					"Name":       userName,
					"Department": userDepartment,
				},
				"courses": courses,
				"source":  config.Source,
			},
		)
	}()
	if err != nil {
		log.Println(err)
		return
	}
}
