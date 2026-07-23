// Package economy implements the economic engine for simplex-node.
//
// This is the largest package (~44 files, ~9,500 lines) and covers all
// financial operations. The economy module is organized into sub-systems:
//
//   - Oracle:        Silver spot price from multiple sources (gold-api, Swissquote, metals.dev)
//   - Treasury:      Treasury state, forecasting, dividend distribution
//   - Assets:        Silver-backed asset mint/burn/list, RWA registry
//   - Reserve:       Proof-of-reserve with audit trail and backing ratio
//   - Invoice:       Invoice creation, payment, expiry, webhook
//   - Subscription:  Subscription management
//   - Onboarding:    New user financial onboarding
//   - Wheel:         Wheel of fortune / gamification
//   - Mining:        Mining rewards
//   - Crafting:      Crafting recipes
//   - Deflation:     Deflation mechanisms
//   - Argentum:      Silver token management
//   - Advertising:   Advertising marketplace
//   - Franchise:     Licensing, royalties, mint auth, settlements
//   - Services:      Service registry and marketplace
//   - Arbitration:   Dispute resolution
//   - POS:           Point-of-sale for merchant payments
//   - Lock:          Genesis token lock
//   - ICO:           Initial coin offering
//   - Swap:          Cross-chain atomic swap (BTC, ETH)
//   - Pack:          Economy pack management
//   - Buyback:       Token buyback management
//   - Auction:       Auction system
//   - Tokenomics:    Token supply, distribution, and economics
//   - Rates:         Multi-currency exchange rates
//   - DividendAdmin: Manual dividend trigger and history
//   - InvoiceWebhookTest: Webhook test endpoint
//   - TreasuryForecast: AI-powered treasury health forecasting
//
// All economy data is persisted to JSON files in the data directory
// and/or SQLite database files.
package economy
