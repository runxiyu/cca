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
	"log"
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
		Fake      *int    `scfg:"fake"`
		Client    *string `scfg:"client"`
		Authorize *string `scfg:"authorize"`
		Jwks      *string `scfg:"jwks"`
		Token     *string `scfg:"token"`
		Secret    *string `scfg:"secret"`
		Expr      *int    `scfg:"expr"`
	} `scfg:"auth"`
	Perf struct {
		MessageArgumentsCap *int  `scfg:"msg_args_cap"`
		MessageBytesCap     *int  `scfg:"msg_bytes_cap"`
		ReadHeaderTimeout   *int  `scfg:"read_header_timeout"`
		UsemDelayShiftBits  *int  `scfg:"usem_delay_shift_bits"`
		PropagateImmediate  *bool `scfg:"propagate_immediate"`
	} `scfg:"perf"`
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
		Fake      int
		Client    string
		Authorize string
		Jwks      string
		Token     string
		Secret    string
		Expr      int
	}
	Perf struct {
		MessageArgumentsCap int
		MessageBytesCap     int
		ReadHeaderTimeout   int
		UsemDelayShiftBits  int
		PropagateImmediate  bool
	} `scfg:"perf"`
}

func fetchConfig(path string) (retErr error) {
	defer func() {
		if v := recover(); v != nil {
			s, ok := v.(error)
			if ok {
				retErr = fmt.Errorf(
					"%w: %w",
					errCannotProcessConfig,
					s,
				)
			}
			retErr = fmt.Errorf("%w: %v", errCannotProcessConfig, v)
			return
		}
		if retErr != nil {
			retErr = fmt.Errorf(
				"%w: %w",
				errCannotProcessConfig,
				retErr,
			)
			return
		}
	}()

	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("%w: %w", errCannotOpenConfig, err)
	}

	err = scfg.NewDecoder(bufio.NewReader(f)).Decode(&configWithPointers)
	if err != nil {
		return fmt.Errorf("%w: %w", errCannotDecodeConfig, err)
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

	if configWithPointers.Auth.Fake == nil {
		config.Auth.Fake = 0
	} else {
		config.Auth.Fake = *(configWithPointers.Auth.Fake)
		switch config.Auth.Fake {
		case 0:
			/* It's okay to set it to 0 in production */
		case 4712, 9080: /* Don't use them unless you know what you're doing */
			if config.Prod {
				return fmt.Errorf(
					"%w: fake authentication is incompatible with production mode",
					errIllegalConfig,
				)
			}
			log.Println(
				"!!! WARNING: Fake authentication is enabled. Any WebSocket connection would have a fake account. This is a HUGE security hole. You should only use this while benchmarking.",
			)
		default:
			return fmt.Errorf(
				"%w: invalid option for auth.fake",
				errIllegalConfig,
			)
		}
	}

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

	return nil
}
