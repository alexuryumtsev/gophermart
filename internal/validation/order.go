package validation

import (
	"strconv"
)

// ValidateOrderNumber проверяет номер заказа на соответствие алгоритму Луна
func ValidateOrderNumber(number string) bool {
	// Проверка, что номер заказа состоит только из цифр
	if _, err := strconv.ParseInt(number, 10, 64); err != nil {
		return false
	}

	// Проверка по алгоритму Луна
	var sum int
	var alternate bool

	for i := len(number) - 1; i >= 0; i-- {
		n, _ := strconv.Atoi(string(number[i]))
		if alternate {
			n *= 2
			if n > 9 {
				n = (n % 10) + 1
			}
		}
		sum += n
		alternate = !alternate
	}

	return sum%10 == 0
}
