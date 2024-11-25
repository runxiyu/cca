/*
 * Database handling
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

var db *pgxpool.Pool

const pgErrUniqueViolation = "23505"

/*
 * This must be run during setup, before the database is accessed by any
 * means. Otherwise, db would be a null pointer.
 */
func setupDatabase() error {
	var err error
	if config.DB.Type != "postgres" {
		return errUnsupportedDatabaseType
	}
	db, err = pgxpool.New(context.Background(), config.DB.Conn)
	if err != nil {
		return wrapError(errUnexpectedDBError, err)
	}
	return nil
}
