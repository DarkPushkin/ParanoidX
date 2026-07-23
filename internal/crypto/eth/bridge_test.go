// Package eth provides Ethereum bridge and status tracking
package eth

import (
	"fmt"
	"testing"
)


// TestCreateBridgeTransfer handles the TestCreateBridgeTransfer HTTP request.
func TestCreateBridgeTransfer(t *testing.T) {
	r := NewBridgeRegistry()
	tf := r.Create("ethereum", "polygon", "ARG", "0xsender", "0xrecipient", 1000000)
	if tf.Status != BridgePending {
		t.Fatalf("expected pending, got %s", tf.Status)
	}
	if tf.ID == "" {
		t.Fatal("expected non-empty ID")
	}
}


// TestConfirmAndComplete handles the TestConfirmAndComplete HTTP request.
func TestConfirmAndComplete(t *testing.T) {
	r := NewBridgeRegistry()
	tf := r.Create("ethereum", "base", "USDC", "0xa", "0xb", 5000)
	tf2, err := r.Confirm(tf.ID, "0xconfirmtx")
	if err != nil {
		t.Fatal(err)
	}
	if tf2.Status != BridgeVerified {
		t.Fatalf("expected verified, got %s", tf2.Status)
	}
	tf3, err := r.Complete(tf.ID, "0xprooftx")
	if err != nil {
		t.Fatal(err)
	}
	if tf3.Status != BridgeComplete {
		t.Fatalf("expected complete, got %s", tf3.Status)
	}
	if tf3.CompletedAt == "" {
		t.Fatal("expected completed_at")
	}
}


// TestBridgeList handles the TestBridgeList HTTP request.
func TestBridgeList(t *testing.T) {
	r := NewBridgeRegistry()
	r.Create("eth", "polygon", "ARG", "a", "b", 100)
	r.Create("polygon", "eth", "ARG", "b", "a", 200)
	transfers := r.List()
	if len(transfers) != 2 {
		t.Fatalf("expected 2, got %d", len(transfers))
	}
}


// TestBridgeFail handles the TestBridgeFail HTTP request.
func TestBridgeFail(t *testing.T) {
	r := NewBridgeRegistry()
	tf := r.Create("eth", "base", "ARG", "a", "b", 100)
	r.Fail(tf.ID, fmt.Errorf("out of gas"))
	if r.Get(tf.ID).Status != BridgeFailed {
		t.Fatal("expected failed status")
	}
}
