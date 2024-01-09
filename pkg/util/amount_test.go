package util

import (
	"fmt"
	"testing"
)

var a = Amount{
	Currency: "EUR",
	Value:    1,
	Fraction: 50000000,
}
var b = Amount{
	Currency: "EUR",
	Value:    23,
	Fraction: 70007000,
}
var c = Amount{
	Currency: "EUR",
	Value:    25,
	Fraction: 20007000,
}

func TestAmountAdd(t *testing.T) {
	d, err := a.Add(b)
	if err != nil {
		t.Errorf("Failed adding amount")
	}
	if c.String() != d.String() {
		t.Errorf("Failed to add to correct amount")
	}
}

func TestAmountSub(t *testing.T) {
	d, err := c.Sub(b)
	if err != nil {
		t.Errorf("Failed substracting amount")
	}
	if a.String() != d.String() {
		t.Errorf("Failed to substract to correct amount")
	}
}

func TestAmountLarge(t *testing.T) {
	x, err := ParseAmount("EUR:50")
	_, err = x.Add(a)
	if nil != err {
		fmt.Println(err)
		t.Errorf("Failed")
	}
}

func TestAmountFormat(t *testing.T) {
	var currencySpec = CurrencySpecification{
		Name:                      "KUDOS",
		NumFractionalNormalDigits: 2,
		AltUnitNames: map[int]string{
			0: "K",
		},
	}
	x, _ := ParseAmount("KUDOS:50.2")
	str, _ := x.FormatWithCurrencySpecification(currencySpec)
	if str != "K 50.20" {
		fmt.Println(str)
		t.Errorf("Failed")
	}
	y, _ := ParseAmount("KUDOS:50.234")
	str, _ = y.Format()
	if str != "KUDOS 50.23" {
		fmt.Println(str)
		t.Errorf("Failed")
	}
	y, _ = ParseAmount("EUR:50.234")
	str, _ = y.Format()
	if str != "€ 50.23" {
		fmt.Println(str)
		t.Errorf("Failed")
	}
	y, _ = ParseAmount("FOO:50.234")
	_, err := y.Format()
	if nil == err {
		t.Errorf("Failed")
	}
}
