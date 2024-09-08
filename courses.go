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
	sport    coursetype_t = "Sport"
	nonsport coursetype_t = "Non-sport"
)

var courses []course_t

var coursesLock sync.RWMutex

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
