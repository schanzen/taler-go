package util

import(
  "errors"
  "regexp"
  "strconv"
  "fmt"
  "math"
  "strings"
)

// The GNU Taler Amount ob
type Amount struct {

  // The type of currency, e.g. EUR
  Currency string

  // The value (before the ".")
  Value uint64

  // The fraction (after the ".", optional)
  Fraction uint64
}

const FractionalLength = 8

const FractionalBase = 1e8

const MaxAmountValue = 2^52

func NewAmount(currency string, value uint64, fraction uint64) Amount {
  return Amount{
    Currency: currency,
    Value: value,
    Fraction: fraction,
  }
}

func (a *Amount) Sub(b Amount) (*Amount,error) {
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
    Value: v,
    Fraction: f,
  }
  return &r, nil
}

func (a *Amount) Add(b Amount) (*Amount,error) {
  if a.Currency != b.Currency {
    return nil, errors.New("Currency mismatch!")
  }
  v := a.Value +
  b.Value +
  uint64(math.Floor((float64(a.Fraction) + float64(b.Fraction)) / FractionalBase))

  if v >= MaxAmountValue {
    return nil, errors.New("Amount Overflow!")
  }
  f := uint64((a.Fraction + b.Fraction) % FractionalBase)
  r := Amount{
    Currency: a.Currency,
    Value: v,
    Fraction: f,
  }
  return &r, nil
}
func ParseAmount(s string) (*Amount,error) {
  re, err := regexp.Compile(`^\s*([-_*A-Za-z0-9]+):([0-9]+)\.?([0-9]+)?\s*$`)
  parsed := re.FindStringSubmatch(s)

  if nil != err {
    return nil, errors.New(fmt.Sprintf("invalid amount: %s", s))
  }
  tail := "0.0"
  if len(parsed) >= 4 {
    tail = "0." + parsed[3]
  }
  if len(tail) > FractionalLength + 1 {
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

func (a *Amount) String() string {
  v := strconv.FormatUint(a.Value, 10)
  if a.Fraction != 0 {
    f := strconv.FormatUint(a.Fraction, 10)
    f = strings.TrimRight(f, "0")
    v = fmt.Sprintf("%s.%s", v, f)
  }
  return fmt.Sprintf("%s:%s", a.Currency, v)
}
