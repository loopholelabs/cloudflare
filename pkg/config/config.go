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
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var (
	ErrUserIDRequired = errors.New("cloudflare user id is required")
	ErrTokenRequired  = errors.New("cloudflare token is required")
)

type Config struct {
	UserID string
	Token  string
}

func New() *Config {
	return new(Config)
}

func (c *Config) Validate() error {
	if c.UserID == "" {
		return ErrUserIDRequired
	}

	if c.Token == "" {
		return ErrTokenRequired
	}

	return nil
}

func (c *Config) RootPersistentFlags(flags *pflag.FlagSet) {
	flags.StringVar(&c.UserID, "cloudflare-user-id", "", "The cloudflare user id")
	flags.StringVar(&c.Token, "cloudflare-token", "", "The cloudflare token")
}

func (c *Config) GlobalRequiredFlags(cmd *cobra.Command) error {
	err := cmd.MarkFlagRequired("cloudflare-user-id")
	if err != nil {
		return err
	}

	err = cmd.MarkFlagRequired("cloudflare-token")
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) GenerateOptions(logName string) (*cloudflare.Options, error) {
	return &cloudflare.Options{
		LogName: logName,
		UserID:  c.UserID,
		Token:   c.Token,
	}, nil
}
