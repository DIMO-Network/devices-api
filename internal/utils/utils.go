package utils

import (
	"math/big"
	"slices"

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

	diffs := make([]string, 0)
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

func ExistsInSlice(needle string, haystack []string) bool {
	return slices.IndexFunc(haystack, func(cmd string) bool {
		return cmd == needle
	}) != -1
}
