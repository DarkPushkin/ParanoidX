// Package treasury implements financial reserve management.
//
// Handles:
//   - Treasury state tracking and proof-of-reserve
//   - Health score calculation and forecasting
//   - Silver-backed asset backing ratio
//   - Reserve audit trail
//
// The treasury monitors the silver reserve balance, calculates the
// backing ratio for silver-backed assets, and provides forecasts
// for economic planning.
package treasury
