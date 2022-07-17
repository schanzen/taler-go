package util

import (
  "testing"
  "fmt"
)

var a = Amount{
  Currency: "EUR",
  Value: 1,
  Fraction: 50000000,
}
var b = Amount{
  Currency: "EUR",
  Value: 23,
  Fraction: 70007000,
}
var c = Amount{
  Currency: "EUR",
  Value: 25,
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
