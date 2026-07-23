// Package economy — Hybrid Digital Silver Standard (30/70).
//
// Каждый TLR состоит из:
//   70% — физическое серебро в резерве (21,772,436,000 ng)
//   30% — утилитарная ценность сети  (9,331,044,000 ng)
//
// Номинальная цена 1 TLR = 31,103,480,000 ng (1 тройская унция по споту)
// Премия 30% встроена в механизм эмиссии: 70 oz → 100 TLR,
// где 30 TLR премии распределяются по 5 пулам.
package economy

// UtilityPremiumPct — доля ценности TLR, обеспеченная сетью (удобство расчётов).
const UtilityPremiumPct = 0.30

// TreasuryCommissionBPS — казна удерживает 2.28% со всех комиссий (операционный бюджет).
const TreasuryCommissionBPS = 228

// MaxTotalFeeBPS — максимальная общая комиссия в системе (4.20%).
const MaxTotalFeeBPS = 420

// DividendShareBPS — доля дивидендного фонда (всё, что выше 2.28% и до 4.20%).
const DividendShareBPS = MaxTotalFeeBPS - TreasuryCommissionBPS // 192 = 1.92%

// === Базовые константы (все в ng, кратные 1 ng) ===

const (
	// SilverPortionNg — серебряная часть 1 TLR (70%)
	SilverPortionNg = int64(NGPerTLR * 70 / 100) // 21,772,436,000

	// UtilityPremiumNg — премия за удобство 1 TLR (30%)
	UtilityPremiumNg = int64(NGPerTLR * 30 / 100) // 9,331,044,000

	// PremiumPerIssue — премия за эмиссию 100 TLR (30% × 100 = 30 × NGPerTLR)
	PremiumPerIssue = TLRPerIssue * UtilityPremiumNg // 30 × NGPerTLR = 933,104,400,000
)

// === Распределение премии (на 1 эмиссионную единицу: 70 oz → 100 TLR) ===
//
//	Категория                 % от премии  % от NgPerIssue  ng (кратно NGPerTLR/10)
//	Казна (Treasury)               8%         2.4%           24 × NGPerTLR/10
//	Дивидендный пул (Dividend)    47%        14.1%          141 × NGPerTLR/10
//	Премия за серебро (BuySilver)  10%         3.0%           30 × NGPerTLR/10
//	Аукционный пул (Auction)      15%         4.5%           45 × NGPerTLR/10
//	Резерв выкупа (Buyback)       20%         6.0%           60 × NGPerTLR/10
//	                             100%        30.0%

const (
	TreasuryPremiumNg   = 24 * NGPerTLR / 10 // 74,648,352,000 (2.4%)
	DividendPremiumNg   = 141 * NGPerTLR / 10 // 438,559,068,000 (14.1%)
	SilverBuyPremiumNg  = 30 * NGPerTLR / 10  // 93,310,440,000 (3.0%)
	AuctionPremiumNg    = 45 * NGPerTLR / 10  // 139,965,660,000 (4.5%)
	BuybackPremiumNg    = 60 * NGPerTLR / 10  // 186,620,880,000 (6.0%)
)

// === Веса редкости для дивидендов и аукционов ===

const (
	RarityWeightCommon    = 1
	RarityWeightRare      = 2
	RarityWeightEpic      = 5
	RarityWeightLegendary = 10
	RarityWeightGenesis   = 20
)

// RarityWeight возвращает дивидендный вес для редкости.
func RarityWeight(rarity string) int {
	switch rarity {
	case "common":
		return RarityWeightCommon
	case "rare":
		return RarityWeightRare
	case "epic":
		return RarityWeightEpic
	case "legendary":
		return RarityWeightLegendary
	case "genesis":
		return RarityWeightGenesis
	default:
		return 1
	}
}

// === Комиссии аукциона (в базисных пунктах, BPS) ===

const (
	// AuctionListingFeeBPS — 0.5% за выставление лота (от startPrice)
	AuctionListingFeeBPS = 50

	// AuctionBuyerPremiumBPS — 2.5% premium с покупателя сверх ставки
	AuctionBuyerPremiumBPS = 250

	// AuctionSellerFeeBPS — 1% комиссия с продавца (из выручки)
	AuctionSellerFeeBPS = 100

	// AuctionMinListingFeeNg — минимальная плата за листинг 1,000,000 ng
	AuctionMinListingFeeNg = 1_000_000
)

// === Структуры ===

// PremiumAllocation описывает распределение премии одной эмиссии (100 TLR).
type PremiumAllocation struct {
	TreasuryNg  int64 `json:"treasury_ng"`
	DividendNg  int64 `json:"dividend_ng"`
	SilverBuyNg int64 `json:"silver_buy_ng"`
	AuctionNg   int64 `json:"auction_ng"`
	BuybackNg   int64 `json:"buyback_ng"`
	TotalNg     int64 `json:"total_ng"`
}

// AllocatePremium возвращает распределение премии для одной эмиссии (100 TLR).
func AllocatePremium() PremiumAllocation {
	return PremiumAllocation{
		TreasuryNg:  TreasuryPremiumNg,
		DividendNg:  DividendPremiumNg,
		SilverBuyNg: SilverBuyPremiumNg,
		AuctionNg:   AuctionPremiumNg,
		BuybackNg:   BuybackPremiumNg,
		TotalNg:     PremiumPerIssue,
	}
}

// IssuanceFull описывает полный результат эмиссии 70 oz → 100 TLR.
// Включает серебряную часть (инвестору) и распределение премии по 5 пулам.
type IssuanceFull struct {
	InvestorNg      int64             `json:"investor_ng"`
	TreasuryNg      int64             `json:"treasury_ng"`
	DividendPoolNg  int64             `json:"dividend_pool_ng"`
	SilverBuyPoolNg int64             `json:"silver_buy_pool_ng"`
	AuctionPoolNg   int64             `json:"auction_pool_ng"`
	BuybackPoolNg   int64             `json:"buyback_pool_ng"`
	TotalNg         int64             `json:"total_ng"`
	Premium         PremiumAllocation `json:"premium"`
}

// CalculateFullIssuance возвращает полную раскладку одной эмиссии.
func CalculateFullIssuance() IssuanceFull {
	premium := AllocatePremium()
	return IssuanceFull{
		InvestorNg:      SilverPortionNg * TLRPerIssue,
		TreasuryNg:      premium.TreasuryNg,
		DividendPoolNg:  premium.DividendNg,
		SilverBuyPoolNg: premium.SilverBuyNg,
		AuctionPoolNg:   premium.AuctionNg,
		BuybackPoolNg:   premium.BuybackNg,
		TotalNg:         premium.TotalNg + SilverPortionNg*TLRPerIssue,
		Premium:         premium,
	}
}
