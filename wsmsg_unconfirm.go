/*
 * Handle the "C" message
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"context"
	"sync/atomic"

	"github.com/coder/websocket"
)

func messageUnconfirm(
	ctx context.Context,
	c *websocket.Conn,
	mar []string,
	userID string,
	yeargroup string,
) error {
	_ = mar

	_state, ok := states[yeargroup]
	if !ok {
		return errNoSuchYearGroup
	}
	if atomic.LoadUint32(_state) != 2 {
		err := writeText(ctx, c, "E :Course selections are not open")
		if err != nil {
			return wrapError(
				errCannotSend,
				err,
			)
		}
		return nil
	}

	select {
	case <-ctx.Done():
		return wrapError(
			errWsHandlerContextCanceled,
			ctx.Err(),
		)
	default:
	}

	_, err := db.Exec(
		ctx,
		"UPDATE users SET confirmed = false WHERE id = $1",
		userID,
	)
	if err != nil {
		return wrapError(errUnexpectedDBError, err)
	}

	return writeText(
		ctx,
		c,
		"NC",
	)
}
