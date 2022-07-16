package taler

import (
  "testing"
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

func TestAmountString(t *testing.T) {
  if c.String() != "EUR:25.20007" {
    t.Errorf("Failed to generate correct string")
  }
}
