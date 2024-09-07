/*
 * Main listener
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
	"html/template"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
)

var tmpl *template.Template

func main() {
	var err error

	if err := fetchConfig("cca.scfg"); err != nil {
		log.Fatal(err)
	}

	log.Println("Setting up database")
	if err := setupDatabase(); err != nil {
		log.Fatal(err)
	}

	log.Println("Setting up JWKS")
	if err := setupJwks(); err != nil {
		log.Fatal(err)
	}

	log.Println("Setting up templates")
	tmpl, err = template.ParseGlob(config.Tmpl)
	if err != nil {
		log.Fatal(err)
	}

	if config.Static {
		log.Println("Registering static handle")
		fs := http.FileServer(http.Dir("./static"))
		http.Handle("/static/", http.StripPrefix("/static/", fs))
	}

	log.Println("Registering handlers")
	http.HandleFunc("/{$}", handleIndex)
	http.HandleFunc("/auth", handleAuth)
	http.HandleFunc("/ws", handleWs)

	log.Printf(
		"Establishing listener for net \"%s\", addr \"%s\"\n",
		config.Listen.Net,
		config.Listen.Addr,
	)
	l, err := net.Listen(config.Listen.Net, config.Listen.Addr)
	if err != nil {
		log.Fatal(err)
	}

	if config.Listen.Proto == "http" {
		log.Println("Serving http")
		err = http.Serve(l, nil)
	} else if config.Listen.Proto == "fcgi" {
		log.Println("Serving fcgi")
		err = fcgi.Serve(l, nil)
	}
	if err != nil {
		log.Fatal(err)
	}
}
