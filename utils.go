/*
 * Utility functions
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
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
)

func wstr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	_, err := w.Write([]byte(msg))
	if err != nil {
		log.Printf("Error wstr'ing to writer: %v", err)
	}
}

/*
 * Generate a random url-safe string.
 * Note that the "sz" parameter specifies the number of bytes taken from the
 * random source divided by three and does NOT represent the length of the
 * encoded string. It's divided by three because we're using base64 and it's
 * ideal to ensure that the entropy remains consistent throughout the string.
 */
func randomString(sz int) (string, error) {
	r := make([]byte, 3*sz)
	_, err := rand.Read(r)
	if err != nil {
		return "", fmt.Errorf("%w: %w", errCannotGenerateRandomString, err)
	}
	return base64.RawURLEncoding.EncodeToString(r), nil
}
