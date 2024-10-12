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
	"fmt"

	"github.com/coder/websocket"
)

func messageConfirm(
	ctx context.Context,
	c *websocket.Conn,
	reportError reportErrorT,
	mar []string,
	userID string,
	department string,
	userCourseTypes *userCourseTypesT,
) error {
	_ = mar

	select {
	case <-ctx.Done():
		return wrapError(
			errContextCanceled,
			ctx.Err(),
		)
	default:
	}

	for courseType := range courseTypes {
		minimum, err := getCourseTypeMinimumForYearGroup(department, courseType)
		if err != nil {
			return reportError("Invalid year group or course type, something is broken")
		}
		if (*userCourseTypes)[courseType] < minimum {
			return writeText(
				ctx,
				c,
				fmt.Sprintf(
					"RC :Cannot confirm choices: You chose %d out of required %d of type %s",
					(*userCourseTypes)[courseType],
					minimum,
					courseType,
				),
			)
		}
	}

	_, err := db.Exec(
		ctx,
		"UPDATE users SET confirmed = true WHERE id = $1",
		userID,
	)
	if err != nil {
		return reportError("error updating database setting confirmation")
	}

	return writeText(
		ctx,
		c,
		"YC",
	)
}
