/*
 * Handle the unified global state
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5"
)

/*
 * The uint32 should be accessed atomically
 * 0: Student access is disabled
 * 1: Student have read-only access
 * 2: Student can choose courses
 */
var states = map[string]*uint32{
	"Y9":  new(uint32),
	"Y10": new(uint32),
	"Y11": new(uint32),
	"Y12": new(uint32),
}

var schedules = map[string]*atomic.Pointer[time.Time]{
	"Y9":  {},
	"Y10": {},
	"Y11": {},
	"Y12": {},
}

func loadStateAndSchedule() error {
	for yeargroup := range states {
		var state uint32
		var schedule time.Time
		err := db.QueryRow(
			context.Background(),
			"SELECT state, schedule FROM states WHERE yeargroup = $1",
			yeargroup,
		).Scan(&state, &schedule)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				state = 0
				_, err := db.Exec(
					context.Background(),
					"INSERT INTO states(yeargroup, state, schedule) VALUES ($1, $2, $3)",
					yeargroup,
					state,
					time.Time{},
				)
				if err != nil {
					return wrapError(errUnexpectedDBError, err)
				}
			} else {
				return wrapError(errUnexpectedDBError, err)
			}
		}
		_state, ok := states[yeargroup]
		if !ok {
			return errNoSuchYearGroup
		}
		_schedule, ok := schedules[yeargroup]
		if !ok {
			return errNoSuchYearGroup
		}
		atomic.StoreUint32(_state, state)
		_schedule.Store(&schedule)
	}
	return nil
}

func saveStateValue(ctx context.Context, yeargroup string, newState uint32) error {
	_, err := db.Exec(
		ctx,
		"UPDATE states SET state = $2 WHERE yeargroup = $1",
		yeargroup,
		newState,
	)
	if err != nil {
		return wrapError(errUnexpectedDBError, err)
	}
	return nil
}

func saveScheduleValue(ctx context.Context, yeargroup string, newSchedule *time.Time) error {
	_, err := db.Exec(
		ctx,
		"UPDATE states SET schedule = $2 WHERE yeargroup = $1",
		yeargroup,
		*newSchedule,
	)
	if err != nil {
		return wrapError(errUnexpectedDBError, err)
	}
	return nil
}

func setSchedule(ctx context.Context, yeargroup string, newSchedule *time.Time) error {
	_schedule, ok := schedules[yeargroup]
	if !ok {
		return errNoSuchYearGroup
	}
	_schedule.Store(newSchedule)
	return saveScheduleValue(ctx, yeargroup, newSchedule)
}

func setState(ctx context.Context, yeargroup string, newState uint32) error {
	switch newState {
	case 0:
	case 1:
		err := propagate(yeargroup, "STOP") /* TODO: propagate by year group */
		if err != nil {
			return err
		}
	case 2:
		err := propagate(yeargroup, "START")
		if err != nil {
			return err
	 	}
	case 3:
		// TODO XXX: Implement this!
	default:
		return errInvalidState
	}
	err := saveStateValue(ctx, yeargroup, newState)
	if err != nil {
		return err
	}
	_state, ok := states[yeargroup]
	if !ok {
		return errNoSuchYearGroup
	}
	atomic.StoreUint32(_state, newState)
	return nil
}
