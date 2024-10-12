/*
 * HTTP handler setting
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
	"net/http"
)

func setHandler(pattern string, handler func(http.ResponseWriter, *http.Request) (string, int, error)) {
	http.HandleFunc(pattern, func(w http.ResponseWriter, req *http.Request) {
		msg, statusCode, err := handler(w, req)
		if err != nil {
			if statusCode == -1 || statusCode == 0 {
				statusCode = 500
			}
			if msg != "" {
				wstr(w, statusCode, msg+"\n"+err.Error())
			} else {
				wstr(w, statusCode, err.Error())
			}
		} else {
			if statusCode == -1 || statusCode == 0 {
				statusCode = 200
			}
			wstr(w, statusCode, msg)
		}
	})
}
