package model

import (
	"database/sql/driver"
	"math/big"
	"open-indexer/utils/decimal"
)

type DDecimal struct {
	value *decimal.Decimal
}

func NewDecimal() *DDecimal {
	return &DDecimal{decimal.New()}
}

func NewDecimalFromHexString(hex string) (*DDecimal, bool) {
	if hex[0:2] == "0x" || hex[0:2] == "0X" {
		hex = hex[2:]
	}

	b := new(big.Int)
	b, ok := b.SetString(hex, 16)
	if !ok {
		return &DDecimal{decimal.New()}, false
	}

	return &DDecimal{decimal.NewFromValue(b)}, true
}

func NewDecimalFromString(s string) (*DDecimal, int, error) {
	d, p, e := decimal.NewFromString(s)

	return &DDecimal{d}, p, e
}

func NewDecimalFromStringValue(s string) *DDecimal {
	d, _, _ := decimal.NewFromString(s)

	return &DDecimal{d}
}

func (dd *DDecimal) Add(other *DDecimal) *DDecimal {
	d := dd.value.Add(other.value)
	return &DDecimal{d}
}

func (dd *DDecimal) Sub(other *DDecimal) *DDecimal {
	d := dd.value.Sub(other.value)
	return &DDecimal{d}
}

func (dd *DDecimal) Cmp(other *DDecimal) int {
	return dd.value.Cmp(other.value)
}

func (dd *DDecimal) Sign() int {
	return dd.value.Sign()
}

func (dd *DDecimal) String() string {
	return dd.value.String()
}

func (dd *DDecimal) Scan(value interface{}) error {
	str := string(value.([]byte))
	d, _, err := decimal.NewFromString(str)
	dd.value = d
	return err
}

func (dd *DDecimal) Value() (driver.Value, error) {
	if dd == nil {
		return "0", nil
	}
	return dd.value.String(), nil
}
