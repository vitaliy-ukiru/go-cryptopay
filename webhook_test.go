package go_crypto_pay

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"hash"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func writeHash(hash hash.Hash, data string) []byte {
	hash.Write([]byte(data))
	return hash.Sum(nil)
}

func writeHmac(key, msg []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(msg)
	return mac.Sum(nil)
}

var tokenHash = writeHash(sha256.New(), "5675:test_token")

func getWebhook(h map[UpdateType][]Handler, e ErrorHandler) *Webhook {
	if h == nil {
		h = map[UpdateType][]Handler{}
	}
	return &Webhook{
		handlers:  h,
		OnError:   e,
		tokenHash: tokenHash,
	}
}

func TestWebhook_verifyUpdate(t *testing.T) {
	requestDate := time.Now()
	update := WebhookUpdate{
		Id:          0,
		UpdateType:  UpdateInvoicePaid,
		RequestDate: requestDate,
		Payload: Invoice{
			Id:              0,
			Status:          "paid",
			Hash:            "someHash",
			Asset:           BTC,
			Amount:          "3.14",
			PayUrl:          "some.url/someHash",
			CreatedAt:       time.Now().Add(-time.Minute),
			AllowComments:   false,
			AllowAnonymous:  false,
			PaidAt:          time.Now(),
			PaidAnonymously: false,
		},
	}
	updateBytes, _ := json.Marshal(update)
	signature := writeHmac(tokenHash, updateBytes)

	w := getWebhook(nil, func(_ *http.Request, err error) {
		t.Error(err)
	})
	t.Run("test correct", func(t *testing.T) {
		if !w.verifyUpdate(updateBytes, signature) {
			t.Errorf("not equal hmac for signature(%s)", hex.EncodeToString(signature))
		}
	})
	t.Run("test incorrect", func(t *testing.T) {
		signature := append([]byte{23}, signature[1:]...)
		if w.verifyUpdate(updateBytes, signature) {
			t.Errorf("signatures must be not equal; signature(%s)", hex.EncodeToString(signature))
		}
	})

}

func TestWebhook_Bind(t *testing.T) {
	w := getWebhook(nil, nil)
	var handlerResult bool
	w.Bind("test", func(_ *WebhookUpdate) {
		handlerResult = true
	})
	w.Bind("test", func(_ *WebhookUpdate) {
		handlerResult = false
	})
	if w.handlers["test"][0](nil); !handlerResult {
		t.Error("handler 1 did not work")
	}

	if w.handlers["test"][1](nil); handlerResult {
		t.Error("handler 2 did not work")
	}
}

func TestWebhook_DeleteHandlerByIndex(t *testing.T) {
	w := getWebhook(nil, nil)
	handlerExecute := -1
	for i := 0; i < 5; i++ {
		func(index int) {
			w.Bind("test", func(_ *WebhookUpdate) {
				handlerExecute = index
			})
		}(i)
	}
	if len(w.handlers["test"]) != 5 {
		t.Error("invalid setup handlers")
	}
	if w.handlers["test"][2](nil); handlerExecute != 2 {
		t.Error("handler[2] did not work", handlerExecute)
	}
	w.DeleteHandlerByIndex("test", 2)
	if w.handlers["test"][2](nil); handlerExecute == 2 {
		t.Error("handler[2 (before 3)] did not work")
	}
}

func TestWebhook_DeleteHandlers(t *testing.T) {
	w := getWebhook(map[UpdateType][]Handler{
		UpdateInvoicePaid: {},
		"test":            {nil, nil},
		"type_2":          {nil},
	}, nil)
	t.Run("current type", func(t *testing.T) {
		w.DeleteHandlers(UpdateInvoicePaid)
		if _, ok := w.handlers[UpdateInvoicePaid]; ok {
			t.Error("handlers for invoice_paid did not deleted")
		}
	})
	t.Run("all", func(t *testing.T) {
		w.DeleteHandlers("*")
		if len(w.handlers) != 0 {
			t.Error("all handlers did not deleted")
		}
	})
}

func TestWebhook_ServeHTTP(t *testing.T) {
	w := getWebhook(nil, nil)
	server := httptest.NewServer(w)
	client := server.Client()
	t.Run("correct", func(t *testing.T) {
		data, _ := json.Marshal(WebhookUpdate{
			Id:          -1,
			UpdateType:  UpdateInvoicePaid,
			RequestDate: time.Now(),
			Payload: Invoice{
				Id:              rand.Int(),
				Status:          StatusPaid,
				Hash:            "exc1Hash",
				Asset:           USDT,
				Amount:          "1",
				PayUrl:          "/excHash",
				CreatedAt:       time.Now(),
				AllowComments:   false,
				AllowAnonymous:  true,
				PaidAt:          time.Now(),
				PaidAnonymously: true,
			},
		})
		var handled int
		w.Bind(UpdateInvoicePaid, func(update *WebhookUpdate) {
			handled = update.Id
		})
		req, _ := http.NewRequest("POST", server.URL, bytes.NewReader(data))
		req.Header.Set(headerSignatureName, hex.EncodeToString(writeHmac(tokenHash, data)))
		resp, err := client.Do(req)
		if err != nil {
			t.Error(err)
		}
		if resp.StatusCode != http.StatusOK {
			t.Errorf("resp.StatusCode(%d) != http.StatusOK", resp.StatusCode)
		}
		if handled != -1 {
			t.Errorf("handler did not work")
		}

	})
	t.Run("incorrect signature", func(t *testing.T) {
		data, _ := json.Marshal(WebhookUpdate{
			Id:          -2,
			UpdateType:  UpdateInvoicePaid,
			RequestDate: time.Now(),
			Payload: Invoice{
				Id:              rand.Int(),
				Status:          StatusPaid,
				Hash:            "exc2Hash",
				Asset:           USDT,
				Amount:          "1",
				PayUrl:          "/exc2Hash",
				CreatedAt:       time.Now(),
				AllowComments:   false,
				AllowAnonymous:  true,
				PaidAt:          time.Now(),
				PaidAnonymously: true,
			},
		})
		req, _ := http.NewRequest("POST", server.URL, bytes.NewReader(data))
		resp, err := client.Do(req)
		if err != nil {
			t.Error(err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Errorf("resp.StatusCode(%d) != http.StatusBadRequest", resp.StatusCode)
		}
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			t.Error(err)
		}
		if string(body) != wrongSignature {
			t.Error("body != wrong signature")
		}
	})
	t.Run("incorrect json", func(t *testing.T) {
		defer func() {
			w.OnError = nil
		}()
		w.OnError = func(_ *http.Request, err error) {
			if !strings.Contains(err.Error(), "invalid character") {
				t.Error("error not contains json error")
			}
		}
		data := []byte("some body")
		req, _ := http.NewRequest("POST", server.URL, bytes.NewReader(data))
		req.Header.Set(headerSignatureName, hex.EncodeToString(writeHmac(tokenHash, data)))
		resp, err := client.Do(req)
		if err != nil {
			t.Error(err)
		}
		if resp.StatusCode != http.StatusBadRequest {
			t.Error("invalid status code")
		}
	})
}
