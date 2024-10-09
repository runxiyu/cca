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
			wstr(
				w,
				http.StatusInternalServerError,
				"Cannot generate authorization URL",
			)
			return
		}
		err = tmpl.ExecuteTemplate(
			w,
			"login",
			struct {
				AuthURL string
				Notes   string
			}{
				authURL,
				"",
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
				wstr(
					w,
					http.StatusInternalServerError,
					"Cannot generate authorization URL",
				)
				return
			}
			err = tmpl.ExecuteTemplate(
				w,
				"login",
				struct {
					AuthURL string
					Notes   string
				}{
					authURL,
					"Your session is invalid or has expired.",
				},
			)
			if err != nil {
				log.Println(err)
				return
			}
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

	/*
	 * Copy courses to _courses. The former is a sync.Map and the latter is
	 * a map[int]*courseT, and the former is very difficult to access from
	 * HTML templates.
	 */
	_courses := make(map[int]*courseT)
	courses.Range(func(key, value interface{}) bool {
		courseID, ok := key.(int)
		if !ok {
			panic("courses map has non-\"int\" keys")
		}
		course, ok := value.(*courseT)
		if !ok {
			panic("courses map has non-\"*courseT\" items")
		}
		_courses[courseID] = course
		return true
	})

	err = func() error {
		/* Horrifying syntax */
		return tmpl.ExecuteTemplate(
			w,
			"student",
			struct {
				Open       bool
				Name       string
				Department string
				Courses    *map[int]*courseT
			}{
				true,
				userName,
				userDepartment,
				&_courses,
			},
		)
	}()
	if err != nil {
		log.Println(err)
		return
	}
}
