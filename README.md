## go-crypto-pay
[![Go](https://github.com/vitaliy-ukiru/go-crypto-pay/actions/workflows/go.yml/badge.svg)](https://github.com/vitaliy-ukiru/go-crypto-pay/actions/workflows/go.yml)

**[Crypto Pay](http://t.me/CryptoBot/?start=pay)** is a payment system based
on [@CryptoBot](http://t.me/CryptoBot), which allows you to accept payments in cryptocurrency using the
API.

This library help you to work with **Crypto Pay**
via [Crypto Pay API](https://telegra.ph/Crypto-Pay-API-11-25).

## Install

```shell
go get -u github.com/vitaliy-ukiru/go-cryptopay
```

## Documentation

For start, you need to create your application and get an API token.
Open [@CryptoBot](http://t.me/CryptoBot?start=pay)
or [@CryptoTestnetBot](http://t.me/CryptoTestnetBot?start=pay) (for testnet), send a command `/pay` to
create a new app and get API Token.

In this library, the high-level is the `Client`.  
Internally, it calls the `ApiCore` methods, which is lower-level and is essentially a gateway for API
requests. The `Client` methods, in addition to the usual errors, can return an api error `(ApiError)`. If
you want to check whether the received error is such, call the `GetApiError` function.

```go
if apiErr := cryptopay.GetApiError(err); apiErr != nil {
// handling error of api. 
// apiErr is *ApiError
}
```

### Configure NewClient

- Token - token of you app.
- ApiHost - url to api host. _Default mainnet_.
- HttpClient - client for make requests. _Default `http.DefaultClient`_.
- Webhook - webhook configure
    - OnError - handler for error handling in webhook.
    - DefaultHandler - set of default handlers. _Default empty_.

### Networks in CryptoPay:

| Net     | Bot                                                          | Hostname                       | Code reference          |
|---------|--------------------------------------------------------------|--------------------------------|-------------------------|
| mainnet | [@CryptoBot](https://t.me/CryptoBot?start=pay)               | https://pay.crypt.bot/         | `cryptopay.MainNetHost` |
| testnet | [@CryptoTestnetBot](https://t.me/CryptoTestnetBot?start=pay) | https://testnet-pay.crypt.bot/ | `cryptopay.TestNetHost` |

### Webhooks

To get started, send `/pay` command to bot, choose "My Apps", select application, open "Webhooks" and set
your endpoint.

To work with webhooks, you need to start the server yourself and install `Webhook.ServeHTTP` as a handler
to the endpoint. If you are running a "net/http" server, you can pass the `Webhook` as `http.Handler`
type. But if you don't use std server, see the [Adaptation](#Webhook-Adaptation) section.

## Examples

<details>
<summary>getMe</summary>

```go
package main

import (
  "fmt"

  "github.com/vitaliy-ukiru/go-cryptopay"
)

func main() {
  client := cryptopay.NewClient(cryptopay.ClientSettings{
    Token:   "your_token_here",
    ApiHost: cryptopay.TestNetHost,
  })
  app, err := client.GetMe()
  if err != nil {
    panic(err)
  }
  fmt.Printf(
    "app_id=%d; name=%q; payment_bot=%q",
    app.Id,
    app.Name,
    app.PaymentBotUsername,
  )

}
```

</details>

<details>
<summary>transfer</summary>


```go
package main

import (
	"fmt"
	"time"

	"github.com/vitaliy-ukiru/go-cryptopay"
)

func main() {
	client := cryptopay.NewClient(cryptopay.ClientSettings{
		Token: "your_token",
	})
	transfer, err := client.DoTransfer(-1, cryptopay.USDT, 100, "generate unique data", cryptopay.DoTransferOptions{
		Comment: "You winner!",
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("Transfer completed at %s", transfer.CompletedAt.Format(time.RFC850))
}
```

</details>

<details>
<summary>webhook</summary>

```go
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/vitaliy-ukiru/go-cryptopay"
)

func main() {
	client := cryptopay.NewClient(cryptopay.ClientSettings{
		Token: "your_token", // token required for webhooks, because using for verification updates
		Webhook: cryptopay.WebhookSettings{
			OnError: func(_ *http.Request, err error) {
				panic(err)
			},
		},
	})
	client.OnInvoicePaid(func(update *cryptopay.WebhookUpdate) {
		invoice := update.Payload
		fmt.Printf(
			"Invoice № %d for %s %s was paid on %s",
			invoice.Id,
			invoice.Amount,
			invoice.Asset,
			invoice.PaidAt.Format(time.RFC850))
	})
}
```

</details>

## Webhook Adaptation

If you use other router you can adapt. For this you must create handler that call `ServeHTTP` method.

For [gin-gonic/gin](https://github.com/gin-gonic/gin):

```go
//  router is gin.Engine
router.POST("/path/", func (c *gin.Context) {
    webhook.ServerHTTP(http.ResponseWriter(c.Writer), c.Request)
})
```

For [julienschmidt/httprouter](https://github.com/julienschmidt/httprouter):

```go
// router is httprouter.Router.
router.POST("/path", func (w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
    webhook.ServerHTTP(w, r)
})
```

For [gorilla/mux](https://github.com/gorilla/mux)

```go
// router is mux.Router
router.Handle("/path", webhook)
```
