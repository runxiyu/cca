/*
 * WebSocket auxiliary functions
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
	"log"
	"sync/atomic"

	"github.com/coder/websocket"
)

/*
 * The message format is a WebSocket message separated with spaces.
 * The contents of each field could contain anything other than spaces,
 * The first character of each argument cannot be a colon. As an exception, the
 * last argument may contain spaces and the first character thereof may be a
 * colon, if the argument is prefixed with a colon. The colon used for the
 * prefix is not considered part of the content of the message. For example, in
 *
 *    SQUISH POP :cat purr!!
 *
 * the first field is "SQUISH", the second field is "POP", and the third
 * field is "cat purr!!".
 *
 * It is essentially an RFC 1459 IRC message without trailing CR-LF and
 * without prefixes. See section 2.3.1 of RFC 1459 for an approximate
 * BNF representation.
 *
 * The reason this was chosen instead of using protobuf etc. is that it
 * is simple to parse without external libraries, and it also happens to
 * be a format I'm very familiar with, having extensively worked with the
 * IRC protocol.
 */

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

func propagateSelectedUpdate(course *courseT) {
	course.Usems.Range(func(key, value interface{}) bool {
		_ = key
		usem, ok := value.(*usemT)
		if !ok {
			panic("Usems contains non-\"*usemT\" value")
		}
		usem.set()
		return true
	})
}

func sendSelectedUpdate(
	ctx context.Context,
	conn *websocket.Conn,
	courseID int,
) error {
	_course, ok := courses.Load(courseID)
	if !ok {
		return fmt.Errorf("%w: %d", errNoSuchCourse, courseID)
	}
	course, ok := _course.(*courseT)
	if !ok {
		panic("courses map has non-\"*courseT\" items")
	}
	if course == nil {
		return fmt.Errorf("%w: %d", errNoSuchCourse, courseID)
	}
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

func propagate(msg string) {
	chanPool.Range(func(_userID, _ch interface{}) bool {
		ch, ok := _ch.(*chan string)
		if !ok {
			panic("chanPool has non-\"*chan string\" key")
		}
		select {
		case *ch <- msg:
		default:
			userID, ok := _userID.(string)
			if !ok {
				panic("chanPool has non-string key")
			}
			log.Println("WARNING: SendQ exceeded for " + userID)
		}
		return true
	})
}

func writeText(ctx context.Context, c *websocket.Conn, msg string) error {
	err := c.Write(ctx, websocket.MessageText, []byte(msg))
	if err != nil {
		return fmt.Errorf("%w: %w", errWebSocketWrite, err)
	}
	return nil
}
