// Package ton — TON blockchain integration for ARGENTUM Jetton.
//
// ARGENTUM (LATIN for "silver") is a TON Jetton (token) pegged 1:1 to
// the Liquid Taler (ng) — our silver-backed digital currency.
//
// Architecture:
//   1 ng = 1 ARGENTUM (virtual, settled via treasury)
//   Users swap TON ↔ ARGENTUM via our treasury contract
//   Treasury holds TON reserves, issues ARGENTUM on TON
//   ARGENTUM can be redeemed for ng at any time (1:1)
package ton

const (
	// Jetton constants for ARGENTUM token on TON
	ArgentumSymbol   = "ARGENTUM"
	ArgentumName     = "Liquid Taler ARGENTUM"
	ArgentumDecimals = 9 // same as ng (nanogram precision)
	ArgentumDescr    = "Silver-backed digital currency on TON — 1 ARGENTUM = 1 ng of Liquid Taler, backed by physical silver reserves"

	// Swap fee (0.5% in basis points)
	SwapFeeBPS = 50

	// Minimum swap amount in ng
	MinSwapNg = 1_000_000 // 1M ng ≈ $0.0024
)

// JettonMasterAddress will be set when the Jetton is deployed on TON.
var JettonMasterAddress = "EQD_____________________________" // placeholder

// ArgentumSwap represents a swap between TON and ARGENTUM.
type ArgentumSwap struct {
	ID          string `json:"id"`
	FromAsset   string `json:"from_asset"`   // "ton" or "argentum"
	ToAsset     string `json:"to_asset"`     // "ton" or "argentum"
	FromAmount  int64  `json:"from_amount"`  // in smallest units (nanoTON or ng)
	ToAmount    int64  `json:"to_amount"`    // after fee
	FeeNg       int64  `json:"fee_ng"`       // treasury commission (2.28%)
	UserPubkey  string `json:"user_pubkey"`
	Status      string `json:"status"`       // "pending", "completed", "failed"
	CreatedAt   string `json:"created_at"`
	CompletedAt string `json:"completed_at,omitempty"`
	TxHash      string `json:"tx_hash,omitempty"` // TON transaction hash
}

// ArgentumMarket holds market data for ARGENTUM.
type ArgentumMarket struct {
	PriceTON       float64 `json:"price_ton"`        // 1 ARGENTUM in TON
	PriceUSD       float64 `json:"price_usd"`        // 1 ARGENTUM in USD
	Volume24hNg    int64   `json:"volume_24h_ng"`
	TotalSupplyNg  int64   `json:"total_supply_ng"`
	CirculatingNg  int64   `json:"circulating_ng"`
	BackingRatio   float64 `json:"backing_ratio"`    // silver backing %
	LastUpdated    string  `json:"last_updated"`
}

// TonAPI is a client for the TON Center API.
type TonAPI struct {
	BaseURL string
	APIKey  string
}


// NewTonAPI handles the NewTonAPI HTTP request.
func NewTonAPI(apiKey string) *TonAPI {
	return &TonAPI{
		BaseURL: "https://toncenter.com/api/v2",
		APIKey:  apiKey,
	}
}

// GetRate returns the current TON/LiquidTaler rate.
// Uses an oracle: market price of TON × silver spot price.
func GetRate(silverPriceUSD float64, tonPriceUSD float64) float64 {
	if tonPriceUSD <= 0 || silverPriceUSD <= 0 {
		return 0
	}
	// 1 ng = $0.00000241 (at $75/oz silver)
	// 1 TON = tonPriceUSD
	// Rate: 1 TON = tonPriceUSD / ngPerUSD ng
	ngPerUSD := 414_713_066.0 // from SilverSpotUSDperOZ / NGPerTLR
	return tonPriceUSD * ngPerUSD
}
