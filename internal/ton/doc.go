// Package ton provides TON blockchain integration for the ARGENTUM Jetton —
// a TON token pegged 1:1 to the Liquid Taler (ng), our silver-backed digital currency.
// It defines ArgentumSwap (TON↔ARGENTUM swap tracking), ArgentumMarket (market data),
// and TonAPI (TON Center client). Rate calculation combines TON market price with
// silver spot price to maintain the ng peg. Swap fee is 0.5% (50 basis points).
package ton
