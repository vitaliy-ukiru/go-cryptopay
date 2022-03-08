package go_crypto_pay

import (
	"net/http"
	"reflect"
	"testing"
)

func getClient() *Client {
	server := ApiClientServer()
	return NewClient(ClientSettings{
		Token:      "5675:test_token",
		ApiHost:    server.URL,
		HttpClient: server.Client(),
	})
}

func TestNewClient(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		c := NewClient(ClientSettings{Token: "0"})
		if c.api.httpClient != http.DefaultClient {
			t.Error("default value for http client not equal")
		}
		if c.api.url != MainNetHost {
			t.Error("default value for api host not equal")
		}
	})

	t.Run("empty token", func(t *testing.T) {
		defer func(t *testing.T) {
			if err := recover(); err == nil {
				t.Error("invalid token filtering")
			}
		}(t)
		NewClient(ClientSettings{})
	})
}

func TestClient_Alias(t *testing.T) {
	a := getApi()
	w := Webhook{tokenHash: []byte(tokenHash)}
	c := &Client{
		api: &a,
		w:   &w,
	}
	if !reflect.DeepEqual(c.Api(), a) {
		t.Error("api instances not equals")
	}
	if !reflect.DeepEqual(c.Webhook(), w) {
		t.Error("webhook instances not equals")
	}
}

func TestClient_GetMe(t *testing.T) {
	c := getClient()
	me, err := c.GetMe()
	if err != nil {
		t.Error(err)
	}
	if me.Id != 5675 {
		t.Error("invalid id")
	}
}

func TestClient_CreateInvoice(t *testing.T) {
	c := getClient()
	t.Run("with params usage", func(t *testing.T) {
		_, err := c.CreateInvoice(BTC, 3.14, CreateInvoiceOptions{})
		if err != nil {
			t.Error(err)
		}
	})
	t.Run("with opt param usage", func(t *testing.T) {
		_, err := c.CreateInvoice("", 0, CreateInvoiceOptions{
			Asset:  ETH,
			Amount: "3.14",
		})
		if err != nil {
			t.Error(err)
		}
	})
	_, err := c.CreateInvoice("", 0, CreateInvoiceOptions{})
	if err == nil {
		t.Error()
	}
}

func TestClient_DoTransfer(t *testing.T) {
	c := getClient()
	_, err := c.DoTransfer(1, BTC, 1, "0", DoTransferOptions{})
	if err != nil {
		t.Error(err)
	}
	_, err = c.DoTransfer(1, BTC, 1, "0", DoTransferOptions{})
	if err == nil {
		t.Error("not unique spend_id")
	}
	_, err = c.DoTransfer(0, "", 0, "", DoTransferOptions{})
	if err == nil {
		t.Error("pass empty values")
	}
}

func TestClient_GetBalance(t *testing.T) {

}

func TestClient_GetExchangeRates(t *testing.T) {
}

func TestClient_GetCurrencies(t *testing.T) {

}

func TestClient_HandlerBinding(t *testing.T) {
	c := getClient()
	var handledType UpdateType
	c.On("test", func(_ *WebhookUpdate) {
		handledType = "test_1"
	})
	c.On("test", func(_ *WebhookUpdate) {
		handledType = "test_2"
	})
	c.w.handlers["test"][0](nil)
	if handledType != "test_1" {
		t.Error("handler test_1 did not work")
	}
	c.w.handlers["test"][1](nil)
	if handledType != "test_2" {
		t.Error("handler test_2 did not work")
	}
	i := c.OnInvoicePaid(func(_ *WebhookUpdate) {
		handledType = UpdateInvoicePaid
	})
	c.w.handlers[UpdateInvoicePaid][0](nil)
	if handledType != UpdateInvoicePaid {
		t.Error("handler OnInvoicePaid did not work")
	}
	c.DeleteHandler(UpdateInvoicePaid, i)
	if len(c.w.handlers[UpdateInvoicePaid]) != 0 {
		t.Error("invoice_paid not empty")
	}

	c.DeleteAllHandlersFor("test")
	if _, ok := c.w.handlers["test"]; ok {
		t.Error("handlers for test not empty")
	}
}

func TestClient_Once(t *testing.T) {
	c := getClient()
	var result bool
	c.Once("test", func(_ *WebhookUpdate) {
		result = true
	})
	c.w.handlers["test"][0](nil)
	if !result {
		t.Error("handler did not work")
	}
	if len(c.w.handlers["test"]) != 0 {
		t.Error("handlers count != 0")
	}
}

func TestBalanceInfo(t *testing.T) {
	var balance BalanceInfo = []BalanceCurrency{
		{
			CurrencyCode: "test",
			Available:    "1",
		},
		{
			CurrencyCode: BTC,
			Available:    "100",
		},
		{
			"err",
			"string",
		},
	}
	if balance.AsMap()[BTC] != "100" {
		t.Error("invalid map")
	}
	if _, err := balance.AsMapFloat(); err == nil {
		t.Error("invalid float parsing")
	}
	balance[2].Available = "3.14"
	m, err := balance.AsMapFloat()
	if err != nil {
		t.Error(err)
	}
	if m["err"] != 3.14 {
		t.Error("invalid result")
	}

}

func TestCurrenciesInfo_AsMap(t *testing.T) {
	var currencies CurrencyInfoArray = []CurrencyInfo{
		{
			IsBlockchain: true,
			Code:         BTC,
		},
		{
			IsBlockchain: false,
			Code:         "USD",
		},
	}
	if !currencies.AsMap()[BTC].IsBlockchain {
		t.Error("invalid map")
	}

}

func TestExchangeRatesInfo(t *testing.T) {
	var exchangeRates ExchangeRateArray = []ExchangeRate{
		{
			IsValid: true,
			Source:  BTC,
			Target:  "USD",
			Rate:    "40000.12",
		},
		{
			IsValid: true,
			Source:  BTC,
			Target:  "EUR",
			Rate:    "33468",
		},
		{
			IsValid: true,
			Source:  ETH,
			Target:  "USD",
			Rate:    "2604.14",
		},
		{
			IsValid: true,
			Source:  ETH,
			Target:  "EUR",
			Rate:    "2297",
		},
		{
			IsValid: false,
			Source:  ETH,
			Target:  BTC,
			Rate:    "0",
		},
	}
	if exchangeRates.AsMap()[RatesKey{ETH, "USD"}].Rate != "2604.14" {
		t.Error("invalid AsMap()")
	}
	if _, ok := exchangeRates.Get("test", "invalid"); ok {
		t.Error("invalid Get()")
	}
	if _, ok := exchangeRates.Get(ETH, BTC); ok {
		t.Error("invalid filtering Get()")
	}
	if v, ok := exchangeRates.Get(BTC, "USD"); !ok || v != "40000.12" {
		t.Errorf("invalid Get() ok=%v, v=%v", ok, v)
	}
}
