/*
 * Let staff update state
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"errors"
	"net/http"
	"strconv"
)

var errMethodNotAllowed = errors.New("method not allowed")
var errInvalidForm = errors.New("invalid form")

func handleState(w http.ResponseWriter, req *http.Request) (string, int, error) {
	if req.Method != http.MethodPost {
		return "", http.StatusMethodNotAllowed, errMethodNotAllowed
	}

	_, _, department, err := getUserInfoFromRequest(req)
	if err != nil {
		return "", http.StatusUnauthorized, err
	}
	if department != staffDepartment {
		return "", http.StatusForbidden, errStaffOnly
	}

	err = req.ParseForm()
	if err != nil {
		return "", http.StatusBadRequest, wrapError(errInvalidForm, err)
	}

	for yeargroup, _ := range states {
		key := "yeargroup_" + yeargroup
		if newStateStr := req.FormValue(key); newStateStr != "" {
			newState, err := strconv.ParseUint(newStateStr, 10, 32)
			if err != nil {
				return "", http.StatusBadRequest, wrapError(errInvalidState, err)
			}
			err = setState(req.Context(), yeargroup, uint32(newState))
			if err != nil {
				return "", http.StatusBadRequest, wrapError(errCannotSetState, err)
			}
		}
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)
	return "", -1, nil
}
