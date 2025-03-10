/*
 * Confirmination checking
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"context"
	"fmt"
)

func getConfirmedStatus(
	ctx context.Context,
	userID string,
) (confirmed bool, retErr error) {
	err := db.QueryRow(
		ctx,
		"SELECT confirmed FROM users WHERE id = $1",
		userID,
	).Scan(&confirmed)
	if err != nil {
		retErr = fmt.Errorf("get confirmed status: %w", err)
	}
	return
}
