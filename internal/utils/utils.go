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

type void struct{}

// GetSliceDiff compares two slices and returns slice of differences
func GetSliceDiff(subset, superset []string) []string {
	ma := make(map[string]void, len(subset))

	var diffs []string
	for _, ka := range subset {
		ma[ka] = void{}
	}

	for _, kb := range superset {
		if _, ok := ma[kb]; !ok {
			diffs = append(diffs, kb)
		}
	}
	return diffs
}
