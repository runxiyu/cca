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

	yeargroupParams := req.URL.Query()["yeargroup"]
	if len(yeargroupParams) == 0 {
		return "", http.StatusBadRequest, errNoSuchYearGroup
	}
	targetParams := req.URL.Query()["target"]
	if len(targetParams) != 1 {
		return "", http.StatusBadRequest, errInvalidState
	}
	newState, err := strconv.ParseUint(targetParams[0], 10, 32)
	if err != nil {
		return "", http.StatusBadRequest, wrapError(errInvalidState, err)
	}

	for _, yeargroup := range yeargroupParams {
		err = setState(req.Context(), yeargroup, uint32(newState))
		if err != nil {
			return "", http.StatusBadRequest, wrapError(
				errCannotSetState,
				err,
			)
		}
	}
	http.Redirect(w, req, "/", http.StatusSeeOther)
	return "", -1, nil
}
