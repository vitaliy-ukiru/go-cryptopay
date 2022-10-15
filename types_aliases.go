package cryptopay

import (
	"math/big"
)

type (
	// ExchangeRates alias for slice of ExchangeRate.
	ExchangeRates []ExchangeRate

	// BalanceInfo alias for slice of BalanceCurrency.
	BalanceInfo []BalanceCurrency

	// CurrencyInfoArray alias for slice of CurrencyInfo
	CurrencyInfoArray []CurrencyInfo
)

// AsMap returns transformed BalanceInfo ([]BalanceCurrency) into map,
// key - currency code (Asset), value - balance for Asset as string
func (b BalanceInfo) AsMap() map[Asset]*big.Float {
	balances := make(map[Asset]*big.Float)
	for _, currency := range b {
		balances[currency.CurrencyCode] = currency.Available
	}
	return balances
}

// AsMapFloat returns transformed BalanceInfo ([]BalanceCurrency) into map,
// key - currency code (Asset), value - balance for Asset as float64
func (b BalanceInfo) AsMapFloat() (map[Asset]*big.Float, error) {
	balances := make(map[Asset]*big.Float)
	for _, currency := range b {
		balances[currency.CurrencyCode] = currency.Available
	}
	return balances, nil
}

// AsMap returns transformed CurrencyInfoArray ([]CurrencyInfo) into map.
func (c CurrencyInfoArray) AsMap() map[Asset]CurrencyInfo {
	currencies := make(map[Asset]CurrencyInfo)
	for _, currency := range c {
		currencies[currency.Code] = currency
	}
	return currencies
}

// RatesKey is two-value key for ExchangeRates.
type RatesKey struct {
	Source Asset
	Target Asset
}

// AsMap returns transformed ExchangeRates ([]ExchangeRate) into map.
func (e ExchangeRates) AsMap() map[RatesKey]ExchangeRate {
	rates := make(map[RatesKey]ExchangeRate)
	for _, exchangeRate := range e {
		key := RatesKey{exchangeRate.Source, exchangeRate.Target}
		rates[key] = exchangeRate
	}
	return rates
}

// Get returns exchange rate of target currency in source currency and the success indicator.
func (e ExchangeRates) Get(source, target Asset) (*big.Float, bool) {
	rate, ok := e.AsMap()[RatesKey{source, target}]
	if !ok || !rate.IsValid {
		return nil, false
	}
	return rate.Rate, true
}
