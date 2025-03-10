/*
 * Custom OAUTH 2.0 implementation for the CCA Selection Service
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var myKeyfunc keyfunc.Keyfunc

const tokenLength = 20

/*
 * These are the claims in the JSON Web Token received from the client, after
 * it redirects from the authorize endpoint. Some of these fields must be
 * explicitly selected in the Azure app registration and might appear as
 * zero strings if it hasn't been configured correctly.
 */
type msclaimsT struct {
	Name   string   `json:"name"`
	Email  string   `json:"email"`
	Oid    string   `json:"oid"`
	Groups []string `json:"groups"`
	jwt.RegisteredClaims
}

func generateAuthorizationURL() (string, error) { // \codelabel{generateAuthorizationURL}
	nonce, err := randomString(tokenLength)
	if err != nil {
		return "", err
	}
	/*
	 * Note that here we use a hybrid authentication flow to obtain an
	 * id token for authentication and an authorization code. The
	 * authorization code may be used like any other; i.e., it may be used
	 * to obtain an access token directly, or the refresh token may be used
	 * to gain persistent access to the upstream API. Sometimes I wish that
	 * the JWT in id token could have more claims. The only reason we
	 * presently use a hybrid flow is to use the authorization code to
	 * obtain an access code to call the user info endpoint to fetch the
	 * user's department information.
	 */
	return fmt.Sprintf(
		"https://login.microsoftonline.com/ddd3d26c-b197-4d00-a32d-1ffd84c0c295/oauth2/authorize?client_id=%s&response_type=id_token%%20code&redirect_uri=%s%%2Fauth&response_mode=form_post&scope=openid+profile+email+User.Read&nonce=%s",
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
func handleAuth(w http.ResponseWriter, req *http.Request) (string, int, error) { // \codelabel{handleAuth}
	if req.Method != http.MethodPost {
		return "", http.StatusMethodNotAllowed, errors.New("only POST is allowed here")
	}

	err := req.ParseForm()
	if err != nil {
		return "", http.StatusBadRequest, fmt.Errorf("malformed form: %w", err)
	}

	returnedError := req.PostFormValue("error")
	if returnedError != "" {
		returnedErrorDescription := req.PostFormValue("error_description")
		return "", http.StatusUnauthorized, fmt.Errorf("jwt auth returned error: %v: %v", returnedError, returnedErrorDescription)
	}

	idTokenString := req.PostFormValue("id_token")
	if idTokenString == "" {
		return "", http.StatusUnauthorized, fmt.Errorf("insufficient fields: id_token")
	}

	claimsTemplate := &msclaimsT{} //exhaustruct:ignore
	token, err := jwt.ParseWithClaims(
		idTokenString,
		claimsTemplate,
		myKeyfunc.Keyfunc,
	)
	if err != nil {
		return "", http.StatusBadRequest, fmt.Errorf("parse jwt claims: %w", err)
	}

	switch {
	case token.Valid:
		break
	case errors.Is(err, jwt.ErrTokenMalformed):
		return "", http.StatusBadRequest, fmt.Errorf("malformed jwt: %w", err)
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		return "", http.StatusBadRequest, fmt.Errorf("invalid jwt signature: %w", err)
	case errors.Is(err, jwt.ErrTokenExpired) ||
		errors.Is(err, jwt.ErrTokenNotValidYet):
		return "", http.StatusBadRequest, fmt.Errorf("invalid jwt timing: %w", err)
	default:
		return "", http.StatusBadRequest, fmt.Errorf("invalid jwt: %w", err)
	}

	claims, claimsOk := token.Claims.(*msclaimsT)

	if !claimsOk {
		return "", http.StatusBadRequest, errors.New("failed to unpack claims")
	}

	var department string
	var ok bool
	department, ok = getDepartmentByUserIDOverride(claims.Oid)
	if !ok {
		department, ok = getDepartmentByGroups(claims.Groups)
		if !ok {
			return "", http.StatusBadRequest, errors.New("unknown department")
		}
	}

	cookieValue, err := randomString(tokenLength)
	if err != nil {
		return "", -1, err
	}

	now := time.Now()
	expr := now.Add(time.Duration(config.Auth.Expr) * time.Second)
	exprU := expr.Unix()

	cookie := http.Cookie{
		Name:     "session",
		Value:    cookieValue,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   config.Prod,
		Expires:  expr,
	} //exhaustruct:ignore

	http.SetCookie(w, &cookie)

	_, err = db.Exec(
		req.Context(),
		"INSERT INTO users (id, name, email, department, session, expr, confirmed) VALUES ($1, $2, $3, $4, $5, $6, false)",
		claims.Oid,
		claims.Name,
		claims.Email,
		department,
		cookieValue,
		exprU,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgErrUniqueViolation {
			_, err := db.Exec(
				req.Context(),
				"UPDATE users SET (name, email, department, session, expr) = ($1, $2, $3, $4, $5) WHERE id = $6",
				claims.Name,
				claims.Email,
				department,
				cookieValue,
				exprU,
				claims.Oid,
			)
			if err != nil {
				return "", -1, fmt.Errorf("update user: %w", err)
			}
		} else {
			return "", -1, fmt.Errorf("insert user: %w", err)
		}
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)

	return "", -1, nil
}

func setupJwks() error {
	var err error
	myKeyfunc, err = keyfunc.NewDefault([]string{config.Auth.Jwks})
	if err != nil {
		return fmt.Errorf("setup jwks: %w", err)
	}
	return nil
}

func getDepartmentByGroups(groups []string) (string, bool) {
	for _, g := range groups {
		d, ok := config.Auth.Departments[g]
		if ok {
			return d, true
		}
	}
	return "", false
}

func getDepartmentByUserIDOverride(userID string) (string, bool) {
	d, ok := config.Auth.Udepts[userID]
	if ok {
		return d, true
	}
	return "", false
}
