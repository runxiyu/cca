/*
 * Let staff update state
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"
)

var (
	errMethodNotAllowed = errors.New("method not allowed")
	errInvalidForm      = errors.New("invalid form")
	errInvalidSchedule  = errors.New("invalid schedule")
)

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

	log.Println(req.Form)

	for k, v := range schedules {
		_v := v.Load()
		if _v != nil {
			log.Printf("before schedule %s: %s\n", k, _v.Format("2006-01-02T15:04"))
		} else {
			log.Printf("before schedule %s: nil\n", k)
		}
	}

	for yeargroup := range states {
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
		keySched := "schedule_" + yeargroup
		if newScheduleStr := req.FormValue(keySched); newScheduleStr != "" {
			newSchedule, err := time.Parse("2006-01-02T15:04", newScheduleStr)
			if err != nil {
				return "", http.StatusBadRequest, wrapError(errInvalidSchedule, err)
			}
			err = setSchedule(req.Context(), yeargroup, &newSchedule)
			if err != nil {
				return "", http.StatusBadRequest, wrapError(errCannotSetSchedule, err)
			}
		}
	}

	for k, v := range schedules {
		_v := v.Load()
		if _v != nil {
			log.Printf("after schedule %s: %s\n", k, _v.Format("2006-01-02T15:04"))
		} else {
			log.Printf("after schedule %s: nil\n", k)
		}
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)
	return "", -1, nil
}
