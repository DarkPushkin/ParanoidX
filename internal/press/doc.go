// Package press implements the Island's banknote press — TemplateManager for loading,
// rendering, and verifying silver-backed digital banknotes. It manages template manifests
// (denominations, rarities, special series), generates PDF banknotes with serial numbers
// and Ed25519 digital signatures, and validates signatures via a Royal public key.
// Serial numbers follow the MB-<rarity>-<date>-<num> scheme.
package press
