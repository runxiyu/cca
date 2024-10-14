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
	"log/slog"
	"net/http"

	"github.com/coder/websocket"
)

/*
 * Handle requests to the WebSocket endpoint and establish a connection.
 * Authentication is handled here, but afterwards, the connection is really
 * handled in handleConn.
 */
func handleWs(w http.ResponseWriter, req *http.Request) {
	defer func() {
		if e := recover(); e != nil {
			slog.Error("panic", "arg", e)
		}
	}()

	wsOptions := &websocket.AcceptOptions{
		Subprotocols: []string{"cca1"},
	} //exhaustruct:ignore
	c, err := websocket.Accept(
		w,
		req,
		wsOptions,
	)
	if err != nil {
		wstr(
			w,
			http.StatusBadRequest,
			"this endpoint only supports valid WebSocket connections: "+err.Error(),
		)
		return
	}
	defer func() {
		_ = c.CloseNow()
	}()

	userID, _, department, err := getUserInfoFromRequest(req)
	if err != nil {
		_ = writeText(req.Context(), c, "U")
		return
	}

	err = handleConn(req.Context(), c, userID, department)
	if err != nil {
		slog.Error(
			"websocket",
			"user", userID,
			"error", err,
		)
		_ = writeText(req.Context(), c, "E :"+err.Error())
		return
	}
}
