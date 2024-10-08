/*
 * Database handling
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
		return fmt.Errorf("%w: %w", errUnexpectedDBError, err)
	}
	return nil
}
