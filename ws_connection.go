/*
 * WebSocket connection routine
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"context"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"

	"github.com/coder/websocket"
)

type errbytesT struct {
	err   error
	bytes *[]byte
}

var usemCount int64 /* atomic */

/*
 * This is more appropriately typed as uint64, but it needs to be cast to int64
 * later anyway due to time.Duration, so let's just use int64.
 */

/*
 * The actual logic in handling the connection, after authentication has been
 * completed.
 */
func handleConn(
	ctx context.Context,
	c *websocket.Conn,
	userID string,
	department string,
) error {
	if atomic.LoadUint32(states[department]) == 0 {
		return errStudentAccessDisabled
	}

	send := make(chan string, config.Perf.SendQ)
	chanSubPool, ok := chanPool[department]
	if !ok {
		return errNoSuchYearGroup
	}
	chanSubPool.Store(userID, &send)
	defer chanSubPool.CompareAndDelete(userID, &send)

	newCtx, newCancel := context.WithCancel(ctx)

	_cancel, ok := cancelPool.Load(userID)
	if ok {
		cancel, ok := _cancel.(*context.CancelFunc)
		if ok && cancel != nil {
			(*cancel)()
		}
		/* TODO: Make the cancel synchronous */
	}
	cancelPool.Store(userID, &newCancel)

	defer func() {
		cancelPool.CompareAndDelete(userID, &newCancel)
	}()

	/* TODO: Tell the user their current choices here. Deprecate HELLO. */

	usems := make(map[int]*usemT)

	/* TODO: Check if the LoadUint32 here is a bit too much overhead */
	atomic.AddInt64(&usemCount, int64(atomic.LoadUint32(&numCourses)))
	courses.Range(func(key, value interface{}) bool {
		/* TODO: Remember to change this too when changing the courseID type */
		courseID, ok := key.(int)
		if !ok {
			panic("courses map has non-\"int\" keys")
		}
		course, ok := value.(*courseT)
		if !ok {
			panic("courses map has non-\"*courseT\" items")
		}
		usem := &usemT{} //exhaustruct:ignore
		usem.init()
		course.Usems.Store(userID, usem)
		usems[courseID] = usem
		return true
	})

	defer func() {
		courses.Range(func(key, value interface{}) bool {
			_ = key
			course, ok := value.(*courseT)
			if !ok {
				panic("courses map has non-\"*courseT\" items")
			}
			course.Usems.Delete(userID)
			return true
		})
		atomic.AddInt64(&usemCount, -int64(atomic.LoadUint32(&numCourses)))
	}()

	usemParent := make(chan int)
	for courseID, usem := range usems {
		go func() {
			defer func() {
				if e := recover(); e != nil {
					slog.Error("panic", "arg", e)
				}
			}()

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
				time.Sleep(
					time.Duration(
						atomic.LoadInt64(&usemCount)>>
							config.Perf.UsemDelayShiftBits,
					) * time.Millisecond,
				)
			}
		}()
	}

	var userCourseGroups userCourseGroupsT = make(map[string]struct{})
	var userCourseTypes userCourseTypesT = make(map[string]int)
	err := populateUserCourseTypesAndGroups(newCtx, &userCourseTypes, &userCourseGroups, userID)
	if err != nil {
		return err
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
		defer func() {
			if e := recover(); e != nil {
				slog.Error("panic", "arg", e)
			}
		}()
		for {
			/*
			 * Here we use the original connection context instead
			 * of the new context we just created. Apparently when
			 * the context passed to Read expires, the connection
			 * gets closed, which makes it impossible for us to
			 * write the context expiry message to the client.
			 * So we pass the original connection context, which
			 * would get canceled anyway once we close the
			 * connection.
			 * See: https://github.com/coder/websocket/issues/242
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
				select {
				case <-newCtx.Done():
					_ = writeText(
						ctx,
						c,
						"E :Context canceled",
					)
					/* Not a typo to use ctx here */
					return
				case recv <- &errbytesT{err: err, bytes: nil}:
				}
				return
			}
			select {
			case <-newCtx.Done():
				_ = writeText(ctx, c, "E :Context canceled")
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
			 * We select this context done channel when entering
			 * other cases too (see below) because we need to
			 * make sure the context cancel works even if both
			 * the cancel signal and another event arrive while
			 * processing a select cycle.
			 */
			return wrapError(
				errWsHandlerContextCanceled,
				newCtx.Err(),
			)
		case sendText := <-send:
			select {
			case <-newCtx.Done():
				return wrapError(
					errWsHandlerContextCanceled,
					newCtx.Err(),
				)
			default:
			}

			err := writeText(newCtx, c, sendText)
			if err != nil {
				return err
			}
		case courseID := <-usemParent:
			select {
			case <-newCtx.Done():
				return wrapError(
					errWsHandlerContextCanceled,
					newCtx.Err(),
				)
			default:
			}

			err := sendSelectedUpdate(newCtx, c, courseID)
			if err != nil {
				return wrapError(
					errCannotSend,
					err,
				)
			}
			continue
		case errbytes := <-recv:
			select {
			case <-newCtx.Done():
				return wrapError(
					errWsHandlerContextCanceled,
					newCtx.Err(),
				)
			default:
			}

			if atomic.LoadUint32(states[department]) == 0 {
				return errStudentAccessDisabled
			}

			if errbytes.err != nil {
				return wrapError(
					errCannotReceiveMessage,
					errbytes.err,
				)
				/*
				 * Note that this cannot return newCtx.Err(),
				 * so we handle the error reporting in the
				 * reading routine
				 */
			}
			mar = splitMsg(errbytes.bytes)
			switch mar[0] {
			case "HELLO":
				err := messageHello(
					newCtx,
					c,
					mar,
					userID,
					department,
				)
				if err != nil {
					return err
				}
			case "Y":
				err := messageChooseCourse(
					newCtx,
					c,
					mar,
					userID,
					department,
					&userCourseGroups,
					&userCourseTypes,
				)
				if err != nil {
					return err
				}
			case "N":
				err := messageUnchooseCourse(
					newCtx,
					c,
					mar,
					userID,
					department,
					&userCourseGroups,
					&userCourseTypes,
				)
				if err != nil {
					return err
				}
			case "YC":
				err := messageConfirm(
					newCtx,
					c,
					mar,
					userID,
					department,
					&userCourseTypes,
				)
				if err != nil {
					return err
				}
			case "NC":
				err := messageUnconfirm(
					newCtx,
					c,
					mar,
					userID,
					department,
				)
				if err != nil {
					return err
				}
			default:
				return wrapAny(errUnknownCommand, mar[0])
			}
		}
	}
}

var cancelPool sync.Map /* string, *context.CancelFunc */

var chanPool = map[string]*sync.Map{
	"Y9":  {},
	"Y10": {},
	"Y11": {},
	"Y12": {},
} /* string, *chan string */
