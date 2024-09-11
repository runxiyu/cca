/*
 * Primary WebSocket routines
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

	"github.com/coder/websocket"
	"github.com/jackc/pgx/v5"
)

/*
 * Handle requests to the WebSocket endpoint and establish a connection.
 * The connection is really handled in handleConn.
 */
func handleWs(w http.ResponseWriter, req *http.Request) {
	c, err := websocket.Accept(w, req, &websocket.AcceptOptions{
		Subprotocols: []string{"cca1"},
	})
	if err != nil {
		w.Write([]byte("This endpoint only supports valid WebSocket connections."))
		return
	}
	defer c.CloseNow()

	session_cookie, err := req.Cookie("session")
	if errors.Is(err, http.ErrNoCookie) {
		c.Write(
			req.Context(),
			websocket.MessageText,
			[]byte("U"),
		)
		return
	} else if err != nil {
		c.Write(
			req.Context(),
			websocket.MessageText,
			[]byte("E :Error fetching cookie"),
		)
		return
	}

	err = handleConn(req.Context(), c, session_cookie.Value)
	if err != nil {
		log.Printf("%v", err)
		return
	}
}

/*
 * Split an IRC-style message of type []byte into type []string where each
 * element is a complete argument. Generally, arguments are separated by
 * spaces, and an argument that begins with a ':' causes the rest of the
 * line to be treated as a single argument.
 */
func splitMsg(b []byte) []string {
	mar := make([]string, 0, 4)
	elem := make([]byte, 0, 5)
	for i, c := range b {
		switch c {
		case ' ':
			if b[i+1] == ':' {
				mar = append(mar, string(elem))
				mar = append(mar, string(b[i+2:]))
				goto endl
			}
			mar = append(mar, string(elem))
			elem = make([]byte, 0, 5)
		default:
			elem = append(elem, c)
		}
	}
	mar = append(mar, string(elem))
endl:
	return mar
}

/*
 * The actual logic in handling the connection.
 */
func handleConn(ctx context.Context, c *websocket.Conn, session string) error {
	var userid string
	var expr int
	err := db.QueryRow(
		context.Background(),
		"SELECT userid, expr FROM sessions WHERE cookie = $1",
		session,
	).Scan(&userid, &expr)
	if errors.Is(err, pgx.ErrNoRows) {
		c.Write(ctx, websocket.MessageText, []byte("U"))
		return err
	} else if err != nil {
		c.Write(ctx, websocket.MessageText, []byte("E :Database error"))
		return err
	}

	/* TODO: Select from this and a broadcast channel */
	for {
		_, b, err := c.Read(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("%s %s\n", session, string(b))

		mar := splitMsg(b)

		switch mar[0] {
		case "HELLO":
		}
	}

	// err = c.Write(ctx, typ, b)
	// if err != nil {
	// 	return err
	// }

	return nil
}
