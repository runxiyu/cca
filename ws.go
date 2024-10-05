/*
 * Primary WebSocket routines
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

package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func writeText(ctx context.Context, c *websocket.Conn, msg string) error {
	err := c.Write(ctx, websocket.MessageText, []byte(msg))
	if err != nil {
		return fmt.Errorf("error writing to connection: %w", err)
	}
	return nil
}

/*
 * Handle requests to the WebSocket endpoint and establish a connection.
 * Authentication is handled here, but afterwards, the connection is really
 * handled in handleConn.
 */
func handleWs(w http.ResponseWriter, req *http.Request) {
	wsOptions := &websocket.AcceptOptions{
		Subprotocols: []string{"cca1"},
	} //exhaustruct:ignore
	c, err := websocket.Accept(
		w,
		req,
		wsOptions,
	)
	if err != nil {
		wstr(w, http.StatusBadRequest, "This endpoint only supports valid WebSocket connections.")
		return
	}
	defer func() {
		_ = c.CloseNow()
	}()

	fake := false

	sessionCookie, err := req.Cookie("session")
	if errors.Is(err, http.ErrNoCookie) {
		if config.Auth.Fake == 0 {
			err := writeText(req.Context(), c, "U")
			if err != nil {
				log.Println(err)
			}
			return
		}
		fake = true
	} else if err != nil {
		err := writeText(req.Context(), c, "E :Error fetching cookie")
		if err != nil {
			log.Println(err)
		}
		return
	}

	var userID string
	var session string
	var expr int

	if fake {
		switch config.Auth.Fake {
		case 9080:
			_uuid, err := uuid.NewRandom()
			if err != nil {
				log.Println(err)
				return
			}
			userID = _uuid.String()
		case 4712:
			userID = "fake"
		default:
			panic("not supposed to happen")
		}
		session, err = randomString(20)
		if err != nil {
			log.Println(err)
			return
		}
		_, err = db.Exec(
			req.Context(),
			"INSERT INTO users (id, name, email, department, session, expr) VALUES ($1, $2, $3, $4, $5, $6)",
			userID,
			"Fake User",
			"fake@runxiyu.org",
			"Y11",
			session,
			time.Now().Add(time.Duration(config.Auth.Expr)*time.Second).Unix(),
		)
		if err != nil && config.Auth.Fake != 4712 {
			/* TODO check pgerr */
			err := writeText(req.Context(), c, "E :Database error while writing fake account info")
			if err != nil {
				log.Println(err)
			}
			return
		}
		err = writeText(req.Context(), c, "FAKE "+userID+" "+session)
		if err != nil {
			log.Println(err)
			return
		}
	} else {
		session = sessionCookie.Value
		err = db.QueryRow(
			req.Context(),
			"SELECT id, expr FROM users WHERE session = $1",
			session,
		).Scan(&userID, &expr)
		if errors.Is(err, pgx.ErrNoRows) {
			err := writeText(req.Context(), c, "U")
			if err != nil {
				log.Println(err)
			}
			return
		} else if err != nil {
			err := writeText(req.Context(), c, "E :Database error while selecting session")
			if err != nil {
				log.Println(err)
			}
			return
		}
	}

	/*
	 * Now that we have an authenticated request, this WebSocket connection
	 * may be simply associated with the session and userID.
	 * TODO: There are various race conditions that could occur if one user
	 * creates multiple connections, with the same or different session
	 * cookies. The last situation could occur in normal use when a user
	 * opens multiple instances of the page in one browser, and is not
	 * unique to custom clients or malicious users. Some effort must be
	 * taken to ensure that each user may only have one connection at a
	 * time.
	 */
	err = handleConn(
		req.Context(),
		c,
		session,
		userID,
	)
	if err != nil {
		log.Printf("%v", err)
		return
	}
}

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

func baseReportError(ctx context.Context, conn *websocket.Conn, e string) error {
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

type errbytesT struct {
	err   error
	bytes *[]byte
}

var (
	cancelPool = make(map[string](*context.CancelFunc))
	/*
	 * Normal Go maps are not thread safe, so we protect large cancelPool
	 * operations such as addition and deletion under a RWMutex.
	 */
	cancelPoolLock sync.RWMutex
)

func propagateSelectedUpdate(courseID int) {
	course := courses[courseID]
	course.UsemsLock.RLock()
	defer course.UsemsLock.RUnlock()
	for _, usem := range course.Usems {
		usem.set()
	}
}

/*
 * The actual logic in handling the connection, after authentication has been
 * completed.
 */
func handleConn(
	ctx context.Context,
	c *websocket.Conn,
	session string,
	userID string,
) (retErr error) {
	reportError := makeReportError(ctx, c)
	newCtx, newCancel := context.WithCancel(ctx)

	func() {
		cancelPoolLock.Lock()
		defer cancelPoolLock.Unlock()
		cancel := cancelPool[userID]
		if cancel != nil {
			(*cancel)()
			/* TODO: Make the cancel synchronous */
		}
		cancelPool[userID] = &newCancel
	}()
	defer func() {
		cancelPoolLock.Lock()
		defer cancelPoolLock.Unlock()
		if cancelPool[userID] == &newCancel {
			delete(cancelPool, userID)
		}
		if errors.Is(retErr, context.Canceled) {
			/*
			 * Only works if it's newCtx that has been cancelled
			 * rather than the original ctx, which is kinda what
			 * we intend
			 */
			_ = writeText(ctx, c, "E :Context canceled")
		}
	}()

	/* TODO: Tell the user their current choices here. Deprecate HELLO. */

	usems := make(map[int]*usemT)
	func() {
		coursesLock.RLock()
		defer coursesLock.RUnlock()
		for courseID, course := range courses {
			usem := &usemT{} //exhaustruct:ignore
			usem.init()
			func() {
				course.UsemsLock.Lock()
				defer course.UsemsLock.Unlock()
				course.Usems[userID] = usem
			}()
			usems[courseID] = usem
		}
	}()
	defer func() {
		coursesLock.RLock()
		defer coursesLock.RUnlock()
		for _, course := range courses {
			func() {
				course.UsemsLock.Lock()
				defer course.UsemsLock.Unlock()
				delete(course.Usems, userID)
			}()
		}
	}()

	usemParent := make(chan int)
	for courseID, usem := range usems {
		go func() {
			for {
				select {
				case <-newCtx.Done():
					return
				case <-usem.ch:
					select {
					case <-newCtx.Done():
						return
					case usemParent <- courseID:
					}
				}
			}
		}()
	}

	/*
	 * userCourseGroups stores whether the user has already chosen a course
	 * in the courseGroup.
	 */
	var userCourseGroups userCourseGroupsT = make(map[courseGroupT]bool)
	err := populateUserCourseGroups(newCtx, &userCourseGroups, userID)
	if err != nil {
		return reportError(fmt.Sprintf("cannot populate user course groups: %v", err))
	}

	/*
	 * Later we need to select from recv and send and perform the
	 * corresponding action. But we can't just select from c.Read because
	 * the function blocks. Therefore, we must spawn a goroutine that
	 * blocks on c.Read and send what it receives to a channel "recv"; and
	 * then we can select from that channel.
	 */
	recv := make(chan *errbytesT)
	go func() {
		for {
			/*
			 * Here we use the original connection context instead
			 * of the new context we just created. Apparently when
			 * the context passed to Read expires, the connection
			 * gets closed, which makes it impossible for us to
			 * write the context expiry message to the client.
			 * So we pass the original connection context, which
			 * would get cancelled anyway once we close the
			 * connection.
			 * We still need to take care of this while sending so
			 * we don't infinitely block, and leak goroutines and
			 * cause the channel to remain out of reach of the
			 * garbage collector.
			 * It would be nice to return the newCtx.Err() but
			 * the only way to really do that is to use the recv
			 * channel which might not have a listener anymore.
			 * It's not really crucial anyways so we could just
			 * close this goroutine by returning here.
			 */
			_, b, err := c.Read(ctx)
			if err != nil {
				/*
				 * TODO: Prioritize context dones... except
				 * that it's not really possible. I would just
				 * have placed newCtx in here but apparently
				 * that causes the connection to be closed when
				 * the context expires, which makes it
				 * impossible to deliver the final error
				 * message. Probably need to look into this
				 * design again.
				 */
				select {
				case <-newCtx.Done():
					_ = writeText(ctx, c, "E :Context canceled")
					/* Not a typo to use ctx here */
					return
				case recv <- &errbytesT{err: err, bytes: nil}:
				}
				return
			}
			select {
			case <-newCtx.Done():
				_ = writeText(ctx, c, "E :Context cancelled")
				/* Not a typo to use ctx here */
				return
			case recv <- &errbytesT{err: nil, bytes: &b}:
			}
		}
	}()

	for {
		var mar []string
		select {
		case <-newCtx.Done():
			/*
			 * TODO: Somehow prioritize this case over all other cases
			 */
			return fmt.Errorf("context done in main event loop: %w", newCtx.Err())
			/*
			 * There are other times when the context could be
			 * cancelled, and apparently some WebSocket functions
			 * just close the connection when the context is
			 * cancelled. So it's kinda impossible to reliably
			 * send this message due to newCtx cancellation.
			 * But in any case, the WebSocket connection would
			 * be closed, and the user would see the connection
			 * closed page which should explain it.
			 */
		case courseID := <-usemParent:
			var selected int
			func() {
				course := courses[courseID]
				course.SelectedLock.RLock()
				defer course.SelectedLock.RUnlock()
				selected = course.Selected
			}()
			err := writeText(newCtx, c, fmt.Sprintf("M %d %d", courseID, selected))
			if err != nil {
				return fmt.Errorf("error sending to websocket for course selected update: %w", err)
			}
			continue
		case errbytes := <-recv:
			if errbytes.err != nil {
				return fmt.Errorf("error fetching message from recv channel: %w", errbytes.err)
				/*
				 * Note that this cannot return newCtx.Err(),
				 * so we handle the error reporting in the
				 * reading routine
				 */
			}
			mar = splitMsg(errbytes.bytes)
			switch mar[0] {
			case "HELLO":
				err := messageHello(newCtx, c, reportError, mar, userID, session)
				if err != nil {
					return err
				}
			case "Y":
				err := messageChooseCourse(newCtx, c, reportError, mar, userID, session, &userCourseGroups)
				if err != nil {
					return err
				}
			case "N":
				err := messageUnchooseCourse(newCtx, c, reportError, mar, userID, session, &userCourseGroups)
				if err != nil {
					return err
				}
			default:
				return reportError("Unknown command " + mar[0])
			}
		}
	}
}
