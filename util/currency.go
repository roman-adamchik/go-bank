package util

// list of supported currencies
const (
	USD = "USD"
	EUR = "EUR"
	ILS = "ILS"
)

func IsSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, ILS:
		return true
	}
	return false
}
