/*
 * Handle the "HELLO" message
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

	if atomic.LoadUint32(&state) == 2 {
		err = writeText(ctx, c, "START")
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
