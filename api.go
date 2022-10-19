package cryptopay

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Aliases for officials hosts
//goland:noinspection ALL
const (
	MainNetHost = "https://pay.crypt.bot"
	TestNetHost = "https://testnet-pay.crypt.bot"
)

const (
	headerTokenName = "Crypto-Pay-API-Token"
)

type Response[T any] struct {
	Ok     bool      `json:"ok"`
	Error  *ApiError `json:"error,omitempty"`
	Result T         `json:"result,omitempty"`
}

type Api struct {
	token      string
	url        string
	httpClient *http.Client
}

// NewApi returns new Api
func NewApi(token, url string, httpClient *http.Client) *Api {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}

	if url == "" {
		url = MainNetHost
	}

	return &Api{token: token, url: url, httpClient: httpClient}
}

func (a *Api) HttpClient() *http.Client {
	return a.httpClient
}

// urlFmt formatting URL, paste query params
func (a *Api) urlFmt(method string, queryParams url.Values) string {
	methodUrl := fmt.Sprintf("%s/api/%s", a.url, method)
	if queryParams == nil {
		return methodUrl
	}
	return methodUrl + "?" + queryParams.Encode()
}

func (a *Api) Do(r Request, dest any) error {
	req, err := http.NewRequest("GET", a.urlFmt(r.Endpoint(), r.Query()), nil)
	if err != nil {
		return err
	}
	req.Header.Set(headerTokenName, a.token)
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return err
	}
	return json.NewDecoder(resp.Body).Decode(&dest)
}

type Request interface {
	Endpoint() string
	Query() url.Values
}
