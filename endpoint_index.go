/*
 * Index page
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
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
)

func handleIndex(w http.ResponseWriter, req *http.Request) {
	_, username, department, err := getUserInfoFromRequest(req)
	if errors.Is(err, errNoCookie) || errors.Is(err, errNoSuchUser) {
		authURL, err2 := generateAuthorizationURL()
		if err2 != nil {
			wstr(
				w,
				http.StatusInternalServerError,
				"Cannot generate authorization URL",
			)
			return
		}
		var noteString string
		if errors.Is(err, errNoSuchUser) {
			noteString = "Your browser provided an invalid session cookie."
		}
		err2 = tmpl.ExecuteTemplate(
			w,
			"login",
			struct {
				AuthURL string
				Notes   string
			}{
				authURL,
				noteString,
			},
		)
		if err2 != nil {
			log.Println(err2)
		}
		return
	} else if err != nil {
		wstr(w, http.StatusInternalServerError, fmt.Sprintf("Error: %v", err))
	}

	/* TODO: The below should be completed on-update. */
	type groupT struct {
		Handle  courseGroupT
		Name    string
		Courses *map[int]*courseT
	}
	_groups := make(map[courseGroupT]groupT)
	for k, v := range courseGroups {
		_coursemap := make(map[int]*courseT)
		_groups[k] = groupT{
			Handle:  k,
			Name:    v,
			Courses: &_coursemap,
		}
	}
	courses.Range(func(key, value interface{}) bool {
		courseID, ok := key.(int)
		if !ok {
			panic("courses map has non-\"int\" keys")
		}
		course, ok := value.(*courseT)
		if !ok {
			panic("courses map has non-\"*courseT\" items")
		}
		(*_groups[course.Group].Courses)[courseID] = course
		return true
	})

	if department == staffDepartment {
		err := tmpl.ExecuteTemplate(
			w,
			"staff",
			struct {
				Name   string
				State  uint32
				Groups *map[courseGroupT]groupT
			}{
				username,
				state,
				&_groups,
			},
		)
		if err != nil {
			log.Println(err)
		}
		return
	}

	if atomic.LoadUint32(&state) == 0 {
		err := tmpl.ExecuteTemplate(
			w,
			"student_disabled",
			struct {
				Name       string
				Department string
			}{
				username,
				department,
			},
		)
		if err != nil {
			log.Println(err)
		}
		return
	}

	err = tmpl.ExecuteTemplate(
		w,
		"student",
		struct {
			Name       string
			Department string
			Groups     *map[courseGroupT]groupT
		}{
			username,
			department,
			&_groups,
		},
	)
	if err != nil {
		log.Println(err)
		return
	}
}
