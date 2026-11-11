package metrics

import (
	"strconv"
	"strings"
)

func (a Amount) String() string {
	value := int64(a)
	if value < 0 {
		return "-" + Amount(-value).String()
	}
	whole := value / 100
	frac := value % 100
	if frac == 0 {
		return strconv.FormatInt(whole, 10)
	}
	if frac%10 == 0 {
		return strconv.FormatInt(whole, 10) + "." + strconv.FormatInt(frac/10, 10)
	}
	return strconv.FormatInt(whole, 10) + "." + strconv.FormatInt(frac+100, 10)[1:]
}

func (a Amount) MarshalJSON() ([]byte, error) {
	return []byte(a.String()), nil
}

func parseAmount(raw string) (Amount, error) {
	text := strings.TrimSpace(raw)
	if text == "" {
		return 0, errInvalidAmount
	}
	negative := false
	if strings.HasPrefix(text, "-") {
		negative = true
		text = text[1:]
	}
	parts := strings.Split(text, ".")
	if len(parts) > 2 {
		return 0, errInvalidAmount
	}
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, errInvalidAmount
	}
	frac := int64(0)
	if len(parts) == 2 {
		digits := parts[1]
		if len(digits) == 1 {
			digits += "0"
		}
		if len(digits) != 2 {
			return 0, errInvalidAmount
		}
		frac, err = strconv.ParseInt(digits, 10, 64)
		if err != nil {
			return 0, errInvalidAmount
		}
	}
	value := whole*100 + frac
	if negative {
		value = -value
	}
	return Amount(value), nil
}
