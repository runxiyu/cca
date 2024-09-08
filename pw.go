package main

import (
	"net/http"
)

/*
 * Password-handling is currently unimplemented, but a stub function is here
 * for easy implementation when that's needed.
 */
func handlePw(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		wstr(w, 405, "Only POST is supported on the password login endpoint")
		return
	}

	err := req.ParseForm()
	if err != nil {
		wstr(w, 400, "Malformed form data")
		return
	}

	username := req.PostFormValue("usernameinput")
	password := req.PostFormValue("passwordinput")

	if username == "" || password == "" {
		wstr(w, 400, "Empty username or password field")
		return
	}

	wstr(w, 401, "Authentication failed")
}
