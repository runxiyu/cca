/*
 * Utility functions
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"crypto/rand"
	"encoding/base64"
	"log/slog"
	"net/http"
)

func wstr(w http.ResponseWriter, code int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(code)
	_, err := w.Write([]byte(msg))
	if err != nil {
		slog.Error(
			"write",
			"writer", &w,
			"error", err,
		)
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
		return "", wrapError(errCannotGenerateRandomString, err)
	}
	return base64.RawURLEncoding.EncodeToString(r), nil
}

func getKeysOfMap[K comparable, V any](i map[K]V) []K {
	o := make([]K, 0, len(i))
	for k := range i {
		o = append(o, k)
	}
	return o
}

func returnFirst[T1 any, T2 any](v1 T1, _ T2) T1 {
	return v1
}
