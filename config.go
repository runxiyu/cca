/*
 * Handle configuration
 *
 * Copyright (c) 2024  Runxi Yu <me@runxiyu.org>
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
	"bufio"
	"fmt"
	"os"

	"git.sr.ht/~emersion/go-scfg"
)

/*
 * We use two structs. The first has all of its values as pointers, and scfg
 * unmarshals the configuration to it. Then we take each value, dereference
 * it, and throw it into a normal config struct without pointers.
 * This means that any missing configuration options will simply cause a
 * segmentation fault.
 */

var configWithPointers struct {
	URL    *string `scfg:"url"`
	Prod   *bool   `scfg:"prod"`
	Tmpl   *string `scfg:"tmpl"`
	Static *bool   `scfg:"static"`
	Listen struct {
		Proto *string `scfg:"proto"`
		Net   *string `scfg:"net"`
		Addr  *string `scfg:"addr"`
	} `scfg:"listen"`
	DB struct {
		Type *string `scfg:"type"`
		Conn *string `scfg:"conn"`
	} `scfg:"db"`
	Auth struct {
		Client    *string `scfg:"client"`
		Authorize *string `scfg:"authorize"`
		Jwks      *string `scfg:"jwks"`
		Token     *string `scfg:"token"`
		Secret    *string `scfg:"secret"`
	} `scfg:"auth"`
	Perf struct {
		CoursesCap          *int `scfg:"courses_cap"`
		MessageArgumentsCap *int `scfg:"msg_args_cap"`
		MessageBytesCap     *int `scfg:"msg_bytes_cap"`
		ReadHeaderTimeout   *int `scfg:"read_header_timeout"`
	} `scfg:"perf"`
}

var config struct {
	URL    string
	Prod   bool
	Tmpl   string
	Static bool
	Listen struct {
		Proto string
		Net   string
		Addr  string
	}
	DB struct {
		Type string
		Conn string
	}
	Auth struct {
		Client    string
		Authorize string
		Jwks      string
		Token     string
		Secret    string
	}
	Perf struct {
		CoursesCap          int
		MessageArgumentsCap int
		MessageBytesCap     int
		ReadHeaderTimeout   int
	} `scfg:"perf"`
}

func fetchConfig(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("error opening configuration file: %w", err)
	}

	err = scfg.NewDecoder(bufio.NewReader(f)).Decode(&configWithPointers)
	if err != nil {
		return fmt.Errorf("error decoding configuration file: %w", err)
	}

	config.URL = *(configWithPointers.URL)
	config.Prod = *(configWithPointers.Prod)
	config.Tmpl = *(configWithPointers.Tmpl)
	config.Static = *(configWithPointers.Static)
	config.Listen.Proto = *(configWithPointers.Listen.Proto)
	config.Listen.Net = *(configWithPointers.Listen.Net)
	config.Listen.Addr = *(configWithPointers.Listen.Addr)
	config.DB.Type = *(configWithPointers.DB.Type)
	config.DB.Conn = *(configWithPointers.DB.Conn)
	config.Auth.Client = *(configWithPointers.Auth.Client)
	config.Auth.Authorize = *(configWithPointers.Auth.Authorize)
	config.Auth.Jwks = *(configWithPointers.Auth.Jwks)
	config.Auth.Token = *(configWithPointers.Auth.Token)
	config.Auth.Secret = *(configWithPointers.Auth.Secret)
	config.Perf.CoursesCap = *(configWithPointers.Perf.CoursesCap)
	config.Perf.MessageArgumentsCap = *(configWithPointers.Perf.MessageArgumentsCap)
	config.Perf.MessageBytesCap = *(configWithPointers.Perf.MessageBytesCap)
	config.Perf.ReadHeaderTimeout = *(configWithPointers.Perf.ReadHeaderTimeout)

	return nil
}
