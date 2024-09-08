package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/jackc/pgx/v5"
)

/*
 * Serve the index page. Also handles the login page in case the user doesn't
 * have any valid login cookies.
 */
func handleIndex(w http.ResponseWriter, req *http.Request) {
	session_cookie, err := req.Cookie("session")
	if errors.Is(err, http.ErrNoCookie) {
		err = tmpl.ExecuteTemplate(
			w,
			"index_login",
			map[string]string{
				"authUrl": generate_authorization_url(),
				/*
				 * We directly generate the login URL here
				 * instead of doing so in a redirect to save
				 * requests.
				 */
			},
		)
		if err != nil {
			log.Println(err)
			return
		}
		return
	} else if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(400)
		w.Write([]byte(fmt.Sprintf(
			"Error\n" +
				"Unable to check cookie.",
		)))
		return
	}

	var userid string
	err = db.QueryRow(context.Background(), "SELECT userid FROM sessions WHERE cookie = $1", session_cookie.Value).Scan(&userid)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = tmpl.ExecuteTemplate(
				w,
				"index_login",
				map[string]interface{}{
					"authUrl": generate_authorization_url(),
					"notes":   []string{"Technically you have a session cookie, but it seems invalid."},
				},
			)
			if err != nil {
				log.Println(err)
				return
			}
			return
		} else {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf(
				"Error\nUnexpected database error.\n%s\n",
				err,
			)))
			return
		}
	}

	var name string
	var department string
	err = db.QueryRow(context.Background(), "SELECT name, department FROM users WHERE id = $1", userid).Scan(&name, &department)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf(
				"Error\nYour user doesn't exist. (This looks like a data integrity error.)\n%s\n",
				err,
			)))
			return
		} else {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf(
				"Error\nUnexpected database error.\n%s\n",
				err,
			)))
			return
		}
	}
	err = tmpl.ExecuteTemplate(
		w,
		"index",
		map[string]interface{}{
			"open": true,
			"user": map[string]interface{}{
				"Name":       name,
				"Department": department,
			},
			"courses": courses,
		},
	)
	if err != nil {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(500)
		w.Write([]byte(fmt.Sprintf(
			"Error\nUnexpected template error.\n%s\n",
			err,
		)))
	}
}
