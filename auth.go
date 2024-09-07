/*
 * Custom OAUTH 2.0 implementation for the CCA Selection Service
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: BSD-2-Clause
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are
 * met:
 *
 *     1. Redistributions of source code must retain the above copyright
 *     notice, this list of conditions and the following disclaimer.
 *
 *     2. Redistributions in binary form must reproduce the above copyright
 *     notice, this list of conditions and the following disclaimer in the
 *     documentation and/or other materials provided with the distribution.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS "AS IS" AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR
 * PURPOSE ARE DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR
 * CONTRIBUTORS BE LIABLE FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL,
 * EXEMPLARY, OR CONSEQUENTIAL DAMAGES (INCLUDING, BUT NOT LIMITED TO,
 * PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES; LOSS OF USE, DATA, OR
 * PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND ON ANY THEORY OF
 * LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT (INCLUDING
 * NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var myKeyfunc keyfunc.Keyfunc

type msclaims_t struct {
	Name  string `json:"name"`  /* Scope: profile */
	Email string `json:"email"` /* Scope: email   */
	jwt.RegisteredClaims
}

func generate_authorization_url() string {
	/*
	 * TODO: Handle nonces and anti-replay. Incremental nonces would be
	 * nice on memory and speed (depending on how maps are implemented in
	 * Go, hopefully it's some sort of btree), but that requires either
	 * hacky atomics or having a multiple goroutines to handle
	 * authentication, neither of which are desirable.
	 */
	nonce := random(30)
	return fmt.Sprintf(
		"https://login.microsoftonline.com/ddd3d26c-b197-4d00-a32d-1ffd84c0c295/oauth2/authorize?client_id=%s&response_type=id_token&redirect_uri=%s/auth&response_mode=form_post&scope=openid+profile+email&nonce=%s",
		config.Auth.Client,
		config.Url,
		nonce,
	)
}

func handleAuth(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		wstr(w, 405, "Error: Only POST is supported on the authentication endpoint")
		return
	}

	err := req.ParseForm()
	if err != nil {
		wstr(w, 400, "Error: Malformed form data")
		return
	}

	returned_error := req.PostFormValue("error")
	if returned_error != "" {
		returned_error_description := req.PostFormValue("error_description")
		if returned_error_description == "" {
			wstr(w, 400, fmt.Sprintf("Error: %s", returned_error))
			return
		} else {
			wstr(w, 400, fmt.Sprintf("Error: %s: %s", returned_error, returned_error_description))
			return
		}
	}

	id_token_string := req.PostFormValue("id_token")
	if id_token_string == "" {
		wstr(w, 400, "Error: Missing id_token")
		return
	}

	token, err := jwt.ParseWithClaims(
		id_token_string,
		&msclaims_t{},
		myKeyfunc.Keyfunc,
	)
	if err != nil {
		wstr(w, 400, "Error: Cannot parse claims")
		return
	}

	switch {
	case token.Valid:
		break
	case errors.Is(err, jwt.ErrTokenMalformed):
		wstr(w, 400, "Error: Malformed JWT token")
		return
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		wstr(w, 400, "Error: Invalid JWS signature")
		return
	case errors.Is(err, jwt.ErrTokenExpired) ||
		errors.Is(err, jwt.ErrTokenNotValidYet):
		wstr(w, 400, "Error: JWT token expired or not yet valid")
		return
	default:
		wstr(w, 400, "Error: Unhandled JWT token error")
		return
	}

	claims, claims_ok := token.Claims.(*msclaims_t)

	if !claims_ok {
		wstr(w, 400, "Error: Cannot unpack claims")
		return
	}

	cookie_value := random(20)

	cookie := http.Cookie{
		Name:     "session",
		Value:    cookie_value,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   config.Prod,
	}

	http.SetCookie(w, &cookie)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	_, err = db.Exec(
		context.Background(),
		"INSERT INTO users (id, name, email) VALUES ($1, $2, $3)",
		claims.Subject,
		claims.Name,
		claims.Email,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				_, err := db.Exec(
					context.Background(),
					"UPDATE users SET (name, email) = ($1, $2) WHERE id = $3",
					claims.Name,
					claims.Email,
					claims.Subject,
				)
				if err != nil {
					w.Header().Set("Content-Type", "text/plain; charset=utf-8")
					w.WriteHeader(500)
					w.Write([]byte(fmt.Sprintf("Error\nDatabase error while updating your account.\n%s\n", err)))
					return
				}
			}
		} else {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("Error\nDatabase error while attempting to insert account info.\n%s\n", err)))
			return
		}
	}

	_, err = db.Exec(
		context.Background(),
		"INSERT INTO sessions(userid, cookie, expr) VALUES ($1, $2, $3)",
		claims.Subject,
		cookie_value,
		1881839332, /* TODO */
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("Error\nCookie collision! Could you try signing in again?\n%s\n", err)))
			return
		} else {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("Error\nDatabase error while attempting to insert session info.\n%s\n", err)))
			return
		}
	}

	http.Redirect(w, req, "/", 303)

	return

}

func setupJwks() error {
	var err error
	myKeyfunc, err = keyfunc.NewDefault([]string{config.Auth.Jwks})
	if err != nil {
		return err
	}
	return nil
}
