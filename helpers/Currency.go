package helpers

import (
	"fmt"
	"strings"
)

var symbols = map[string]string{
	"IDR": "Rp",
	"USD": "$",
	"EUR": "€",
	"JPY": "¥",
	"GBP": "£",
	"AUD": "A$",
	"SGD": "S$",
	"CNY": "¥",
	"KRW": "₩",
	"THB": "฿",
}

type FormatterMoney struct{}

func NewFormatterMoney() *FormatterMoney {
	return &FormatterMoney{}
}

func (formatter *FormatterMoney) formatIDR(amount int) string {
	n := fmt.Sprintf("%d", amount)
	var result []string
	for i, count := len(n)-1, 0; i >= 0; i, count = i-1, count+1 {
		result = append([]string{string(n[i])}, result...)
		if count%3 == 2 && i != 0 {
			result = append([]string{"."}, result...)
		}
	}
	return strings.Join(result, "")
}

func (formatter *FormatterMoney) formatIntl(amount float64) string {
	return fmt.Sprintf("%,.2f", amount)
}

func (formatter *FormatterMoney) FormaterCurrency(amount float64, currencyCode string) string {
	symbol, ok := symbols[currencyCode]
	if !ok {
		symbol = ""
	}

	switch currencyCode {
	case "IDR":
		return symbol + formatter.formatIDR(int(amount))
	default:
		return symbol + formatter.formatIntl(amount)
	}
}
