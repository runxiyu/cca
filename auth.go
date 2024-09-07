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
	"encoding/json"
	"errors"
	"fmt"
	// "io"
	"net/http"
	"net/url"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var myKeyfunc keyfunc.Keyfunc

type msclaims_t struct {
	Name  string `json:"name"`  /* Scope: profile */
	Email string `json:"email"` /* Scope: email   */
	Oid   string `json:"oid"`   /* Scope: profile */
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
		"https://login.microsoftonline.com/ddd3d26c-b197-4d00-a32d-1ffd84c0c295/oauth2/authorize?client_id=%s&response_type=id_token%%20code&redirect_uri=%s%%2Fauth&response_mode=form_post&scope=openid+profile+email+User.Read&nonce=%s", // hybrid auth flow
		config.Auth.Client,
		config.Url,
		nonce,
	)
}

func handleAuth(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		wstr(w, 405, "Only POST is supported on the authentication endpoint")
		return
	}

	err := req.ParseForm()
	if err != nil {
		wstr(w, 400, "Malformed form data")
		return
	}

	returned_error := req.PostFormValue("error")
	if returned_error != "" {
		returned_error_description := req.PostFormValue("error_description")
		if returned_error_description == "" {
			wstr(w, 400, fmt.Sprintf("%s", returned_error))
			return
		} else {
			wstr(w, 400, fmt.Sprintf("%s: %s", returned_error, returned_error_description))
			return
		}
	}

	id_token_string := req.PostFormValue("id_token")
	if id_token_string == "" {
		wstr(w, 400, "Missing id_token")
		return
	}

	token, err := jwt.ParseWithClaims(
		id_token_string,
		&msclaims_t{},
		myKeyfunc.Keyfunc,
	)
	if err != nil {
		wstr(w, 400, "Cannot parse claims")
		return
	}

	switch {
	case token.Valid:
		break
	case errors.Is(err, jwt.ErrTokenMalformed):
		wstr(w, 400, "Malformed JWT token")
		return
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		wstr(w, 400, "Invalid JWS signature")
		return
	case errors.Is(err, jwt.ErrTokenExpired) ||
		errors.Is(err, jwt.ErrTokenNotValidYet):
		wstr(w, 400, "JWT token expired or not yet valid")
		return
	default:
		wstr(w, 400, "Unhandled JWT token error")
		return
	}

	claims, claims_ok := token.Claims.(*msclaims_t)

	if !claims_ok {
		wstr(w, 400, "Cannot unpack claims")
		return
	}

	authorization_code := req.PostFormValue("code")

	access_token, err := getAccessToken(authorization_code)
	if err != nil {
		wstr(w, 500, "Unable to fetch access token")
		return
	}

	// TODO: validate access token

	department, err := getDepartment(*(access_token.Content))
	if err != nil {
		wstr(w, 500, err.Error())
		return
	}

	if department == "SJ Co-Curricular Activities Office 松江课外项目办公室" || department == "High School Teaching & Learning 高中教学部门" {
		department = "Staff"
	} else if department == "Y9" || department == "Y10" || department == "Y11" || department == "Y12" {
	} else {
		wstr(
			w,
			403,
			fmt.Sprintf(
				"Your department \"%s\" is unknown.\nWe currently only allow Y9, Y10, Y11, Y12, and the CCA office.",
				department,
			),
		)
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

	_, err = db.Exec(
		context.Background(),
		"INSERT INTO users (id, name, email, department) VALUES ($1, $2, $3, $4)",
		claims.Oid,
		claims.Name,
		claims.Email,
		department,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				_, err := db.Exec(
					context.Background(),
					"UPDATE users SET (name, email, department) = ($1, $2, $3) WHERE id = $4",
					claims.Name,
					claims.Email,
					department,
					claims.Oid,
				)
				if err != nil {
					wstr(w, 500, "Database error while updating account.")
					return
				}
			}
		} else {
			wstr(w, 500, "Database error while writing account info.")
			return
		}
	}

	_, err = db.Exec(
		context.Background(),
		"INSERT INTO sessions(userid, cookie, expr) VALUES ($1, $2, $3)",
		claims.Oid,
		cookie_value,
		1881839332, /* TODO */
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			wstr(w, 500, "Cookie collision. Try signing in again.")
			return
		} else {
			wstr(w, 500, "Database error while inserting session.")
			return
		}
	}

	http.Redirect(w, req, "/", 303)

}

func setupJwks() error {
	var err error
	myKeyfunc, err = keyfunc.NewDefault([]string{config.Auth.Jwks})
	if err != nil {
		return err
	}
	return nil
}

func getDepartment(access_token string) (string, error) {
	req, err := http.NewRequest("GET", "https://graph.microsoft.com/v1.0/me?$select=department", nil)
	if err != nil {
		return "", errors.New("Cannot make the Graph API request")
	}
	req.Header.Set("Authorization", "Bearer "+access_token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.New("Graph API request failed")
	}
	defer resp.Body.Close()

	var departmentWrap struct {
		Department *string `json:"department"`
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&departmentWrap)
	if err != nil {
		return "", errors.New("Department unmarshaling failed")
	}

	if departmentWrap.Department == nil {
		return "", errors.New("Department pointer is nil")
	}

	return *(departmentWrap.Department), nil
}

type access_token_t struct {
	OriginalExpiresIn *int `json:"expires_in"` // When the access token is obtained
	Expiration        time.Time
	Content           *string `json:"access_token"`
}

func getAccessToken(authorization_code string) (access_token_t, error) {
	var access_token access_token_t
	t := time.Now()
	v := url.Values{}
	v.Set("client_id", config.Auth.Client)
	v.Set("scope", "https://graph.microsoft.com/User.Read")
	v.Set("code", authorization_code)
	v.Set("redirect_uri", config.Url+"/auth")
	v.Set("grant_type", "authorization_code")
	v.Set("client_secret", config.Auth.Secret)
	resp, err := http.PostForm(config.Auth.Token, v)
	if err != nil {
		return access_token, err
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&access_token)
	if err != nil {
		return access_token, err
	}
	if access_token.Content == nil || access_token.OriginalExpiresIn == nil {
		return access_token, errors.New("Insufficient fields")
	}
	access_token.Expiration = t.Add(time.Duration(*(access_token.OriginalExpiresIn)) * time.Second)

	return access_token, nil
}
