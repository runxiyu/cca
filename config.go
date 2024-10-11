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
 * it, and throw it into a normal config struct without pointers, reporting
 * missing values.
 * We should probably use reflection instead.
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
		Client    *string `scfg:"client"`
		Authorize *string `scfg:"authorize"`
		Jwks      *string `scfg:"jwks"`
		Token     *string `scfg:"token"`
		Secret    *string `scfg:"secret"`
		Expr      *int    `scfg:"expr"`
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
	defer func() {
		if retErr != nil {
			retErr = wrapError(errCannotProcessConfig, retErr)
		}
	}()

	f, err := os.Open(path)
	if err != nil {
		return wrapError(errCannotOpenConfig, err)
	}

	err = scfg.NewDecoder(bufio.NewReader(f)).Decode(&configWithPointers)
	if err != nil {
		return wrapError(errCannotDecodeConfig, err)
	}

	if configWithPointers.URL == nil {
		return fmt.Errorf("%w: url", errMissingConfigValue)
	}
	config.URL = *(configWithPointers.URL)

	if configWithPointers.Prod == nil {
		return fmt.Errorf("%w: prod", errMissingConfigValue)
	}
	config.Prod = *(configWithPointers.Prod)

	if configWithPointers.Listen.Proto == nil {
		return fmt.Errorf("%w: listen.proto", errMissingConfigValue)
	}
	config.Listen.Proto = *(configWithPointers.Listen.Proto)

	if configWithPointers.Listen.Net == nil {
		return fmt.Errorf("%w: listen.net", errMissingConfigValue)
	}
	config.Listen.Net = *(configWithPointers.Listen.Net)

	if configWithPointers.Listen.Addr == nil {
		return fmt.Errorf("%w: listen.addr", errMissingConfigValue)
	}
	config.Listen.Addr = *(configWithPointers.Listen.Addr)

	if configWithPointers.Listen.Trans == nil {
		return fmt.Errorf("%w: listen.trans", errMissingConfigValue)
	}
	config.Listen.Trans = *(configWithPointers.Listen.Trans)

	if config.Listen.Trans == "tls" {
		if configWithPointers.Listen.TLS.Cert == nil {
			return fmt.Errorf(
				"%w: listen.tls.cert",
				errMissingConfigValue,
			)
		}
		config.Listen.TLS.Cert = *(configWithPointers.Listen.TLS.Cert)

		if configWithPointers.Listen.TLS.Key == nil {
			return fmt.Errorf(
				"%w: listen.tls.key",
				errMissingConfigValue,
			)
		}
		config.Listen.TLS.Key = *(configWithPointers.Listen.TLS.Key)
	}

	if configWithPointers.DB.Type == nil {
		return fmt.Errorf("%w: db.type", errMissingConfigValue)
	}
	config.DB.Type = *(configWithPointers.DB.Type)

	if configWithPointers.DB.Conn == nil {
		return fmt.Errorf("%w: db.conn", errMissingConfigValue)
	}
	config.DB.Conn = *(configWithPointers.DB.Conn)

	if configWithPointers.Auth.Client == nil {
		return fmt.Errorf("%w: auth.client", errMissingConfigValue)
	}
	config.Auth.Client = *(configWithPointers.Auth.Client)

	if configWithPointers.Auth.Authorize == nil {
		return fmt.Errorf("%w: auth.authorize", errMissingConfigValue)
	}
	config.Auth.Authorize = *(configWithPointers.Auth.Authorize)

	if configWithPointers.Auth.Jwks == nil {
		return fmt.Errorf("%w: auth.jwks", errMissingConfigValue)
	}
	config.Auth.Jwks = *(configWithPointers.Auth.Jwks)

	if configWithPointers.Auth.Token == nil {
		return fmt.Errorf("%w: auth.token", errMissingConfigValue)
	}
	config.Auth.Token = *(configWithPointers.Auth.Token)

	if configWithPointers.Auth.Secret == nil {
		return fmt.Errorf("%w: auth.secret", errMissingConfigValue)
	}
	config.Auth.Secret = *(configWithPointers.Auth.Secret)

	if configWithPointers.Auth.Expr == nil {
		return fmt.Errorf("%w: auth.expr", errMissingConfigValue)
	}
	config.Auth.Expr = *(configWithPointers.Auth.Expr)

	if configWithPointers.Perf.SendQ == nil {
		return fmt.Errorf(
			"%w: perf.sendq",
			errMissingConfigValue,
		)
	}
	config.Perf.SendQ = *(configWithPointers.Perf.SendQ)

	if configWithPointers.Perf.MessageArgumentsCap == nil {
		return fmt.Errorf(
			"%w: perf.msg_args_cap",
			errMissingConfigValue,
		)
	}
	config.Perf.MessageArgumentsCap = *(configWithPointers.Perf.MessageArgumentsCap)

	if configWithPointers.Perf.MessageBytesCap == nil {
		return fmt.Errorf(
			"%w: perf.msg_bytes_cap",
			errMissingConfigValue,
		)
	}
	config.Perf.MessageBytesCap = *(configWithPointers.Perf.MessageBytesCap)

	if configWithPointers.Perf.ReadHeaderTimeout == nil {
		return fmt.Errorf(
			"%w: perf.read_header_timeout",
			errMissingConfigValue,
		)
	}
	config.Perf.ReadHeaderTimeout = *(configWithPointers.Perf.ReadHeaderTimeout)

	if configWithPointers.Perf.UsemDelayShiftBits == nil {
		return fmt.Errorf(
			"%w: perf.usem_delay_shift_bits",
			errMissingConfigValue,
		)
	}
	config.Perf.UsemDelayShiftBits = *(configWithPointers.Perf.UsemDelayShiftBits)

	if configWithPointers.Perf.PropagateImmediate == nil {
		return fmt.Errorf(
			"%w: perf.propagate_immediate",
			errMissingConfigValue,
		)
	}
	config.Perf.PropagateImmediate = *(configWithPointers.Perf.PropagateImmediate)

	if configWithPointers.Req.Y9.Sport == nil {
		return fmt.Errorf(
			"%w: req.y9.sport",
			errMissingConfigValue,
		)
	}
	config.Req.Y9.Sport = *(configWithPointers.Req.Y9.Sport)
	if configWithPointers.Req.Y9.NonSport == nil {
		return fmt.Errorf(
			"%w: req.y9.non_sport",
			errMissingConfigValue,
		)
	}
	config.Req.Y9.NonSport = *(configWithPointers.Req.Y9.NonSport)
	if configWithPointers.Req.Y10.Sport == nil {
		return fmt.Errorf(
			"%w: req.y10.non_sport",
			errMissingConfigValue,
		)
	}
	config.Req.Y10.Sport = *(configWithPointers.Req.Y10.Sport)
	if configWithPointers.Req.Y10.NonSport == nil {
		return fmt.Errorf(
			"%w: req.y10.non_sport",
			errMissingConfigValue,
		)
	}
	config.Req.Y10.NonSport = *(configWithPointers.Req.Y10.NonSport)
	if configWithPointers.Req.Y11.Sport == nil {
		return fmt.Errorf(
			"%w: req.y11.sport",
			errMissingConfigValue,
		)
	}
	config.Req.Y11.Sport = *(configWithPointers.Req.Y11.Sport)
	if configWithPointers.Req.Y11.NonSport == nil {
		return fmt.Errorf(
			"%w: req.y11.non_sport",
			errMissingConfigValue,
		)
	}
	config.Req.Y11.NonSport = *(configWithPointers.Req.Y11.NonSport)
	if configWithPointers.Req.Y12.Sport == nil {
		return fmt.Errorf(
			"%w: req.y12.sport",
			errMissingConfigValue,
		)
	}
	config.Req.Y12.Sport = *(configWithPointers.Req.Y12.Sport)
	if configWithPointers.Req.Y12.NonSport == nil {
		return fmt.Errorf(
			"%w: req.y12.non_sport",
			errMissingConfigValue,
		)
	}
	config.Req.Y12.NonSport = *(configWithPointers.Req.Y12.NonSport)

	return nil
}
