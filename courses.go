/*
 * Course data structures and locking
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: BSD-2-Clause
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are
 * met:
 *
 *     1. Redistributions of source code must retain the above copyright
 *     notice, this list of conditions and the following disclaimer.
 *
 *     2. Redistributions in binary form must reproduce the above copyright
 *     notice, this list of conditions and the following disclaimer in the
 *     documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS "AS IS" AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR
 * PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR
 * CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
 * EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
 * PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
 * PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
 * LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
 * NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package main

import (
	"context"
	"fmt"
	"sync"
)

type coursetypeT string

type courseT struct {
	ID        int
	Confirmed int
	Selected  int
	Max       int
	Title     string
	Type      coursetypeT
	Teacher   string
	Location  string
	Lock      sync.RWMutex
}

/*
 * const (
 * 	sport      coursetypeT = "Sport"
 * 	enrichment coursetypeT = "Enrichment"
 * 	culture    coursetypeT = "Culture"
 * )
 */

/*
 * The courses are simply stored in a map indexed by the course ID, although
 * the course struct itself also contains an ID field. A lock is embedded
 * inside the struct; we use a lock here instead of a pointer to a lock as
 * it would be easy to forget to initialize the lock when creating the
 * struct. However, this means that the struct could not be copied (though
 * this should only ever happen during creation anyways), therefore we use a
 * pointer to the struct as the value of the map, instead of the struct itself.
 */
var courses map[int](*courseT)

/*
 * This RWMutex is only for massive modifications of the course struct, since
 * locking it on every write would be inefficient; in normal operation the only
 * write that could occur to the courses struct is changing the Confirmed and
 * Selected numbers, which should be handled with the small RWMutex within the
 * course struct.
 */
var coursesLock sync.RWMutex

/*
 * Read course information from the database. This should be called during
 * setup. Failure to do so before accessing course information may lead to
 * a null pointer dereference.
 */
func setupCourses() error {
	coursesLock.Lock()
	defer coursesLock.Unlock()

	courses = make(map[int](*courseT))

	rows, err := db.Query(
		context.Background(),
		"SELECT id, nmax, title, ctype, teacher, location FROM courses",
	)
	if err != nil {
		return fmt.Errorf("error fetching courses: %w", err)
	}

	for {
		if !rows.Next() {
			err := rows.Err()
			if err != nil {
				return fmt.Errorf("error fetching courses: %w", err)
			}
			break
		}
		currentCourse := courseT{} //exhaustruct:ignore
		err = rows.Scan(
			&currentCourse.ID,
			&currentCourse.Max,
			&currentCourse.Title,
			&currentCourse.Type,
			&currentCourse.Teacher,
			&currentCourse.Location,
		)
		if err != nil {
			return fmt.Errorf("error fetching courses: %w", err)
		}
		courses[currentCourse.ID] = &currentCourse
	}

	return nil
}
