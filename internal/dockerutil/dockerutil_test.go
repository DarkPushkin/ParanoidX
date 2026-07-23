// Package dockerutil provides Docker container management utilities
package dockerutil

import (
	"testing"
)


// TestApplyStatusSMP handles the TestApplyStatusSMP HTTP request.
func TestApplyStatusSMP(t *testing.T) {
	smp := "unknown"
	xftp := "unknown"

	ApplyStatus(&smp, &xftp, map[string]any{"name": "smp"})

	if smp != "running" {
		t.Fatalf("expected smp= running, got %s", smp)
	}
	if xftp != "unknown" {
		t.Fatalf("expected xftp unchanged, got %s", xftp)
	}
}


// TestApplyStatusXFTP handles the TestApplyStatusXFTP HTTP request.
func TestApplyStatusXFTP(t *testing.T) {
	smp := "unknown"
	xftp := "unknown"

	ApplyStatus(&smp, &xftp, map[string]any{"name": "xftp"})

	if xftp != "running" {
		t.Fatalf("expected xftp= running, got %s", xftp)
	}
	if smp != "unknown" {
		t.Fatalf("expected smp unchanged, got %s", smp)
	}
}


// TestApplyStatusUnknown handles the TestApplyStatusUnknown HTTP request.
func TestApplyStatusUnknown(t *testing.T) {
	smp := "unknown"
	xftp := "unknown"

	ApplyStatus(&smp, &xftp, map[string]any{"name": "other"})

	if smp != "unknown" || xftp != "unknown" {
		t.Fatal("expected both unchanged for unknown name")
	}
}


// TestApplyStatusNoName handles the TestApplyStatusNoName HTTP request.
func TestApplyStatusNoName(t *testing.T) {
	smp := "unknown"
	xftp := "unknown"

	ApplyStatus(&smp, &xftp, map[string]any{})

	if smp != "unknown" || xftp != "unknown" {
		t.Fatal("expected both unchanged when no name")
	}
}
