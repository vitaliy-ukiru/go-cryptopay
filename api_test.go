package go_crypto_pay

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	apiServerInstance *httptest.Server
	onceApiServer     sync.Once
)

func getApi() ApiCore {
	s := ApiClientServer()
	return ApiCore{
		token:      "5675:test_token",
		url:        s.URL,
		httpClient: s.Client(),
	}
}

func TestApiCore_urlFmt(t *testing.T) {
	api := ApiCore{}
	var cases = []struct{ name, query, expected string }{
		{
			name:     "empty query",
			query:    emptyQuery,
			expected: "/api/test",
		},
		{
			name:     "regular query",
			query:    "count=none&offset=1",
			expected: "/api/test?count=none&offset=1",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := api.urlFmt("test", tc.query)
			if got != tc.expected {
				t.Errorf(`expected "%s", but got "%s"`, tc.expected, got)
			}
		})
	}
}

func TestApiCore_apiCall(t *testing.T) {
	if getApi().apiCall("%^&escape test", emptyQuery, nil) == nil {
		t.Error("http.NewRequest pass invalid URL escape")
	}
}

func TestApiCore_GetMe(t *testing.T) {
	me, err := getApi().GetMe()
	if err != nil {
		t.Errorf("err(%s) != nil", err)
	}
	if !me.Ok || me.Result.Id != 5675 {
		t.Errorf("me.Ok(%v) != true || me.Result.Id(%d) != 5657", me.Ok, me.Result.Id)
	}
}

func TestApiCore_CreateInvoice(t *testing.T) {
		inv, err := getApi().CreateInvoice(CreateInvoiceOptions{
		Asset:     TON,
		Amount:    "3.14",
		ExpiresIn: 1,
	})
	if err != nil {
		t.Error(err)
	}
	if inv.Result.Amount != "3.14" {
		t.Errorf("amount(%s) != 3.14", inv.Result.Amount)
	}
}

func TestApiCore_DoTransfer(t *testing.T) {
	api := getApi()
	t.Run("with error", func(t *testing.T) {
		//api.DoTransfer(DoTransferOptions{
		//	SpendId:                 "#400",
		//	Comment:                 "",
		//	DisableSendNotification: false,
		//})
	})
	t.Run("correct", func(t *testing.T) {
		_, err := api.DoTransfer(DoTransferOptions{
			Asset:   BTC,
			Amount:  "4.4",
			SpendId: "random?",
		})
		if err != nil {
			t.Error(err)
		}
	})
}

func TestApiCore_GetInvoices(t *testing.T) {
	api := getApi()
	t.Run("empty params", func(t *testing.T) {
		invoices, err := api.GetInvoices(nil)
		if err != nil {
			t.Error(err)
		}
		if len(invoices.Result.Items) != 4 {
			t.Error("count != 4")
		}
	})
	t.Run("with params", func(t *testing.T) {
		t.Run("asset", func(t *testing.T) {
			r, _ := api.GetInvoices(&GetInvoicesOptions{
				Asset: "BUSD",
			})
			if len(r.Result.Items) != 0 {
				t.Errorf("count(%d) != 0", len(r.Result.Items))
			}
		})
		t.Run("asset + status", func(t *testing.T) {
			r, _ := api.GetInvoices(&GetInvoicesOptions{
				Asset:  "BTC",
				Status: StatusActive,
				Count:  5,
			})
			if len(r.Result.Items) != 2 {
				t.Errorf("count(%d) != 2", len(r.Result.Items))
			}
		})
		t.Run("invoice id", func(t *testing.T) {
			r, _ := api.GetInvoices(&GetInvoicesOptions{
				InvoiceIds: []string{"3"},
			})
			if r.Result.Items[0].Id != 3 {
				t.Errorf("invoice_id(%d) != 3", r.Result.Items[0].Id)
			}
		})

	})
}

func TestApiCore_GetBalance(t *testing.T) {
	_, err := getApi().GetBalance()
	if err != nil {
		t.Error(err)
	}
}

func TestApiCore_GetExchangeRates(t *testing.T) {
	_, err := getApi().GetExchangeRates()
	if err != nil {
		t.Error(err)
	}
}

func TestApiCore_GetCurrencies(t *testing.T) {
	//api := getApi()
	_, err := getApi().GetCurrencies()
	if err != nil {
		t.Error(err)
	}
}



func TestEmptyToken(t *testing.T) {
	api := getApi()
	api.token = ""
	app, err := api.GetMe()
	if err != nil {
		t.Error(err)
	}
	if app.IsSuccessfully() {
		t.Error("app.IsSuccessfully() == true")
	}
	if app.Error == nil {
		t.Error("app.Error == nil")
	}
	if app.Error.Code != 401 {
		t.Errorf("error code(%d) != StatusUnauthorized", app.Error.Code)
	}

	if app.Result != nil {
		t.Errorf("app.Result(%#v) != nil", app.Result)
	}
}

/*
	Service region for local api test
*/

type JSON map[string]interface{}

var (
	UnauthorizedError = []byte(`{"ok": false, "error": {"code": 401, "name": ""}}`)
	apiErrorF         = `{"ok":false,"error":{"code":%d,"name":"%s"}}`
)

func writeJson(rw http.ResponseWriter, code int, v interface{}, args ...interface{}) {
	// fix for superfluous response.WriteHeader call"
	if code != 200 {
		rw.WriteHeader(code)
	}
	if s, ok := v.(string); ok {
		fmt.Fprint(rw, s, args)
	}
	data, err := json.Marshal(v)
	if err != nil {
		fmt.Fprintf(rw, apiErrorF, 500, err)
	}
	rw.Write(data)
}



func ApiClientServer() *httptest.Server {
	usedSpendIds := make(map[string]bool)
	onceApiServer.Do(func() {
		apiServerInstance = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			token := strings.Split(r.Header.Get(headerTokenName), ":")
			if len(token) != 2 {
				rw.Write(UnauthorizedError)
				return
			}
			appId, err := strconv.Atoi(token[0])
			if err != nil {
				rw.Write(UnauthorizedError)
				return
			}
			switch r.URL.Path {
			case "/api/getMe":
				writeJson(rw, 200, JSON{
					"ok": true,
					"result": JSON{
						"app_id": appId,
						"name": "",
						"payment_processing_bot_username": "",
					},
				})
			case "/api/createInvoice":
				values := r.URL.Query()
				if values.Get("asset") == "" {
					writeJson(rw, 400, fmt.Sprintf(apiErrorF, 400, "invalid asset"))
				}
				writeJson(rw, 200, JSON{
					"ok": true,
					"result": JSON{
						"invoice_id":       rand.Int(),
						"status":           "paid",
						"hash":             "exc10sld",
						"asset":            values.Get("asset"),
						"amount":           values.Get("amount"),
						"description":      values.Get("description"),
						"payload":          values.Get("payload"),
						"paid_btn_name":    values.Get("paid_btn_name"),
						"paid_btn_url":     values.Get("paid_btn_url"),
						"hidden_message":   values.Get("hidden_message"),
						"allow_comments":   true,
						"allow_anonymous":  true,
						"pay_url":          "/exc10sld",
						"created_at":       time.Now(),
						"paid_at":          time.Now(),
						"paid_anonymously": true,
						"expiration_date":  "",
					},
				})
			case "/api/transfer":
				values := r.URL.Query()
				if values.Get("user_id") == "" {
					writeJson(rw, 400, fmt.Sprintf(apiErrorF, 400, "invalid user id"))
					return
				}

				if values.Get("asset") == "" {
					writeJson(rw, 400, fmt.Sprintf(apiErrorF, 400, "invalid asset"))
					return
				}

				if values.Get("amount") == "" {
					writeJson(rw, 400, fmt.Sprintf(apiErrorF, 400, "invalid amount"))
					return
				}
				spendId := values.Get("spend_id")
				if spendId == "" {
					writeJson(rw, 400, fmt.Sprintf(apiErrorF, 400, "invalid spend_id"))
					return
				}
				if usedSpendIds[spendId] {
					writeJson(rw, 400, fmt.Sprintf(apiErrorF, 400, "not unique spend_id"))
					return
				} else {
					usedSpendIds[spendId] = true
				}

				writeJson(rw, 200, JSON{
					"ok": true, "result": JSON{
						"transfer_id":  rand.Int(),
						"user_id":      0,
						"asset":        values.Get("asset"),
						"amount":       values.Get("amount"),
						"status":       "completed",
						"completed_at": time.Now(),
						"comment":      values.Get("comment"),
					},
				})
			case "/api/getInvoices":
				invoices := []JSON{
					{
						"invoice_id": 0,
						"status":     "paid",
						"asset":      "BTC",
						"amount":     "1",
						"created_at": time.Now(),
					},
					{
						"invoice_id": 1,
						"status":     "active",
						"asset":      "BTC",
						"amount":     "2",
						"created_at": time.Now().Add(-time.Minute),
					},
					{
						"invoice_id": 2,
						"status":     "active",
						"asset":      "USDT",
						"amount":     "3",
						"created_at": time.Now().Add(-time.Hour),
					},
					{
						"invoice_id": 3,
						"status":     "active",
						"asset":      "BTC",
						"amount":     "2",
						"created_at": time.Now().Add(-time.Minute),
					},
				}
				if r.URL.RawQuery == "" {
					writeJson(rw, 200, JSON{
						"ok": true,
						"result": JSON{
							"items": invoices,
						},
					})
					return
				}
				values := r.URL.Query()
				filter := func(key string, source []JSON) []JSON {
					var result []JSON

					if filter := values.Get(key); filter != "" {
						for _, v := range source {
							if v[key] == filter {
								result = append(result, v)
							}
						}
					} else {
						result = source
					}
					return result
				}
				resultFilters := filter("status", filter("asset", invoices))
				var resultIds []JSON

				if strIds := values.Get("invoices_ids"); strIds != "" {
					if invoicesIds := strings.Split(strIds, ","); len(invoicesIds) > 0 {
						ids := make(map[string]struct{})
						for _, id := range invoicesIds {
							ids[id] = struct{}{}
						}

						for _, v := range resultFilters {
							field := strconv.Itoa(v["invoice_id"].(int))
							if _, ok := ids[field]; ok {
								resultIds = append(resultIds, v)
							}
						}
					}

				} else {
					resultIds = resultFilters
				}
				writeJson(rw, 200, JSON{
					"ok": true,
					"result": JSON{
						"items": resultIds,
					},
				})
			case "/api/getBalance":
				writeJson(rw, 200, JSON{
					"ok": true,
					"result": []JSON{
						{
							"currency_code": "BTC",
							"available":     "0",
						},
						{
							"currency_code": "ETH",
							"available":     "0",
						},
						{
							"currency_code": "TON",
							"available":     "0",
						},
						{
							"currency_code": "BNB",
							"available":     "0",
						},
						{
							"currency_code": "BUSD",
							"available":     "0",
						},
						{
							"currency_code": "USDC",
							"available":     "0",
						},
						{
							"currency_code": "USDT",
							"available":     "0",
						},
					},
				})
			case "/api/getExchangeRates":
				writeJson(rw, 200, JSON{
					"ok": true,
					"result": []JSON{
						{
							"is_valid": true,
							"source":   "BTC",
							"target":   "USD",
							"rate":     "40000.12",
						},
						{
							"is_valid": true,
							"source":   "BTC",
							"target":   "EUR",
							"rate":     "33468",
						},
						{
							"is_valid": true,
							"source":   "ETH",
							"target":   "USD",
							"rate":     "2604.14",
						},
						{
							"is_valid": true,
							"source":   "ETH",
							"target":   "EUR",
							"rate":     "2297",
						},
					},
				})
			case "/api/getCurrencies":
				writeJson(rw, 200, JSON{
					"ok": true,
					"result": []JSON{
						{
							"is_blockchain": true,
							"is_stablecoin": false,
							"is_fiat":       false,
							"name":          "Bitcoin",
							"code":          "BTC",
							"url":           "https://bitcoin.org/",
							"decimals":      8,
						},
						{
							"is_blockchain": true,
							"is_stablecoin": false,
							"is_fiat":       false,
							"name":          "Ethereum",
							"code":          "ETH",
							"url":           "https://ethereum.org/",
							"decimals":      18,
						},
						{
							"is_blockchain": false,
							"is_stablecoin": false,
							"is_fiat":       true,
							"name":          "Russian ruble",
							"code":          "RUB",
							"decimals":      8,
						},
						{
							"is_blockchain": false,
							"is_stablecoin": false,
							"is_fiat":       true,
							"name":          "United States dollar",
							"code":          "USD",
							"decimals":      8,
						},
					},
				})
			default:
				fmt.Fprintf(rw, apiErrorF, 404, "NOT FOUND")
			}
		}))
	})
	return apiServerInstance
}
