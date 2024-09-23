/*
 * Index page
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: BSD-2-Clause
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are
 * met:
 *
 *     1. Redistributions of source code must retain the above copyright
 *     notice, this list of conditions and the following disclaimer.
 *
 *     2. Redistributions in binary form must reproduce the above copyright
 *     notice, this list of conditions and the following disclaimer in the
 *     documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS "AS IS" AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR
 * PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR
 * CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
 * EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
 * PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
 * PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
 * LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
 * NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package main

import (
	"context"
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
	session_cookie, err := req.Cookie("session")
	if errors.Is(err, http.ErrNoCookie) {
		authUrl, err := generate_authorization_url()
		if err != nil {
			wstr(w, 500, "Cannot generate authorization URL")
			return
		}
		err = tmpl.ExecuteTemplate(
			w,
			"index_login",
			map[string]string{
				"authUrl": authUrl,
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
		wstr(w, 400, "Error: Unable to check cookie.")
		return
	}

	var userid string
	err = db.QueryRow(
		context.Background(),
		"SELECT userid FROM sessions WHERE cookie = $1",
		session_cookie.Value,
	).Scan(&userid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			authUrl, err := generate_authorization_url()
			if err != nil {
				wstr(w, 500, "Cannot generate authorization URL")
				return
			}
			err = tmpl.ExecuteTemplate(
				w,
				"index_login",
				map[string]interface{}{
					"authUrl": authUrl,
					"notes":   []string{"Technically you have a session cookie, but it seems invalid."},
				},
			)
			if err != nil {
				log.Println(err)
				return
			}
			return
		} else {
			wstr(w, 500, fmt.Sprintf("Error: Unexpected database error: %s", err))
			return
		}
	}

	var name string
	var department string
	err = db.QueryRow(
		context.Background(),
		"SELECT name, department FROM users WHERE id = $1",
		userid,
	).Scan(&name, &department)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			wstr(w, 500, "Error: User does not exist (database error?)")
			return
		} else {
			wstr(w, 500, "Error: Unexpected database error")
			return
		}
	}
	err = tmpl.ExecuteTemplate(
		w,
		"index",
		map[string]interface{}{
			"open": true,
			"user": map[string]interface{}{
				"Name":       name,
				"Department": department,
			},
			"courses": courses,
		},
	)
	if err != nil {
		log.Println(err)
		return
	}
}
