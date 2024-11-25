/*
 * WebSocket auxiliary functions
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"context"
	"fmt"
	"log/slog"
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
			slog.Warn(
				"sendq",
				"user", userID,
				"msg", msg,
			)
		}
		return true
	})
}

func writeText(ctx context.Context, c *websocket.Conn, msg string) error {
	err := c.Write(ctx, websocket.MessageText, []byte(msg))
	if err != nil {
		return wrapError(errWebSocketWrite, err)
	}
	return nil
}
