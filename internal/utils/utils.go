package utils

import (
	"math/big"

	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/types"
)

// BigToDecimal translates a big integer, which must not be nil, into a
// decimal for SQLBoiler.
func BigToDecimal(x *big.Int) types.Decimal {
	return types.NewDecimal(new(decimal.Big).SetBigMantScale(x, 0))
}
