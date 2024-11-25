/*
 * Increase-unblocking capped semaphores
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

/*
 * usemT is basically a semaphore capped at 1. Adding is always non-blocking;
 * adding it multiple times without a read in between is equivalent to setting
 * it once. Reading blocks after the first read after the last set.
 */

type usemT struct {
	ch (chan struct{})
}

func (s *usemT) init() {
	s.ch = make(chan struct{}, 1)
}

func (s *usemT) set() {
	select {
	case s.ch <- struct{}{}:
	default:
	}
}
