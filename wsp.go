/*
 * WebSocket-based protocol auxiliary functions
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
	"fmt"
	"sync/atomic"

	"github.com/coder/websocket"
)

/*
 * Split an IRC-style message of type []byte into type []string where each
 * element is a complete argument. Generally, arguments are separated by
 * spaces, and an argument that begins with a ':' causes the rest of the
 * line to be treated as a single argument.
 */
func splitMsg(b *[]byte) []string {
	mar := make([]string, 0, config.Perf.MessageArgumentsCap)
	elem := make([]byte, 0, config.Perf.MessageBytesCap)
	for i, c := range *b {
		switch c {
		case ' ':
			if (*b)[i+1] == ':' {
				mar = append(mar, string(elem))
				mar = append(mar, string((*b)[i+2:]))
				goto endl
			}
			mar = append(mar, string(elem))
			elem = make([]byte, 0, config.Perf.MessageBytesCap)
		default:
			elem = append(elem, c)
		}
	}
	mar = append(mar, string(elem))
endl:
	return mar
}

func baseReportError(
	ctx context.Context,
	conn *websocket.Conn,
	e string,
) error {
	err := writeText(ctx, conn, "E :"+e)
	if err != nil {
		return fmt.Errorf("error reporting protocol violation: %w", err)
	}
	err = conn.Close(websocket.StatusProtocolError, e)
	if err != nil {
		return fmt.Errorf("error closing websocket: %w", err)
	}
	return nil
}

type reportErrorT func(e string) error

func makeReportError(ctx context.Context, conn *websocket.Conn) reportErrorT {
	return func(e string) error {
		return baseReportError(ctx, conn, e)
	}
}

func propagateSelectedUpdate(courseID int) {
	course := courses[courseID]
	course.UsemsLock.RLock()
	defer course.UsemsLock.RUnlock()
	for _, usem := range course.Usems {
		usem.set()
	}
}

func sendSelectedUpdate(
	ctx context.Context,
	conn *websocket.Conn,
	courseID int,
) error {
	course := courses[courseID]
	selected := atomic.LoadUint32(&course.Selected)
	err := writeText(ctx, conn, fmt.Sprintf("M %d %d", courseID, selected))
	if err != nil {
		return fmt.Errorf(
			"error sending to websocket for course selected update: %w",
			err,
		)
	}
	return nil
}
