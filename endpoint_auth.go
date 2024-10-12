/*
 * Custom OAUTH 2.0 implementation for the CCA Selection Service
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
	nonce, err := randomString(tokenLength)
	if err != nil {
		return "", err
	}
	/*
	 * Note that here we use a hybrid authentication flow to obtain an
	 * id_token for authentication and an authorization code. The
	 * authorization code may be used like any other; i.e., it may be used
	 * to obtain an access token directly, or the refresh token may be used
	 * to gain persistent access to the upstream API. Sometimes I wish that
	 * the JWT in id_token could have more claims. The only reason we
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
func handleAuth(w http.ResponseWriter, req *http.Request) (string, int, error) {
	if req.Method != http.MethodPost {
		return "", http.StatusMethodNotAllowed, errPostOnly
	}

	err := req.ParseForm()
	if err != nil {
		return "", http.StatusBadRequest, wrapError(errMalformedForm, err)
	}

	returnedError := req.PostFormValue("error")
	if returnedError != "" {
		returnedErrorDescription := req.PostFormValue("error_description")
		return "", http.StatusUnauthorized, wrapAny(
			errAuthorizeEndpointError,
			returnedError+": "+returnedErrorDescription,
		)
	}

	idTokenString := req.PostFormValue("id_token")
	if idTokenString == "" {
		return "", http.StatusBadRequest, wrapAny(
			errInsufficientFields,
			"id_token",
		)
	}

	claimsTemplate := &msclaimsT{} //exhaustruct:ignore
	token, err := jwt.ParseWithClaims(
		idTokenString,
		claimsTemplate,
		myKeyfunc.Keyfunc,
	)
	if err != nil {
		return "", http.StatusBadRequest, wrapError(
			errCannotParseClaims,
			err,
		)
	}

	switch {
	case token.Valid:
		break
	case errors.Is(err, jwt.ErrTokenMalformed):
		return "", http.StatusBadRequest, wrapError(
			errJWTMalformed,
			err,
		)
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		return "", http.StatusBadRequest, wrapError(
			errJWTSignatureInvalid,
			err,
		)
	case errors.Is(err, jwt.ErrTokenExpired) ||
		errors.Is(err, jwt.ErrTokenNotValidYet):
		return "", http.StatusBadRequest, wrapError(
			errJWTExpired,
			err,
		)
	default:
		return "", http.StatusBadRequest, wrapError(
			errJWTInvalid,
			err,
		)
	}

	claims, claimsOk := token.Claims.(*msclaimsT)

	if !claimsOk {
		return "", http.StatusBadRequest, errCannotUnpackClaims
	}

	authorizationCode := req.PostFormValue("code")

	accessToken, err := getAccessToken(req.Context(), authorizationCode)
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	department, err := getDepartment(req.Context(), *(accessToken.Content))
	if err != nil {
		return "", http.StatusInternalServerError, err
	}

	switch {
	case department == "SJ Co-Curricular Activities Office 松江课外项目办公室":
		department = staffDepartment
	case department == "Y9" || department == "Y10" ||
		department == "Y11" || department == "Y12":
	default:
		return "", http.StatusForbidden, errUnknownDepartment
	}

	cookieValue, err := randomString(tokenLength)
	if err != nil {
		return "", http.StatusInternalServerError, err
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
				return "", http.StatusInternalServerError, wrapError(errUnexpectedDBError, err)
			}
		} else {
			return "", http.StatusInternalServerError, wrapError(errUnexpectedDBError, err)
		}
	}

	http.Redirect(w, req, "/", http.StatusSeeOther)

	return "", -1, nil
}

func setupJwks() error {
	var err error
	myKeyfunc, err = keyfunc.NewDefault([]string{config.Auth.Jwks})
	if err != nil {
		return wrapError(errCannotSetupJwks, err)
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
		return "", wrapError(errCannotGetDepartment, err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	client := &http.Client{} //exhaustruct:ignore
	resp, err := client.Do(req)
	if err != nil {
		return "", wrapError(errCannotGetDepartment, err)
	}
	defer resp.Body.Close()

	var departmentWrap struct {
		Department *string `json:"department"`
	}

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&departmentWrap)
	if err != nil {
		return "", wrapError(errCannotGetDepartment, err)
	}

	if departmentWrap.Department == nil {
		/*
		 * This is probably because the response does not contain a
		 * "department" field, which hopefully doesn't occur as we
		 * have specified $select=department in the OData query.
		 */
		return "", wrapError(
			errCannotGetDepartment,
			errInsufficientFields,
		)
	}

	return *(departmentWrap.Department), nil
}

type accessTokenT struct {
	Content          *string `json:"access_token"`
	Error            *string `json:"error"`
	ErrorDescription *string `json:"error_description"`
	ErrorCodes       *[]int  `json:"error_codes"`
}

func getAccessToken(
	ctx context.Context,
	authorizationCode string,
) (accessTokenT, error) {
	var accessToken accessTokenT
	v := url.Values{}
	v.Set("client_id", config.Auth.Client)
	v.Set("scope", "https://graph.microsoft.com/User.Read")
	v.Set("code", authorizationCode)
	v.Set("redirect_uri", config.URL+"/auth")
	v.Set("grant_type", "authorization_code")
	v.Set("client_secret", config.Auth.Secret)
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		config.Auth.Token,
		strings.NewReader(v.Encode()),
	)
	if err != nil {
		return accessToken,
			wrapError(errCannotFetchAccessToken, err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return accessToken,
			wrapError(errCannotFetchAccessToken, err)
	}
	defer resp.Body.Close()

	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&accessToken)
	if err != nil {
		return accessToken,
			wrapError(errCannotFetchAccessToken, err)
	}
	if accessToken.Error != nil || accessToken.ErrorCodes != nil ||
		accessToken.ErrorDescription != nil {
		if accessToken.Error == nil || accessToken.ErrorCodes == nil ||
			accessToken.ErrorDescription == nil {
			return accessToken, errCannotFetchAccessToken
		}
		return accessToken,
			fmt.Errorf(
				"%w: %v",
				errTokenEndpointReturnedError,
				*accessToken.ErrorDescription,
			)
	}
	if accessToken.Content == nil {
		return accessToken,
			fmt.Errorf(
				"error extracting access token: %w",
				errInsufficientFields,
			)
	}

	return accessToken, nil
}
