package domain

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidMoney = errors.New("invalid money amount")
var ErrUnsupportedCurrency = errors.New("unsupported currency")

// ParseMinorUnits parses canonical decimal text without floating point. Supported currencies use two minor digits.
func ParseMinorUnits(amount, currency string) (int64, error) {
	if currency != "USD" {
		return 0, ErrUnsupportedCurrency
	}
	if amount == "" || strings.TrimSpace(amount) != amount || strings.HasPrefix(amount, "-") || strings.HasPrefix(amount, "+") {
		return 0, ErrInvalidMoney
	}
	parts := strings.Split(amount, ".")
	if len(parts) > 2 || parts[0] == "" {
		return 0, ErrInvalidMoney
	}
	if len(parts[0]) > 1 && parts[0][0] == '0' {
		return 0, ErrInvalidMoney
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, ErrInvalidMoney
	}
	frac := ""
	if len(parts) == 2 {
		frac = parts[1]
		if len(frac) == 0 || len(frac) > 2 {
			return 0, ErrInvalidMoney
		}
	}
	for _, c := range frac {
		if c < '0' || c > '9' {
			return 0, ErrInvalidMoney
		}
	}
	if len(frac) == 1 {
		frac += "0"
	}
	if frac == "" {
		frac = "00"
	}
	cents, _ := strconv.ParseInt(frac, 10, 64)
	if whole > (int64(^uint64(0)>>1)-cents)/100 {
		return 0, ErrInvalidMoney
	}
	return whole*100 + cents, nil
}
