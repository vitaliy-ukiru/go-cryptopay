package cryptopay

type Asset string

//goland:noinspection ALL
const (
	BTC  Asset = "BTC"
	TON  Asset = "TON"
	ETH  Asset = "ETH"
	USDT Asset = "USDT"
	USDC Asset = "USDC"
	BUSD Asset = "BUSD"
)

// PaidButton is name of the button that will be shown to a user after the invoice is paid.
type PaidButton string

//goland:noinspection ALL
const (
	ButtonViewItem    PaidButton = "viewItem"
	ButtonOpenChannel PaidButton = "openChannel"
	ButtonOpenBot     PaidButton = "openBot"
	ButtonCallback    PaidButton = "callback"
)

// InvoiceStatus is status of the invoice.
type InvoiceStatus string

//goland:noinspection ALL
const (
	StatusActive  InvoiceStatus = "active"
	StatusPaid    InvoiceStatus = "paid"
	StatusExpired InvoiceStatus = "expired"
)

func (a Asset) String() string         { return string(a) }
func (p PaidButton) String() string    { return string(p) }
func (i InvoiceStatus) String() string { return string(i) }
