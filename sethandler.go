/*
 * HTTP handler setting
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"log/slog"
	"net/http"
)

func setHandler(pattern string, handler func(
	http.ResponseWriter,
	*http.Request,
) (string, int, error)) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		defer func() {
			if e := recover(); e != nil {
				slog.Error("panic", "arg", e)
			}
		}()

		msg, statusCode, err := handler(w, req)
		if err != nil {
			if statusCode == -1 || statusCode == 0 {
				statusCode = 500
			}
			slog.Error(
				"handler",
				"path", req.URL.Path,
				"status", statusCode,
				"error", err,
			)
			if msg != "" {
				wstr(w, statusCode, msg+"\n"+err.Error())
			} else {
				wstr(w, statusCode, err.Error())
			}
		} else if msg != "" {
			if statusCode == -1 || statusCode == 0 {
				statusCode = 200
			}
			wstr(w, statusCode, msg)
		}
	})
}
