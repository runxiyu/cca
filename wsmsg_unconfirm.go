/*
 * Handle the "C" message
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
	"sync/atomic"

	"github.com/coder/websocket"
)

func messageUnconfirm(
	ctx context.Context,
	c *websocket.Conn,
	reportError reportErrorT,
	mar []string,
	userID string,
) error {
	_ = mar

	if atomic.LoadUint32(&state) != 2 {
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
			errContextCanceled,
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
		return reportError("error updating database setting confirmation")
	}

	return writeText(
		ctx,
		c,
		"NC",
	)
}
