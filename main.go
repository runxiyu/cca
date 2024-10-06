/*
 * Main listener
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
	"crypto/tls"
	"embed"
	"flag"
	"html/template"
	"io/fs"
	"log"
	"net"
	"net/http"
	"time"
)

var tmpl *template.Template

//go:embed build/static/* tmpl/*
//go:embed build/iadocs/*.pdf build/iadocs/*.htm build/docs/*
var runFS embed.FS

//go:embed go.* *.go
//go:embed docs/* iadocs/*
//go:embed frontend/* tmpl/*
//go:embed README.md LICENSE Makefile .editorconfig .gitignore
//go:embed scripts/* sql/*
var srcFS embed.FS

func main() {
	var err error

	var configPath string

	flag.StringVar(
		&configPath,
		"config",
		"cca.scfg",
		"path to configuration file",
	)
	flag.Parse()

	if err := fetchConfig(configPath); err != nil {
		log.Fatal(err)
	}

	log.Println("Setting up templates")
	tmpl, err = template.ParseFS(runFS, "tmpl/*")
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Registering static handle")
	staticFS, err := fs.Sub(runFS, "build/static")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(http.FS(staticFS)),
		),
	)

	log.Println("Registering iadocs handle")
	iaDocsFS, err := fs.Sub(runFS, "build/iadocs")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle("/iadocs/",
		http.StripPrefix(
			"/iadocs/",
			http.FileServer(http.FS(iaDocsFS)),
		),
	)

	log.Println("Registering docs handle")
	docsFS, err := fs.Sub(runFS, "build/docs")
	if err != nil {
		log.Fatal(err)
	}
	http.Handle(
		"/docs/",
		http.StripPrefix(
			"/docs/",
			http.FileServer(http.FS(docsFS)),
		),
	)

	log.Println("Registering source handle")
	http.Handle(
		"/src/",
		http.StripPrefix(
			"/src/", http.FileServer(http.FS(srcFS)),
		),
	)

	log.Println("Registering handlers")
	http.HandleFunc("/{$}", handleIndex)
	http.HandleFunc("/auth", handleAuth)
	http.HandleFunc("/ws", handleWs)

	var l net.Listener

	switch config.Listen.Trans {
	case "plain":
		log.Printf(
			"Establishing plain listener for net \"%s\", addr \"%s\"\n",
			config.Listen.Net,
			config.Listen.Addr,
		)
		l, err = net.Listen(config.Listen.Net, config.Listen.Addr)
		if err != nil {
			log.Fatalf(
				"Failed to establish plain listener: %v\n",
				err,
			)
		}
	case "tls":
		cer, err := tls.LoadX509KeyPair(
			config.Listen.TLS.Cert,
			config.Listen.TLS.Key,
		)
		if err != nil {
			log.Fatalf(
				"Failed to load TLS certificate and key: %v\n",
				err,
			)
		}
		tlsconfig := &tls.Config{
			Certificates: []tls.Certificate{cer},
			MinVersion:   tls.VersionTLS13,
		} //exhaustruct:ignore
		log.Printf(
			"Establishing TLS listener for net \"%s\", addr \"%s\"\n",
			config.Listen.Net,
			config.Listen.Addr,
		)
		l, err = tls.Listen(
			config.Listen.Net,
			config.Listen.Addr,
			tlsconfig,
		)
		if err != nil {
			log.Fatalf(
				"Failed to establish TLS listener: %v\n",
				err,
			)
		}
	default:
		log.Fatalln("listen.trans must be \"plain\" or \"tls\"")
	}

	log.Println("Setting up database")
	if err := setupDatabase(); err != nil {
		log.Fatal(err)
	}

	log.Println("Setting up courses")
	err = setupCourses()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Setting up JWKS")
	if err := setupJwks(); err != nil {
		log.Fatal(err)
	}

	if config.Listen.Proto == "http" {
		log.Println("Serving http")
		srv := &http.Server{
			ReadHeaderTimeout: time.Duration(
				config.Perf.ReadHeaderTimeout,
			) * time.Second,
		} //exhaustruct:ignore
		err = srv.Serve(l)
	} else {
		log.Fatalln("Unsupported protocol")
	}
	if err != nil {
		log.Fatal(err)
	}
}
