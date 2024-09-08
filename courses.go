package main

import (
	"context"
	"sync"
)

type coursetype_t string

type course_t struct {
	Id        int
	Confirmed int
	Selected  int
	Max       int
	Title     string
	Type      coursetype_t
	Teacher   string
	Location  string
}

const (
	sport      coursetype_t = "Sport"
	enrichment coursetype_t = "Enrichment"
	culture    coursetype_t = "Culture"
)

var courses []course_t

/*
 * TODO: revamp this.
 * This RWMutex is only for massive modifications of the course struct, since
 * locking it on every write would be inefficient; in normal operation the only
 * write that could occur to the courses struct is changing the Confirmed and
 * Selected numbers, which should be handled with either atomics compare and
 * swap, or a small RWMutex exclusive to that course. More idiomatic Go style
 * would suggest putting course handling code in a separate goroutine but that
 * seems needlessly inefficient in a highly concurrent application.
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

	courses = make([]course_t, 0, 64)

	rows, err := db.Query(
		context.Background(),
		"SELECT id, nmax, title, ctype, teacher, location FROM courses",
	)
	if err != nil {
		return err
	}

	for {
		if !rows.Next() {
			err := rows.Err()
			if err != nil {
				return err
			}
			break
		}
		currentCourse := course_t{}
		err = rows.Scan(
			&currentCourse.Id,
			&currentCourse.Max,
			&currentCourse.Title,
			&currentCourse.Type,
			&currentCourse.Teacher,
			&currentCourse.Location,
		)
		if err != nil {
			return err
		}
		courses = append(courses, currentCourse)
	}

	return nil
}
