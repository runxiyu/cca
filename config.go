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
	"os"

	"git.sr.ht/~emersion/go-scfg"
)

var configWithPointers struct {
	Url    *string `scfg:"url"`
	Prod   *bool   `scfg:"prod"`
	Tmpl   *string `scfg:"tmpl"`
	Static *bool   `scfg:"static"`
	Listen struct {
		Proto *string `scfg:"proto"`
		Net   *string `scfg:"net"`
		Addr  *string `scfg:"addr"`
	} `scfg:"listen"`
	Db struct {
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
}

var config struct {
	Url    string
	Prod   bool
	Tmpl   string
	Static bool
	Listen struct {
		Proto string
		Net   string
		Addr  string
	}
	Db struct {
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
}

func fetchConfig(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	err = scfg.NewDecoder(bufio.NewReader(f)).Decode(&configWithPointers)
	if err != nil {
		return err
	}

	/*
	 * TODO: We segfault when there are missing configuration options.
	 * There should be better ways to handle this.
	 */
	config.Url = *(configWithPointers.Url)
	config.Prod = *(configWithPointers.Prod)
	config.Tmpl = *(configWithPointers.Tmpl)
	config.Static = *(configWithPointers.Static)
	config.Listen.Proto = *(configWithPointers.Listen.Proto)
	config.Listen.Net = *(configWithPointers.Listen.Net)
	config.Listen.Addr = *(configWithPointers.Listen.Addr)
	config.Db.Type = *(configWithPointers.Db.Type)
	config.Db.Conn = *(configWithPointers.Db.Conn)
	config.Auth.Client = *(configWithPointers.Auth.Client)
	config.Auth.Authorize = *(configWithPointers.Auth.Authorize)
	config.Auth.Jwks = *(configWithPointers.Auth.Jwks)
	config.Auth.Token = *(configWithPointers.Auth.Token)
	config.Auth.Secret = *(configWithPointers.Auth.Secret)

	return nil
}
