/*
 * Handle configuration
 *
 * Copyright (c) 2024  Runxi Yu <me@runxiyu.org>
 * SPDX-License-Identifier: AGPL-3.0-or-later
 */

package main

import (
	"bufio"
	"fmt"
	"os"

	"codeberg.org/emersion/go-scfg"
)

/*
 * We use two structs. The first has all of its values as pointers, and scfg
 * unmarshals the configuration to it. Then we take each value, dereference
 * it, and throw it into a normal config struct without pointers, reporting
 * missing values.
 * We should probably use reflection instead. This is stupid.
 */

var configWithPointers struct {
	URL    *string `scfg:"url"`
	Prod   *bool   `scfg:"prod"`
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
		Client      *string            `scfg:"client"`
		Authorize   *string            `scfg:"authorize"`
		Jwks        *string            `scfg:"jwks"`
		Token       *string            `scfg:"token"`
		Expr        *int               `scfg:"expr"`
		Departments *map[string]string `scfg:"depts"`
		Udepts      *map[string]string `scfg:"udepts"`
	} `scfg:"auth"`
	Perf struct {
		SendQ               *int  `scfg:"sendq"`
		MessageArgumentsCap *int  `scfg:"msg_args_cap"`
		MessageBytesCap     *int  `scfg:"msg_bytes_cap"`
		ReadHeaderTimeout   *int  `scfg:"read_header_timeout"`
		UsemDelayShiftBits  *int  `scfg:"usem_delay_shift_bits"`
		PropagateImmediate  *bool `scfg:"propagate_immediate"`
	} `scfg:"perf"`
	Req struct {
		Y9 struct {
			Sport    *int `scfg:"sport"`
			NonSport *int `scfg:"non_sport"`
		} `scfg:"y9"`
		Y10 struct {
			Sport    *int `scfg:"sport"`
			NonSport *int `scfg:"non_sport"`
		} `scfg:"y10"`
		Y11 struct {
			Sport    *int `scfg:"sport"`
			NonSport *int `scfg:"non_sport"`
		} `scfg:"y11"`
		Y12 struct {
			Sport    *int `scfg:"sport"`
			NonSport *int `scfg:"non_sport"`
		} `scfg:"y12"`
	} `scfg:"req"`
}

var config struct {
	URL    string
	Prod   bool
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
		Client      string
		Authorize   string
		Jwks        string
		Token       string
		Expr        int
		Departments map[string]string
		Udepts      map[string]string
	}
	Perf struct {
		SendQ               int
		MessageArgumentsCap int
		MessageBytesCap     int
		ReadHeaderTimeout   int
		UsemDelayShiftBits  int
		PropagateImmediate  bool
	}
	Req struct {
		Y9 struct {
			Sport    int
			NonSport int
		}
		Y10 struct {
			Sport    int
			NonSport int
		}
		Y11 struct {
			Sport    int
			NonSport int
		}
		Y12 struct {
			Sport    int
			NonSport int
		}
	}
}

func fetchConfig(path string) (retErr error) {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open config: %w", err)
	}

	err = scfg.NewDecoder(bufio.NewReader(f)).Decode(&configWithPointers)
	if err != nil {
		return fmt.Errorf("decode config: %w", err)
	}

	if configWithPointers.URL == nil {
		return fmt.Errorf("missing config value: url")
	}
	config.URL = *(configWithPointers.URL)

	if configWithPointers.Prod == nil {
		return fmt.Errorf("missing config value: prod")
	}
	config.Prod = *(configWithPointers.Prod)

	if configWithPointers.Listen.Proto == nil {
		return fmt.Errorf("missing config value: listen.proto")
	}
	config.Listen.Proto = *(configWithPointers.Listen.Proto)

	if configWithPointers.Listen.Net == nil {
		return fmt.Errorf("missing config value: listen.net")
	}
	config.Listen.Net = *(configWithPointers.Listen.Net)

	if configWithPointers.Listen.Addr == nil {
		return fmt.Errorf("missing config value: listen.addr")
	}
	config.Listen.Addr = *(configWithPointers.Listen.Addr)

	if configWithPointers.Listen.Trans == nil {
		return fmt.Errorf("missing config value: listen.trans")
	}
	config.Listen.Trans = *(configWithPointers.Listen.Trans)

	if config.Listen.Trans == "tls" {
		if configWithPointers.Listen.TLS.Cert == nil {
			return fmt.Errorf("missing config value: listen.tls.cert")
		}
		config.Listen.TLS.Cert = *(configWithPointers.Listen.TLS.Cert)

		if configWithPointers.Listen.TLS.Key == nil {
			return fmt.Errorf("missing config value: listen.tls.key")
		}
		config.Listen.TLS.Key = *(configWithPointers.Listen.TLS.Key)
	}

	if configWithPointers.DB.Type == nil {
		return fmt.Errorf("missing config value: db.type")
	}
	config.DB.Type = *(configWithPointers.DB.Type)

	if configWithPointers.DB.Conn == nil {
		return fmt.Errorf("missing config value: db.conn")
	}
	config.DB.Conn = *(configWithPointers.DB.Conn)

	if configWithPointers.Auth.Client == nil {
		return fmt.Errorf("missing config value: auth.client")
	}
	config.Auth.Client = *(configWithPointers.Auth.Client)

	if configWithPointers.Auth.Authorize == nil {
		return fmt.Errorf("missing config value: auth.authorize")
	}
	config.Auth.Authorize = *(configWithPointers.Auth.Authorize)

	if configWithPointers.Auth.Jwks == nil {
		return fmt.Errorf("missing config value: auth.jwks")
	}
	config.Auth.Jwks = *(configWithPointers.Auth.Jwks)

	if configWithPointers.Auth.Token == nil {
		return fmt.Errorf("missing config value: auth.token")
	}
	config.Auth.Token = *(configWithPointers.Auth.Token)

	if configWithPointers.Auth.Expr == nil {
		return fmt.Errorf("missing config value: auth.expr")
	}
	config.Auth.Expr = *(configWithPointers.Auth.Expr)

	if configWithPointers.Auth.Departments == nil {
		return fmt.Errorf("missing config value: auth.depts")
	}
	config.Auth.Departments = *(configWithPointers.Auth.Departments)
	if config.Auth.Departments == nil {
		return fmt.Errorf("missing config value: auth.depts")
	}

	if configWithPointers.Auth.Udepts == nil {
		return fmt.Errorf("missing config value: auth.udepts")
	}
	config.Auth.Udepts = *(configWithPointers.Auth.Udepts)
	if config.Auth.Udepts == nil {
		return fmt.Errorf("missing config value: auth.udepts")
	}

	if configWithPointers.Perf.SendQ == nil {
		return fmt.Errorf("missing config value: perf.sendq")
	}
	config.Perf.SendQ = *(configWithPointers.Perf.SendQ)

	if configWithPointers.Perf.MessageArgumentsCap == nil {
		return fmt.Errorf("missing config value: perf.msg_args_cap")
	}
	config.Perf.MessageArgumentsCap = *(configWithPointers.Perf.MessageArgumentsCap)

	if configWithPointers.Perf.MessageBytesCap == nil {
		return fmt.Errorf("missing config value: perf.msg_bytes_cap")
	}
	config.Perf.MessageBytesCap = *(configWithPointers.Perf.MessageBytesCap)

	if configWithPointers.Perf.ReadHeaderTimeout == nil {
		return fmt.Errorf("missing config value: perf.read_header_timeout")
	}
	config.Perf.ReadHeaderTimeout = *(configWithPointers.Perf.ReadHeaderTimeout)

	if configWithPointers.Perf.UsemDelayShiftBits == nil {
		return fmt.Errorf("missing config value: perf.usem_delay_shift_bits")
	}
	config.Perf.UsemDelayShiftBits = *(configWithPointers.Perf.UsemDelayShiftBits)

	if configWithPointers.Perf.PropagateImmediate == nil {
		return fmt.Errorf("missing config value: perf.propagate_immediate")
	}
	config.Perf.PropagateImmediate = *(configWithPointers.Perf.PropagateImmediate)

	if configWithPointers.Req.Y9.Sport == nil {
		return fmt.Errorf("missing config value: req.y9.sport")
	}
	config.Req.Y9.Sport = *(configWithPointers.Req.Y9.Sport)
	if configWithPointers.Req.Y9.NonSport == nil {
		return fmt.Errorf("missing config value: req.y9.non_sport")
	}
	config.Req.Y9.NonSport = *(configWithPointers.Req.Y9.NonSport)
	if configWithPointers.Req.Y10.Sport == nil {
		return fmt.Errorf("missing config value: req.y10.sport")
	}
	config.Req.Y10.Sport = *(configWithPointers.Req.Y10.Sport)
	if configWithPointers.Req.Y10.NonSport == nil {
		return fmt.Errorf("missing config value: req.y10.non_sport")
	}
	config.Req.Y10.NonSport = *(configWithPointers.Req.Y10.NonSport)
	if configWithPointers.Req.Y11.Sport == nil {
		return fmt.Errorf("missing config value: req.y11.sport")
	}
	config.Req.Y11.Sport = *(configWithPointers.Req.Y11.Sport)
	if configWithPointers.Req.Y11.NonSport == nil {
		return fmt.Errorf("missing config value: req.y11.non_sport")
	}
	config.Req.Y11.NonSport = *(configWithPointers.Req.Y11.NonSport)
	if configWithPointers.Req.Y12.Sport == nil {
		return fmt.Errorf("missing config value: req.y12.sport")
	}
	config.Req.Y12.Sport = *(configWithPointers.Req.Y12.Sport)
	if configWithPointers.Req.Y12.NonSport == nil {
		return fmt.Errorf("missing config value: req.y12.non_sport")
	}
	config.Req.Y12.NonSport = *(configWithPointers.Req.Y12.NonSport)

	return nil
}
