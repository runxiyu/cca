/*
 * Let staff update state
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
	"fmt"
	"net/http"
	"strconv"
)

func handleState(w http.ResponseWriter, req *http.Request) {
	_, _, department, err := getUserInfoFromRequest(req)
	if err != nil {
		wstr(
			w,
			http.StatusInternalServerError,
			fmt.Sprintf("Error: %v", err),
		)
	}
	if department != staffDepartment {
		wstr(
			w,
			http.StatusForbidden,
			"You are not authorized to view this page",
		)
		return
	}

	basePath := req.PathValue("s")
	newState, err := strconv.ParseUint(basePath, 10, 32)
	if err != nil {
		wstr(
			w,
			http.StatusBadRequest,
			"State must be an unsigned 32-bit integer",
		)
		return
	}
	err = setState(req.Context(), uint32(newState))
	if err != nil {
		wstr(
			w,
			http.StatusInternalServerError,
			"Failed setting state, please return to previous page; are you sure it's within limits?",
		)
		return
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)
}
