/*
 * Index page
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"errors"
	"net/http"
	"sync/atomic"
)

func handleIndex(w http.ResponseWriter, req *http.Request) (string, int, error) {
	_, username, department, err := getUserInfoFromRequest(req)
	if errors.Is(err, errNoCookie) || errors.Is(err, errNoSuchUser) {
		authURL, err2 := generateAuthorizationURL()
		if err2 != nil {
			return "", -1, err2
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
			return "", -1, wrapError(errCannotWriteTemplate, err2)
		}
		return "", -1, nil
	} else if err != nil {
		return "", -1, err
	}

	/* TODO: The below should be completed on-update. */
	type groupT struct {
		Handle  string
		Name    string
		Courses *map[int]*courseT
	}
	_groups := make(map[string]groupT)
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
		StatesDereferenced := map[string]uint32{}
		for k, v := range states {
			StatesDereferenced[k] = atomic.LoadUint32(v)
		}
		err := tmpl.ExecuteTemplate(
			w,
			"staff",
			struct {
				Name     string
				States   map[string]uint32
				StatesOr uint32
				Groups   *map[string]groupT
			}{
				username,
				StatesDereferenced,
				func() uint32 {
					var ret uint32 /* all zero bits */
					for _, v := range StatesDereferenced {
						ret |= v
					}
					return ret
				}(),
				&_groups,
			},
		)
		if err != nil {
			return "", -1, wrapError(errCannotWriteTemplate, err)
		}
		return "", -1, nil
	}

	_state, ok := states[department]
	if !ok {
		return "", -1, errNoSuchYearGroup
	}
	if atomic.LoadUint32(_state) == 0 {
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
			return "", -1, wrapError(errCannotWriteTemplate, err)
		}
		return "", -1, nil
	}
	sportRequired, err := getCourseTypeMinimumForYearGroup(
		department, sport,
	)
	if err != nil {
		return "", -1, err
	}
	nonSportRequired, err := getCourseTypeMinimumForYearGroup(
		department, nonSport,
	)
	if err != nil {
		return "", -1, err
	}

	err = tmpl.ExecuteTemplate(
		w,
		"student",
		struct {
			Name       string
			Department string
			Groups     *map[string]groupT
			Required   struct {
				Sport    int
				NonSport int
			}
		}{
			username,
			department,
			&_groups,
			struct {
				Sport    int
				NonSport int
			}{sportRequired, nonSportRequired},
		},
	)
	if err != nil {
		return "", -1, wrapError(errCannotWriteTemplate, err)
	}
	return "", -1, nil
}
