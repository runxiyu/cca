/*
 * Handle the "HELLO" message
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"context"
	"strings"
	"sync/atomic"

	"github.com/coder/websocket"
	"github.com/jackc/pgx/v5"
)

func messageHello(
	ctx context.Context,
	c *websocket.Conn,
	mar []string,
	userID string,
	yeargroup string,
) error {
	_ = mar

	select {
	case <-ctx.Done():
		return wrapError(
			errWsHandlerContextCanceled,
			ctx.Err(),
		)
	default:
	}

	rows, err := db.Query(
		ctx,
		"SELECT courseid FROM choices WHERE userid = $1",
		userID,
	)
	if err != nil {
		return wrapError(errUnexpectedDBError, err)
	}
	courseIDs, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return wrapError(errUnexpectedDBError, err)
	}

	_state, ok := states[yeargroup]
	if !ok {
		return errNoSuchYearGroup
	}
	if atomic.LoadUint32(_state) == 2 {
		err = writeText(ctx, c, "START")
		if err != nil {
			return wrapError(errCannotSend, err)
		}
	} else {
		err = writeText(ctx, c, "STOP")
		if err != nil {
			return wrapError(errCannotSend, err)
		}
	}

	confirmed, err := getConfirmedStatus(ctx, userID)
	if err != nil {
		return err
	}
	if !confirmed {
		err = writeText(ctx, c, "NC")
		if err != nil {
			return wrapError(errCannotSend, err)
		}
	} else {
		err = writeText(ctx, c, "YC")
		if err != nil {
			return wrapError(errCannotSend, err)
		}
	}

	err = writeText(ctx, c, "HI :"+strings.Join(courseIDs, ","))
	if err != nil {
		return wrapError(errCannotSend, err)
	}

	return nil
}
