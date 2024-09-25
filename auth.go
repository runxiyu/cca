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
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var myKeyfunc keyfunc.Keyfunc

var errInsufficientFields = errors.New("insufficient fields")

const tokenLength = 20

/*
 * These are the claims in the JSON Web Token received from the client, after
 * it redirects from the authorize endpoint. Some of these fields must be
 * explicitly selected in the Azure app registration and might appear as
 * zero strings if it hasn't been configured correctly.
 */
type msclaimsT struct {
	Name  string `json:"name"`  /* Scope: profile */
	Email string `json:"email"` /* Scope: email   */
	Oid   string `json:"oid"`   /* Scope: profile */
	jwt.RegisteredClaims
}

func generateAuthorizationURL() (string, error) {
	/*
	 * TODO: Handle nonces and anti-replay. Incremental nonces would be
	 * nice on memory and speed (depending on how maps are implemented in
	 * Go, hopefully it's some sort of btree), but that requires either
	 * hacky atomics or having a multiple goroutines to handle
	 * authentication, neither of which are desirable.
	 */
	nonce, err := random(tokenLength)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(
		"https://login.microsoftonline.com/ddd3d26c-b197-4d00-a32d-1ffd84c0c295/oauth2/authorize?client_id=%s&response_type=id_token%%20code&redirect_uri=%s%%2Fauth&response_mode=form_post&scope=openid+profile+email+User.Read&nonce=%s", // hybrid auth flow
		config.Auth.Client,
		config.URL,
		nonce,
	), nil
}

/*
 * Handles redirects to the /auth endpoint from the authorize endpoint.
 * Expects JSON Web Keys to be already set up correctly; if myKeyfunc is null,
 * a null pointer is dereferenced and the thread panics.
 */
func handleAuth(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		wstr(w, http.StatusMethodNotAllowed, "Only POST is supported on the authentication endpoint")
		return
	}

	err := req.ParseForm()
	if err != nil {
		wstr(w, http.StatusBadRequest, "Malformed form data")
		return
	}

	returnedError := req.PostFormValue("error")
	if returnedError != "" {
		returnedErrorDescription := req.PostFormValue("error_description")
		if returnedErrorDescription == "" {
			wstr(w, http.StatusBadRequest, returnedError)
			return
		}
		wstr(w, http.StatusBadRequest, fmt.Sprintf(
			"%s: %s",
			returnedError,
			returnedErrorDescription,
		))
		return
	}

	idTokenString := req.PostFormValue("id_token")
	if idTokenString == "" {
		wstr(w, http.StatusBadRequest, "Missing id_token")
		return
	}

	claimsTemplate := &msclaimsT{} //exhaustruct:ignore
	token, err := jwt.ParseWithClaims(
		idTokenString,
		claimsTemplate,
		myKeyfunc.Keyfunc,
	)
	if err != nil {
		wstr(w, http.StatusBadRequest, "Cannot parse claims")
		return
	}

	switch {
	case token.Valid:
		break
	case errors.Is(err, jwt.ErrTokenMalformed):
		wstr(w, http.StatusBadRequest, "Malformed JWT token")
		return
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		wstr(w, http.StatusBadRequest, "Invalid JWS signature")
		return
	case errors.Is(err, jwt.ErrTokenExpired) ||
		errors.Is(err, jwt.ErrTokenNotValidYet):
		wstr(w, http.StatusBadRequest, "JWT token expired or not yet valid")
		return
	default:
		wstr(w, http.StatusBadRequest, "Unhandled JWT token error")
		return
	}

	claims, claimsOk := token.Claims.(*msclaimsT)

	if !claimsOk {
		wstr(w, http.StatusBadRequest, "Cannot unpack claims")
		return
	}

	authorizationCode := req.PostFormValue("code")

	accessToken, err := getAccessToken(req.Context(), authorizationCode)
	if err != nil {
		wstr(w, http.StatusInternalServerError, "Unable to fetch access token")
		return
	}

	department, err := getDepartment(req.Context(), *(accessToken.Content))
	if err != nil {
		wstr(w, http.StatusInternalServerError, err.Error())
		return
	}

	switch {
	case department == "SJ Co-Curricular Activities Office 松江课外项目办公室" || department == "High School Teaching & Learning 高中教学部门":
		department = "Staff"
	case department == "Y9" || department == "Y10" || department == "Y11" || department == "Y12":
	default:
		wstr(
			w,
			http.StatusForbidden,
			fmt.Sprintf(
				"Your department \"%s\" is unknown.\nWe currently only allow Y9, Y10, Y11, Y12, and the CCA office.",
				department,
			),
		)
		return
	}

	cookieValue, err := random(tokenLength)
	if err != nil {
		wstr(w, http.StatusInternalServerError, err.Error())
		return
	}

	cookie := http.Cookie{
		Name:     "session",
		Value:    cookieValue,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   config.Prod,
		/*
		 * TODO: Cookies should also have an expiration; cookies
		 * without expiration don't even persist across browser
		 * sessions in most browsers.
		 */
	} //exhaustruct:ignore

	http.SetCookie(w, &cookie)

	_, err = db.Exec(
		req.Context(),
		"INSERT INTO users (id, name, email, department) VALUES ($1, $2, $3, $4)",
		claims.Oid,
		claims.Name,
		claims.Email,
		department,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			_, err := db.Exec(
				req.Context(),
				"UPDATE users SET (name, email, department) = ($1, $2, $3) WHERE id = $4",
				claims.Name,
				claims.Email,
				department,
				claims.Oid,
			)
			if err != nil {
				wstr(w, http.StatusInternalServerError, "Database error while updating account.")
				return
			}
		} else {
			wstr(w, http.StatusInternalServerError, "Database error while writing account info.")
			return
		}
	}

	_, err = db.Exec(
		req.Context(),
		"INSERT INTO sessions(userid, cookie, expr) VALUES ($1, $2, $3)",
		claims.Oid,
		cookieValue,
		1881839332, /* TODO */
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			wstr(w, http.StatusInternalServerError, "Cookie collision. Try signing in again.")
			return
		}
		wstr(w, http.StatusInternalServerError, "Database error while inserting session.")
		return
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)
}

/*
 * Setting up JSON Web Keys. Note that myKeyfunc is a global variable.
 */
func setupJwks() error {
	var err error
	myKeyfunc, err = keyfunc.NewDefault([]string{config.Auth.Jwks})
	if err != nil {
		return fmt.Errorf("error setting up jwks: %w", err)
	}
	return nil
}

/*
 * Fetch the department name of the user, mostly to identify which grade
 * a student is in. This expects an accessToken obtained from the OAUTH 2.0
 * token endpoint obtained via an authorization code. It might also be able
 * to use this as part of a hybrid flow that directly provides access tokens,
 * but this flow seems to be only usable for single-page applications according
 * to the Azure portal.
 */
func getDepartment(ctx context.Context, accessToken string) (string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		"https://graph.microsoft.com/v1.0/me?$select=department",
		nil,
	)
	if err != nil {
		return "", fmt.Errorf("error getting department: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{} //exhaustruct:ignore
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error getting department: %w", err)
	}
	defer resp.Body.Close()

	var departmentWrap struct {
		Department *string `json:"department"`
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&departmentWrap)
	if err != nil {
		return "", fmt.Errorf("error getting department: %w", err)
	}

	if departmentWrap.Department == nil {
		/*
		 * This is probably because the response does not contain a
		 * "department" field, which hopefully doesn't occur as we
		 * have specified $select=department in the OData query.
		 */
		return "", fmt.Errorf("error getting department: %w", errInsufficientFields)
	}

	return *(departmentWrap.Department), nil
}

/*
 * TODO: Access token expiration is not checked anywhere.
 */
type accessTokenT struct {
	OriginalExpiresIn *int `json:"expires_in"` /* Original time to expiration */
	Expiration        time.Time
	Content           *string `json:"access_token"`
}

/*
 * Obtain an access token from the token endpoint with an existing
 * authorization code.
 */
func getAccessToken(ctx context.Context, authorizationCode string) (accessTokenT, error) {
	var accessToken accessTokenT
	t := time.Now()
	v := url.Values{}
	v.Set("client_id", config.Auth.Client)
	v.Set("scope", "https://graph.microsoft.com/User.Read")
	v.Set("code", authorizationCode)
	v.Set("redirect_uri", config.URL+"/auth")
	v.Set("grant_type", "authorization_code")
	v.Set("client_secret", config.Auth.Secret)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, config.Auth.Token, strings.NewReader(v.Encode()))
	if err != nil {
		return accessToken, fmt.Errorf("error making access token request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return accessToken, fmt.Errorf("error requesting access token: %w", err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&accessToken)
	if err != nil {
		return accessToken, fmt.Errorf("error decoding access token: %w", err)
	}
	if accessToken.Content == nil || accessToken.OriginalExpiresIn == nil {
		return accessToken, fmt.Errorf("error decoding access token: %w", errInsufficientFields)
	}
	accessToken.Expiration = t.Add(time.Duration(*(accessToken.OriginalExpiresIn)) * time.Second)

	return accessToken, nil
}
