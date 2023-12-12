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
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// The DD51 currency specification for formatting
type CurrencySpecification struct {
	// e.g. “Japanese Yen” or "Bitcoin (Mainnet)"
	Name string

	// how many digits the user may enter after the decimal separator
	NumFractionalInputDigits uint

	// €,$,£: 2; some arabic currencies: 3, ¥: 0
	NumFractionalNormalDigits uint

	// usually same as fractionalNormalDigits, but e.g. might be 2 for ¥
	NumFractionalTrailingZeroDigits uint

	// map of powers of 10 to alternative currency names / symbols,
	// must always have an entry under "0" that defines the base name,
	// e.g.  "0 : €" or "3 : k€". For BTC, would be "0 : BTC, -3 : mBTC".
	// This way, we can also communicate the currency symbol to be used.
	AllUnitNames map[int]string
	
}

var Currencies = map[string]CurrencySpecification {
	"KUDOS": CurrencySpecification{
		Name: "KUDOS",
		NumFractionalInputDigits: 2,
		NumFractionalNormalDigits: 2,
		AllUnitNames: map[int]string{
			0: "KUDOS",
		},
	},
	"USD": CurrencySpecification{
		Name: "US Dollar",
		NumFractionalInputDigits: 2,
		NumFractionalNormalDigits: 2,
		AllUnitNames: map[int]string{
			0: "$",
		},
	},
	"EUR": CurrencySpecification{
		Name: "Euro",
		NumFractionalInputDigits: 2,
		NumFractionalNormalDigits: 2,
		AllUnitNames: map[int]string{
			0: "€",
		},
	},
	"JPY": CurrencySpecification{
		Name: "Japanese Yen",
		NumFractionalInputDigits: 2,
		NumFractionalNormalDigits: 0,
		AllUnitNames: map[int]string{
			0: "¥",
		},
	},
}

// The GNU Taler Amount object
type Amount struct {

	// The type of currency, e.g. EUR
	Currency string

	// The value (before the ".")
	Value uint64

	// The fraction (after the ".", optional)
	Fraction uint64
}

// The maximim length of a fraction (in digits)
const FractionalLength = 8

// The base of the fraction.
const FractionalBase = 1e8

// The maximum value
var MaxAmountValue = uint64(math.Pow(2, 52))

// Create a new amount from value and fraction in a currency
func NewAmount(currency string, value uint64, fraction uint64) Amount {
	return Amount{
		Currency: currency,
		Value:    value,
		Fraction: fraction,
	}
}

// FIXME also use allUnitNames.
func (a *Amount) FormatWithCurrencySpecification(cf CurrencySpecification) (string, error) {
	
	if cf.NumFractionalNormalDigits == 0 {
		return fmt.Sprintf("%s %d", cf.AllUnitNames[0], a.Value), nil
	}
	return fmt.Sprintf("%s %d.%0*d", cf.AllUnitNames[0], a.Value, cf.NumFractionalNormalDigits, a.Fraction/1e6), nil
}

func (a *Amount) Format() (string,error) {
	cf, idx := Currencies[a.Currency]
	if idx {
		return a.FormatWithCurrencySpecification(cf)
	}
	return "", errors.New("No currency specification found for " + a.Currency)
}

// Subtract the amount b from a and return the result.
// a and b must be of the same currency and a >= b
func (a *Amount) Sub(b Amount) (*Amount, error) {
	if a.Currency != b.Currency {
		return nil, errors.New("Currency mismatch!")
	}
	v := a.Value
	f := a.Fraction
	if a.Fraction < b.Fraction {
		v -= 1
		f += FractionalBase
	}
	f -= b.Fraction
	if v < b.Value {
		return nil, errors.New("Amount Overflow!")
	}
	v -= b.Value
	r := Amount{
		Currency: a.Currency,
		Value:    v,
		Fraction: f,
	}
	return &r, nil
}

// Add b to a and return the result.
// Returns an error if the currencies do not match or the addition would
// cause an overflow of the value
func (a *Amount) Add(b Amount) (*Amount, error) {
	if a.Currency != b.Currency {
		return nil, errors.New("Currency mismatch!")
	}
	v := a.Value +
		b.Value +
		uint64(math.Floor((float64(a.Fraction)+float64(b.Fraction))/FractionalBase))

	if v >= MaxAmountValue {
		return nil, errors.New(fmt.Sprintf("Amount Overflow (%d > %d)!", v, MaxAmountValue))
	}
	f := uint64((a.Fraction + b.Fraction) % FractionalBase)
	r := Amount{
		Currency: a.Currency,
		Value:    v,
		Fraction: f,
	}
	return &r, nil
}

// Parses an amount string in the format <currency>:<value>[.<fraction>]
func ParseAmount(s string) (*Amount, error) {
	re, err := regexp.Compile(`^\s*([-_*A-Za-z0-9]+):([0-9]+)\.?([0-9]+)?\s*$`)
	parsed := re.FindStringSubmatch(s)

	if nil != err {
		return nil, errors.New(fmt.Sprintf("invalid amount: %s", s))
	}
	tail := "0.0"
	if len(parsed) >= 4 {
		tail = "0." + parsed[3]
	}
	if len(tail) > FractionalLength+1 {
		return nil, errors.New("fraction too long")
	}
	value, err := strconv.ParseUint(parsed[2], 10, 64)
	if nil != err {
		return nil, errors.New(fmt.Sprintf("Unable to parse value %s", parsed[2]))
	}
	fractionF, err := strconv.ParseFloat(tail, 64)
	if nil != err {
		return nil, errors.New(fmt.Sprintf("Unable to parse fraction %s", tail))
	}
	fraction := uint64(math.Round(fractionF * FractionalBase))
	currency := parsed[1]
	a := NewAmount(currency, value, fraction)
	return &a, nil
}

// Check if this amount is zero
func (a *Amount) IsZero() bool {
	return (a.Value == 0) && (a.Fraction == 0)
}

// Returns the string representation of the amount: <currency>:<value>[.<fraction>]
// Omits trailing zeroes.
func (a *Amount) String() string {
	v := strconv.FormatUint(a.Value, 10)
	if a.Fraction != 0 {
		f := strconv.FormatUint(a.Fraction, 10)
		f = strings.TrimRight(f, "0")
		v = fmt.Sprintf("%s.%s", v, f)
	}
	return fmt.Sprintf("%s:%s", a.Currency, v)
}
