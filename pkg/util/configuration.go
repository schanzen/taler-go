// This file is part of taler-go, the Taler Go implementation.
// Copyright (C) 2022 Martin Schanzenbach
//
// Taler Go is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published
// by the Free Software Foundation, either version 3 of the License,
// or (at your option) any later version.
//
// Taler Go is distributed in the hope that it will be useful, but
// WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
// Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.
//
// SPDX-License-Identifier: AGPL3.0-or-later

package util

import (
	"errors"
	"os"
	"strings"
	"time"

	"gopkg.in/ini.v1"
)

type TalerConfiguration struct {
	// The ini file
	Ini *ini.File
}

func (c *TalerConfiguration) GetString(section string, key string, fallback string) string {
	value := c.Ini.Section(section).Key(strings.ToUpper(key)).String()
	if len(value) == 0 {
		return c.Ini.Section(section).Key(strings.ToLower(key)).MustString(fallback)
	}
	if len(value) == 0 {
		return fallback
	}
	return value
}

func (c *TalerConfiguration) GetDuration(section string, key string, fallback time.Duration) time.Duration {
	value := c.GetString(section, key, "")
	d, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return d
}


func (c *TalerConfiguration) GetBool(section string, key string, fallback bool) bool {
	value, err := c.Ini.Section(section).Key(strings.ToUpper(key)).Bool()
	if err != nil {
		return c.Ini.Section(section).Key(strings.ToLower(key)).MustBool(fallback)
	}
	return value
}


func (c *TalerConfiguration) GetInt64(section string, key string, fallback int64) int64 {
	value, err := c.Ini.Section(section).Key(strings.ToUpper(key)).Int64()
	if err != nil {
		return c.Ini.Section(section).Key(strings.ToLower(key)).MustInt64(fallback)
	}
	return value
}

func (c *TalerConfiguration) GetInt(section string, key string, fallback int) int {
	value, err := c.Ini.Section(section).Key(strings.ToUpper(key)).Int()
	if err != nil {
		return c.Ini.Section(section).Key(strings.ToLower(key)).MustInt(fallback)
	}
	return value
}

func getFileName(prefixPath string, relativeFileName string) string {
	_, err := os.Stat(relativeFileName)
	if errors.Is(err, os.ErrNotExist) {
		_, err := os.Stat(prefixPath + "/" + relativeFileName)
		if errors.Is(err, os.ErrNotExist) {
			return ""
		}
		return prefixPath + "/" + relativeFileName
	}
	return relativeFileName
}

func (c *TalerConfiguration) GetFilename(section string, key string, fallback string, prefixPath string) string {
	value := c.GetString(section, key, fallback)
	return getFileName(prefixPath, value)
}

func LoadConfiguration(file string) (TalerConfiguration, error) {
	cfgIni, err := ini.LooseLoad(file)
	if err != nil {
		return TalerConfiguration{}, err
	}
	return TalerConfiguration{Ini: cfgIni}, nil
}
