/*
 * Main listener
 *
 * Copyright (C) 2024  Runxi Yu <https://runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"context"
	"crypto/tls"
	"embed"
	"flag"
	"html/template"
	"io/fs"
	"log"
	"log/slog"
	"net"
	"net/http"
	"time"
)

var tmpl *template.Template

//go:embed build/static/* templates/*
//go:embed build/iadocs/*.pdf build/iadocs/*.htm build/iadocs/*.html
//go:embed build/docs/*
var runFS embed.FS

//go:embed go.* *.go
//go:embed docs/* iadocs/*
//go:embed frontend/* templates/*
//go:embed README.md LICENSE Makefile .editorconfig .gitignore .gitattributes
//go:embed scripts/* sql/*
var srcFS embed.FS

func main() {
	var err error

	var configPath string

	flag.StringVar(
		&configPath,
		"c",
		"cca.scfg",
		"path to configuration file",
	)
	flag.Parse()

	if err := fetchConfig(configPath); err != nil {
		log.Fatalln(err)
	}

	slog.Info("setting up templates")
	tmpl, err = template.ParseFS(runFS, "templates/*")
	if err != nil {
		log.Fatalln(err)
	}

	slog.Info("registering static handle")
	staticFS, err := fs.Sub(runFS, "build/static")
	if err != nil {
		log.Fatalln(err)
	}
	http.Handle("/static/",
		http.StripPrefix(
			"/static/",
			http.FileServer(http.FS(staticFS)),
		),
	)

	slog.Info("registering iadocs handle")
	iaDocsFS, err := fs.Sub(runFS, "build/iadocs")
	if err != nil {
		log.Fatalln(err)
	}
	http.Handle("/iadocs/",
		http.StripPrefix(
			"/iadocs/",
			http.FileServer(http.FS(iaDocsFS)),
		),
	)

	slog.Info("registering docs handle")
	docsFS, err := fs.Sub(runFS, "build/docs")
	if err != nil {
		log.Fatalln(err)
	}
	http.Handle(
		"/docs/",
		http.StripPrefix(
			"/docs/",
			http.FileServer(http.FS(docsFS)),
		),
	)

	slog.Info("registering source handle")
	http.Handle(
		"/src/",
		http.StripPrefix(
			"/src/", http.FileServer(http.FS(srcFS)),
		),
	)

	slog.Info("registering handlers")
	http.HandleFunc("/ws", handleWs)
	setHandler("/{$}", handleIndex)
	setHandler("/export/choices", handleExportChoices)
	setHandler("/export/students", handleExportStudents)
	setHandler("/auth", handleAuth)
	setHandler("/state", handleState)
	setHandler("/newcourses", handleNewCourses)

	var l net.Listener

	switch config.Listen.Trans {
	case "plain":
		slog.Info(
			"plain",
			"net", config.Listen.Net,
			"addr", config.Listen.Addr,
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
		slog.Info(
			"tls",
			"net", config.Listen.Net,
			"addr", config.Listen.Addr,
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

	slog.Info("setting up database")
	if err := setupDatabase(); err != nil {
		log.Fatalln(err)
	}

	slog.Info("loading state")
	if err := loadStateAndSchedule(); err != nil {
		log.Fatalln(err)
	}

	slog.Info("setting up courses")
	err = setupCourses(context.Background())
	if err != nil {
		log.Fatalln(err)
	}

	slog.Info("setting up JWKS")
	if err := setupJwks(); err != nil {
		log.Fatalln(err)
	}

	if config.Listen.Proto == "http" {
		slog.Info("serving http")
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
		log.Fatalln(err)
	}
}
