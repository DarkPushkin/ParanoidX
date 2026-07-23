// Package api provides HTTP handlers and API endpoints for the simplex-node server
package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ── P2P Marketplace (Cycle C2) ──────────────────────────────────────────────

// Listing represents a sell listing in the P2P marketplace.
type Listing struct {
	ID          string    `json:"id"`
	SellerID    string    `json:"seller_id"`    // SimpleX contact ID
	SellerName  string    `json:"seller_name"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	PriceNg     int64     `json:"price_ng"`
	Category    string    `json:"category"`
	Status      string    `json:"status"` // active, sold, cancelled
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Offer represents a buy offer on a marketplace listing.
type Offer struct {
	ID         string    `json:"id"`
	ListingID  string    `json:"listing_id"`
	BuyerID    string    `json:"buyer_id"`
	BuyerName  string    `json:"buyer_name"`
	PriceNg    int64     `json:"price_ng"`
	Message    string    `json:"message"`
	Status     string    `json:"status"` // pending, accepted, rejected, cancelled
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// MarketplaceStore manages listings and offers with JSON persistence.
type MarketplaceStore struct {
	mu       sync.RWMutex
	path     string
	Listings []Listing `json:"listings"`
	Offers   []Offer   `json:"offers"`
	nextID   int64
}


// NewMarketplaceStore creates a MarketplaceStore and loads persisted data from disk.
func NewMarketplaceStore(dataDir string) *MarketplaceStore {
	store := &MarketplaceStore{
		path:   filepath.Join(dataDir, "marketplace.json"),
		nextID: 1,
	}
	store.load()
	return store
}

func (s *MarketplaceStore) load() {
	data, err := os.ReadFile(s.path)
	if err != nil {
		return
	}
	if err := json.Unmarshal(data, s); err != nil {
		slog.Warn("marketplace load", "error", err)
		return
	}
	// Find next ID
	for _, l := range s.Listings {
		if id := parseID(l.ID); id >= s.nextID {
			s.nextID = id + 1
		}
	}
	for _, o := range s.Offers {
		if id := parseID(o.ID); id >= s.nextID {
			s.nextID = id + 1
		}
	}
	slog.Info("marketplace loaded", "listings", len(s.Listings), "offers", len(s.Offers))
}

func (s *MarketplaceStore) save() {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		slog.Error("marketplace save", "error", err)
		return
	}
	if err := os.WriteFile(s.path, data, 0644); err != nil {
		slog.Error("marketplace save write", "error", err)
	}
}

func parseID(id string) int64 {
	var n int64
	fmt.Sscanf(id, "mkt-%d", &n)
	return n
}

func (s *MarketplaceStore) nextIDStr(prefix string) string {
	s.nextID++
	return fmt.Sprintf("%s-%d", prefix, s.nextID-1)
}


// CreateListing handles the CreateListing HTTP request.
func (s *MarketplaceStore) CreateListing(sellerID, sellerName, title, desc string, priceNg int64, category string) *Listing {
	s.mu.Lock()
	defer s.mu.Unlock()
	l := &Listing{
		ID:          s.nextIDStr("lst"),
		SellerID:    sellerID,
		SellerName:  sellerName,
		Title:       title,
		Description: desc,
		PriceNg:     priceNg,
		Category:    category,
		Status:      "active",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	s.Listings = append(s.Listings, *l)
	s.save()
	return l
}


// GetActiveListings handles the GetActiveListings HTTP request.
func (s *MarketplaceStore) GetActiveListings() []Listing {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Listing
	for _, l := range s.Listings {
		if l.Status == "active" {
			out = append(out, l)
		}
	}
	return out
}


// CreateOffer handles the CreateOffer HTTP request.
func (s *MarketplaceStore) CreateOffer(listingID, buyerID, buyerName string, priceNg int64, message string) *Offer {
	s.mu.Lock()
	defer s.mu.Unlock()
	o := &Offer{
		ID:        s.nextIDStr("off"),
		ListingID: listingID,
		BuyerID:   buyerID,
		BuyerName: buyerName,
		PriceNg:   priceNg,
		Message:   message,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	s.Offers = append(s.Offers, *o)
	s.save()
	return o
}


// GetOffers handles the GetOffers HTTP request.
func (s *MarketplaceStore) GetOffers(listingID string) []Offer {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Offer
	for _, o := range s.Offers {
		if o.ListingID == listingID {
			out = append(out, o)
		}
	}
	return out
}


// AcceptOffer handles the AcceptOffer HTTP request.
func (s *MarketplaceStore) AcceptOffer(offerID string) (*Offer, *Listing) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, o := range s.Offers {
		if o.ID == offerID && o.Status == "pending" {
			s.Offers[i].Status = "accepted"
			s.Offers[i].UpdatedAt = time.Now()
			// Mark listing as sold
			for j, l := range s.Listings {
				if l.ID == o.ListingID {
					s.Listings[j].Status = "sold"
					s.Listings[j].UpdatedAt = time.Now()
					s.save()
					return &s.Offers[i], &s.Listings[j]
				}
			}
		}
	}
	return nil, nil
}


// RejectOffer handles the RejectOffer HTTP request.
func (s *MarketplaceStore) RejectOffer(offerID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, o := range s.Offers {
		if o.ID == offerID && o.Status == "pending" {
			s.Offers[i].Status = "rejected"
			s.Offers[i].UpdatedAt = time.Now()
			s.save()
			return true
		}
	}
	return false
}


// CancelListing handles the CancelListing HTTP request.
func (s *MarketplaceStore) CancelListing(listingID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, l := range s.Listings {
		if l.ID == listingID && l.Status == "active" {
			s.Listings[i].Status = "cancelled"
			s.Listings[i].UpdatedAt = time.Now()
			s.save()
			return true
		}
	}
	return false
}

// ── HTTP Handler ─────────────────────────────────────────────────────────────

var globalMarketplace *MarketplaceStore


// InitMarketplace handles the InitMarketplace HTTP request.
func InitMarketplace(dataDir string) {
	globalMarketplace = NewMarketplaceStore(dataDir)
}


// MarketplaceHandler handles the MarketplaceHandler HTTP request.
func MarketplaceHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if globalMarketplace == nil {
			writeJSON(w, map[string]any{"ok": false, "error": "marketplace not initialized"})
			return
		}

		switch r.Method {
		case "GET":
			// GET /api/marketplace/listings — get active listings
			// GET /api/marketplace/offers?listing_id=X — get offers for listing
			listingID := r.URL.Query().Get("listing_id")
			if listingID != "" {
				offers := globalMarketplace.GetOffers(listingID)
				writeJSON(w, map[string]any{
					"ok":      true,
					"offers":  offers,
					"count":   len(offers),
				})
				return
			}
			listings := globalMarketplace.GetActiveListings()
			writeJSON(w, map[string]any{
				"ok":       true,
				"listings": listings,
				"count":    len(listings),
			})

		case "POST":
			var req struct {
				Action      string `json:"action"`
				Title       string `json:"title"`
				Description string `json:"description"`
				PriceNg     int64  `json:"price_ng"`
				Category    string `json:"category"`
				SellerID    string `json:"seller_id"`
				SellerName  string `json:"seller_name"`
				ListingID   string `json:"listing_id"`
				BuyerID     string `json:"buyer_id"`
				BuyerName   string `json:"buyer_name"`
				Message     string `json:"message"`
				OfferID     string `json:"offer_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "bad json", 400)
				return
			}

			switch req.Action {
			case "create":
				if req.Title == "" || req.PriceNg <= 0 {
					http.Error(w, "title and price_ng required", 400)
					return
				}
				l := globalMarketplace.CreateListing(req.SellerID, req.SellerName, req.Title, req.Description, req.PriceNg, req.Category)
				writeJSON(w, map[string]any{"ok": true, "listing": l})

			case "offer":
				if req.ListingID == "" || req.PriceNg <= 0 {
					http.Error(w, "listing_id and price_ng required", 400)
					return
				}
				o := globalMarketplace.CreateOffer(req.ListingID, req.BuyerID, req.BuyerName, req.PriceNg, req.Message)
				writeJSON(w, map[string]any{"ok": true, "offer": o})

			case "accept":
				if req.OfferID == "" {
					http.Error(w, "offer_id required", 400)
					return
				}
				offer, listing := globalMarketplace.AcceptOffer(req.OfferID)
				if offer == nil {
					http.Error(w, "offer not found or already processed", 404)
					return
				}
				writeJSON(w, map[string]any{"ok": true, "offer": offer, "listing": listing})

			case "reject":
				if req.OfferID == "" {
					http.Error(w, "offer_id required", 400)
					return
				}
				if globalMarketplace.RejectOffer(req.OfferID) {
					writeJSON(w, map[string]any{"ok": true})
				} else {
					http.Error(w, "offer not found or already processed", 404)
				}

			case "cancel":
				if req.ListingID == "" {
					http.Error(w, "listing_id required", 400)
					return
				}
				if globalMarketplace.CancelListing(req.ListingID) {
					writeJSON(w, map[string]any{"ok": true})
				} else {
					http.Error(w, "listing not found or already sold", 404)
				}

			default:
				http.Error(w, fmt.Sprintf("unknown action: %s", req.Action), 400)
			}

		default:
			http.Error(w, "GET or POST", 400)
		}
	}
}
