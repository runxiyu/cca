/*
 * WebSocket endpoint handler
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
	"log"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

/*
 * Handle requests to the WebSocket endpoint and establish a connection.
 * Authentication is handled here, but afterwards, the connection is really
 * handled in handleConn.
 */
func handleWs(w http.ResponseWriter, req *http.Request) {
	wsOptions := &websocket.AcceptOptions{
		Subprotocols: []string{"cca1"},
	} //exhaustruct:ignore
	c, err := websocket.Accept(
		w,
		req,
		wsOptions,
	)
	if err != nil {
		wstr(w, http.StatusBadRequest, "This endpoint only supports valid WebSocket connections.")
		return
	}
	defer func() {
		_ = c.CloseNow()
	}()

	fake := false

	sessionCookie, err := req.Cookie("session")
	if errors.Is(err, http.ErrNoCookie) {
		if config.Auth.Fake == 0 {
			err := writeText(req.Context(), c, "U")
			if err != nil {
				log.Println(err)
			}
			return
		}
		fake = true
	} else if err != nil {
		err := writeText(req.Context(), c, "E :Error fetching cookie")
		if err != nil {
			log.Println(err)
		}
		return
	}

	var userID string
	var session string
	var expr int

	if fake {
		switch config.Auth.Fake {
		case 9080:
			_uuid, err := uuid.NewRandom()
			if err != nil {
				log.Println(err)
				return
			}
			userID = _uuid.String()
		case 4712:
			userID = "fake"
		default:
			panic("not supposed to happen")
		}
		session, err = randomString(20)
		if err != nil {
			log.Println(err)
			return
		}
		_, err = db.Exec(
			req.Context(),
			"INSERT INTO users (id, name, email, department, session, expr) VALUES ($1, $2, $3, $4, $5, $6)",
			userID,
			"Fake User",
			"fake@runxiyu.org",
			"Y11",
			session,
			time.Now().Add(time.Duration(config.Auth.Expr)*time.Second).Unix(),
		)
		if err != nil && config.Auth.Fake != 4712 {
			/* TODO check pgerr */
			err := writeText(req.Context(), c, "E :Database error while writing fake account info")
			if err != nil {
				log.Println(err)
			}
			return
		}
		err = writeText(req.Context(), c, "FAKE "+userID+" "+session)
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		session = sessionCookie.Value
		err = db.QueryRow(
			req.Context(),
			"SELECT id, expr FROM users WHERE session = $1",
			session,
		).Scan(&userID, &expr)
		if errors.Is(err, pgx.ErrNoRows) {
			err := writeText(req.Context(), c, "U")
			if err != nil {
				log.Println(err)
			}
			return
		} else if err != nil {
			err := writeText(req.Context(), c, "E :Database error while selecting session")
			if err != nil {
				log.Println(err)
			}
			return
		}
	}

	/*
	 * Now that we have an authenticated request, this WebSocket connection
	 * may be simply associated with the session and userID.
	 * TODO: There are various race conditions that could occur if one user
	 * creates multiple connections, with the same or different session
	 * cookies. The last situation could occur in normal use when a user
	 * opens multiple instances of the page in one browser, and is not
	 * unique to custom clients or malicious users. Some effort must be
	 * taken to ensure that each user may only have one connection at a
	 * time.
	 */
	err = handleConn(
		req.Context(),
		c,
		session,
		userID,
	)
	if err != nil {
		log.Printf("%v", err)
		return
	}
}
