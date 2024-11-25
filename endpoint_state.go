/*
 * Let staff update state
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"net/http"
	"strconv"
)

func handleState(w http.ResponseWriter, req *http.Request) (string, int, error) {
	_, _, department, err := getUserInfoFromRequest(req)
	if err != nil {
		return "", http.StatusUnauthorized, err
	}
	if department != staffDepartment {
		return "", http.StatusForbidden, errStaffOnly
	}

	basePath := req.PathValue("s")
	newState, err := strconv.ParseUint(basePath, 10, 32)
	if err != nil {
		return "", http.StatusBadRequest, wrapError(errInvalidState, err)
	}
	err = setState(req.Context(), uint32(newState))
	if err != nil {
		return "", http.StatusBadRequest, wrapError(errCannotSetState, err)
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)
	return "", -1, nil
}
