package money

import (
	"errors"
	"fmt"
)

var (
	ErrCurrencyMismatch = errors.New("currency mismatch")
	ErrInvalidCurrency  = errors.New("invalid currency code")
)

type Money struct {
	Amount   int64
	Currency string
}

func New(amount int64, currency string) Money {
	return Money{
		Amount:   amount,
		Currency: currency,
	}
}

func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{
		Amount:   m.Amount + other.Amount,
		Currency: m.Currency,
	}, nil
}

func (m Money) Subtract(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, ErrCurrencyMismatch
	}
	return Money{
		Amount:   m.Amount - other.Amount,
		Currency: m.Currency,
	}, nil
}

func (m Money) Multiply(factor int64) Money {
	return Money{
		Amount:   m.Amount * factor,
		Currency: m.Currency,
	}
}

func (m Money) IsZero() bool {
	return m.Amount == 0
}

func (m Money) IsPositive() bool {
	return m.Amount > 0
}

func (m Money) IsNegative() bool {
	return m.Amount < 0
}

func (m Money) String() string {
	return fmt.Sprintf("%d %s", m.Amount, m.Currency)
}

func ValidateCurrency(currency string) error {
	if len(currency) != 3 {
		return ErrInvalidCurrency
	}
	return nil
}
