// Package economy implements the island economy system
package economy

import (
	"testing"
)


// TestCreateDispute handles the TestCreateDispute HTTP request.
func TestCreateDispute(t *testing.T) {
	dir := t.TempDir()
	am := NewArbitrationManager(dir)

	d, err := am.CreateDispute("alice", "bob", "Payment dispute", "Alice paid but Bob didn't deliver", "TX12345")
	if err != nil {
		t.Fatal(err)
	}
	if d.Status != DisputeStatusOpen {
		t.Errorf("expected open, got %s", d.Status)
	}
	if d.Initiator != "alice" {
		t.Errorf("expected alice, got %s", d.Initiator)
	}
}


// TestCreateDisputeSelf handles the TestCreateDisputeSelf HTTP request.
func TestCreateDisputeSelf(t *testing.T) {
	dir := t.TempDir()
	am := NewArbitrationManager(dir)

	_, err := am.CreateDispute("alice", "alice", "Self dispute", "", "")
	if err == nil {
		t.Error("expected error for self-dispute")
	}
}


// TestRespondToDispute handles the TestRespondToDispute HTTP request.
func TestRespondToDispute(t *testing.T) {
	dir := t.TempDir()
	am := NewArbitrationManager(dir)

	d, _ := am.CreateDispute("alice", "bob", "Test", "desc", "evidence1")
	d, err := am.Respond(d.ID, "bob", "evidence2")
	if err != nil {
		t.Fatal(err)
	}
	if d.Status != DisputeStatusResponded {
		t.Errorf("expected responded, got %s", d.Status)
	}
	if len(d.Evidence) != 2 {
		t.Errorf("expected 2 evidence, got %d", len(d.Evidence))
	}
}


// TestRespondWrongParty handles the TestRespondWrongParty HTTP request.
func TestRespondWrongParty(t *testing.T) {
	dir := t.TempDir()
	am := NewArbitrationManager(dir)

	d, _ := am.CreateDispute("alice", "bob", "Test", "desc", "ev1")
	_, err := am.Respond(d.ID, "charlie", "ev2")
	if err == nil {
		t.Error("expected error for wrong party")
	}
}


// TestFullDisputeLifecycle handles the TestFullDisputeLifecycle HTTP request.
func TestFullDisputeLifecycle(t *testing.T) {
	dir := t.TempDir()
	am := NewArbitrationManager(dir)

	d, _ := am.CreateDispute("alice", "bob", "Full test", "desc", "initiator_evidence")
	d, _ = am.Respond(d.ID, "bob", "respondent_evidence")
	d, err := am.Analyze(d.ID, "AI recommends: Alice is correct")
	if err != nil {
		t.Fatal(err)
	}
	if d.Status != DisputeStatusAnalyzed {
		t.Errorf("expected analyzed, got %s", d.Status)
	}

	d, err = am.IssueRuling(d.ID, "Alice wins. Bob must pay 5 TLR.", "steward")
	if err != nil {
		t.Fatal(err)
	}
	if d.Status != DisputeStatusRuled {
		t.Errorf("expected ruled, got %s", d.Status)
	}

	d, err = am.Appeal(d.ID, "bob", "Evidence was forged")
	if err != nil {
		t.Fatal(err)
	}
	if d.Status != DisputeStatusAppealed {
		t.Errorf("expected appealed, got %s", d.Status)
	}
}


// TestListDisputesByPubkey handles the TestListDisputesByPubkey HTTP request.
func TestListDisputesByPubkey(t *testing.T) {
	dir := t.TempDir()
	am := NewArbitrationManager(dir)

	am.CreateDispute("alice", "bob", "D1", "desc", "ev1")
	am.CreateDispute("bob", "charlie", "D2", "desc", "ev2")

	bobDisputes := am.ListDisputes("bob")
	if len(bobDisputes) != 2 {
		t.Errorf("expected 2 disputes for bob, got %d", len(bobDisputes))
	}

	aliceDisputes := am.ListDisputes("alice")
	if len(aliceDisputes) != 1 {
		t.Errorf("expected 1 dispute for alice, got %d", len(aliceDisputes))
	}
}


// TestDisputePersistence handles the TestDisputePersistence HTTP request.
func TestDisputePersistence(t *testing.T) {
	dir := t.TempDir()

	am1 := NewArbitrationManager(dir)
	d, _ := am1.CreateDispute("alice", "bob", "Persist test", "desc", "ev1")

	am2 := NewArbitrationManager(dir)
	d2, err := am2.GetDispute(d.ID)
	if err != nil {
		t.Fatal(err)
	}
	if d2.Initiator != "alice" {
		t.Errorf("expected alice, got %s", d2.Initiator)
	}
}
