/*
 * Handle configuration
 *
 * Copyright (c) 2024  Runxi Yu <me@runxiyu.org>
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
	Source *string `scfg:"source"`
	Listen struct {
		Proto *string `scfg:"proto"`
		Net   *string `scfg:"net"`
		Addr  *string `scfg:"addr"`
		Trans *string `scfg:"trans"`
		TLS   struct {
			Cert *string `scfg:"cert"`
			Key  *string `scfg:"key"`
		} `scfg:"tls"`
	} `scfg:"listen"`
	DB struct {
		Type *string `scfg:"type"`
		Conn *string `scfg:"conn"`
	} `scfg:"db"`
	Auth struct {
		Fake      *int    `scfg:"fake"`
		Client    *string `scfg:"client"`
		Authorize *string `scfg:"authorize"`
		Jwks      *string `scfg:"jwks"`
		Token     *string `scfg:"token"`
		Secret    *string `scfg:"secret"`
		Expr      *int    `scfg:"expr"`
	} `scfg:"auth"`
	Perf struct {
		SendQ               *int `scfg:"sendq"`
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
	Source string
	Listen struct {
		Proto string
		Net   string
		Addr  string
		Trans string
		TLS   struct {
			Cert string
			Key  string
		}
	}
	DB struct {
		Type string
		Conn string
	}
	Auth struct {
		Fake      int
		Client    string
		Authorize string
		Jwks      string
		Token     string
		Secret    string
		Expr      int
	}
	Perf struct {
		SendQ               int
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
	config.Source = *(configWithPointers.Source)
	config.Listen.Proto = *(configWithPointers.Listen.Proto)
	config.Listen.Net = *(configWithPointers.Listen.Net)
	config.Listen.Addr = *(configWithPointers.Listen.Addr)
	config.Listen.Trans = *(configWithPointers.Listen.Trans)
	if config.Listen.Trans == "tls" {
		config.Listen.TLS.Cert = *(configWithPointers.Listen.TLS.Cert)
		config.Listen.TLS.Key = *(configWithPointers.Listen.TLS.Key)
	}
	config.DB.Type = *(configWithPointers.DB.Type)
	config.DB.Conn = *(configWithPointers.DB.Conn)
	if configWithPointers.Auth.Fake == nil {
		config.Auth.Fake = 0
	} else {
		config.Auth.Fake = *(configWithPointers.Auth.Fake)
		switch config.Auth.Fake {
		case 0, 4712, 9080: /* Don't use them unless you know what you're doing */
			if config.Prod {
				panic("auth.fake not allowed in production mode")
			}
		default:
			panic("illegal auth.fake config option")
		}
	}
	config.Auth.Client = *(configWithPointers.Auth.Client)
	config.Auth.Authorize = *(configWithPointers.Auth.Authorize)
	config.Auth.Jwks = *(configWithPointers.Auth.Jwks)
	config.Auth.Token = *(configWithPointers.Auth.Token)
	config.Auth.Secret = *(configWithPointers.Auth.Secret)
	config.Auth.Expr = *(configWithPointers.Auth.Expr)
	config.Perf.SendQ = *(configWithPointers.Perf.SendQ)
	config.Perf.MessageArgumentsCap = *(configWithPointers.Perf.MessageArgumentsCap)
	config.Perf.MessageBytesCap = *(configWithPointers.Perf.MessageBytesCap)
	config.Perf.ReadHeaderTimeout = *(configWithPointers.Perf.ReadHeaderTimeout)

	return nil
}
