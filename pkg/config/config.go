/*
	Copyright 2023 Loophole Labs

	Licensed under the Apache License, Version 2.0 (the "License");
	you may not use this file except in compliance with the License.
	You may obtain a copy of the License at

		   http://www.apache.org/licenses/LICENSE-2.0

	Unless required by applicable law or agreed to in writing, software
	distributed under the License is distributed on an "AS IS" BASIS,
	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
	See the License for the specific language governing permissions and
	limitations under the License.
*/

package config

import (
	"errors"
	"github.com/loopholelabs/cloudflare"
	"github.com/spf13/pflag"
)

var (
	ErrUserIDRequired             = errors.New("cloudflare user id is required")
	ErrTokenRequired              = errors.New("cloudflare token is required")
	ErrPrefixRequired             = errors.New("cloudflare prefix is required")
	ErrUpstreamRootDomainRequired = errors.New("cloudflare upstream root domain is required")
)

const (
	DefaultDisabled = false
)

type Config struct {
	Disabled           bool   `mapstructure:"disabled"`
	UserID             string `mapstructure:"user_id"`
	Token              string `mapstructure:"token"`
	Prefix             string `mapstructure:"prefix"`
	UpstreamRootDomain string `mapstructure:"upstream_root_domain"`
}

func New() *Config {
	return &Config{
		Disabled: DefaultDisabled,
	}
}

func (c *Config) Validate() error {
	if !c.Disabled {
		if c.UserID == "" {
			return ErrUserIDRequired
		}

		if c.Token == "" {
			return ErrTokenRequired
		}

		if c.Prefix == "" {
			return ErrPrefixRequired
		}

		if c.UpstreamRootDomain == "" {
			return ErrUpstreamRootDomainRequired
		}
	}

	return nil
}

func (c *Config) RootPersistentFlags(flags *pflag.FlagSet) {
	flags.BoolVar(&c.Disabled, "cloudflare-disabled", DefaultDisabled, "Disable cloudflare")
	flags.StringVar(&c.UserID, "cloudflare-user-id", "", "The cloudflare user id")
	flags.StringVar(&c.Token, "cloudflare-token", "", "The cloudflare token")
	flags.StringVar(&c.Prefix, "cloudflare-prefix", "", "The cloudflare resource prefix")
	flags.StringVar(&c.UpstreamRootDomain, "cloudflare-upstream-root-domain", "", "The cloudflare upstream root domain")
}

func (c *Config) GenerateOptions(logName string) (*cloudflare.Options, error) {
	return &cloudflare.Options{
		LogName:  logName,
		Disabled: c.Disabled,
		UserID:   c.UserID,
		Token:    c.Token,
		Prefix:   c.Prefix,
	}, nil
}
