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

func loadState() error {
	for yeargroup := range states {
		var _state uint32
		err := db.QueryRow(
			context.Background(),
			"SELECT state FROM states WHERE yeargroup = $1",
			yeargroup,
		).Scan(&_state)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				_state = 0
				_, err := db.Exec(
					context.Background(),
					"INSERT INTO states(yeargroup, state) VALUES ($1, $2)",
					yeargroup,
					_state,
				)
				if err != nil {
					return wrapError(errUnexpectedDBError, err)
				}
			} else {
				return wrapError(errUnexpectedDBError, err)
			}
		}
		__state, ok := states[yeargroup]
		if !ok {
			return errNoSuchYearGroup
		}
		atomic.StoreUint32(__state, _state)
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

func setState(ctx context.Context, yeargroup string, newState uint32) error {
	switch newState {
	case 0:
		/*
		 * cancelPool.Range(func(_, value interface{}) bool {
		 * 	cancel, ok := value.(*context.CancelFunc)
		 * 	if !ok {
		 * 		panic("chanPool has non-\"*contect.CancelFunc\" values")
		 * 	}
		 * 	(*cancel)()
		 * 	return false
		 * })
		 * We previously used the above but now we just check for state
		 * before handling each message, so no changes have to be made
		 * to cancelPool.
		 */
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
