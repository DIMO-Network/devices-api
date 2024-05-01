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

// SliceDiff compares two slices and returns slice of differences
func SliceDiff(set, other []string) []string {
	otherSet := make(map[string]struct{}, len(other))

	for _, x := range other {
		otherSet[x] = struct{}{}
	}

	var diff []string
	for _, x := range set {
		if _, ok := otherSet[x]; !ok {
			diff = append(diff, x)
		}
	}
	return diff
}
