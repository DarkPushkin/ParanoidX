// Package royal implements the Royal API with treasury and economy endpoints
package royal

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)


// TestNewService handles the TestNewService HTTP request.
func TestNewService(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	if svc == nil {
		t.Fatal("NewService returned nil")
	}
	if svc.Pubkey == nil {
		t.Fatal("expected pubkey to be generated")
	}
	if svc.Privkey == nil {
		t.Fatal("expected privkey to be generated")
	}
	if len(svc.Pubkey) != ed25519.PublicKeySize {
		t.Fatalf("bad pubkey size: %d", len(svc.Pubkey))
	}
	if len(svc.Privkey) != ed25519.PrivateKeySize {
		t.Fatalf("bad privkey size: %d", len(svc.Privkey))
	}
	if svc.Nodes == nil {
		t.Fatal("expected non-nil Nodes map")
	}
	if len(svc.Nodes) != 0 {
		t.Fatalf("expected empty nodes, got %d", len(svc.Nodes))
	}
	if svc.DataDir != dir {
		t.Fatalf("DataDir mismatch: %s vs %s", svc.DataDir, dir)
	}
}


// TestNewServiceLoadExistingKey handles the TestNewServiceLoadExistingKey HTTP request.
func TestNewServiceLoadExistingKey(t *testing.T) {
	dir := t.TempDir()
	svc1 := NewService(dir)
	pubHex := svc1.PublicKeyHex()

	svc2 := NewService(dir)
	if svc2.PublicKeyHex() != pubHex {
		t.Fatal("expected same pubkey after reload")
	}
}


// TestNewServiceLoadExistingNodes handles the TestNewServiceLoadExistingNodes HTTP request.
func TestNewServiceLoadExistingNodes(t *testing.T) {
	dir := t.TempDir()
	svc1 := NewService(dir)
	svc1.RegisterNode("abc123", "test-node", "onion:abc.onion")

	svc2 := NewService(dir)
	if len(svc2.Nodes) != 1 {
		t.Fatalf("expected 1 node after reload, got %d", len(svc2.Nodes))
	}
	if svc2.Nodes["abc123"].Label != "test-node" {
		t.Fatalf("unexpected label: %s", svc2.Nodes["abc123"].Label)
	}
}


// TestPublicKeyHex handles the TestPublicKeyHex HTTP request.
func TestPublicKeyHex(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	hexKey := svc.PublicKeyHex()
	if hexKey == "" {
		t.Fatal("expected non-empty pubkey hex")
	}
	decoded, err := hex.DecodeString(hexKey)
	if err != nil {
		t.Fatalf("hex decode error: %v", err)
	}
	if len(decoded) != ed25519.PublicKeySize {
		t.Fatalf("bad decoded key size: %d", len(decoded))
	}
}


// TestPublicKeyHexNilKey handles the TestPublicKeyHexNilKey HTTP request.
func TestPublicKeyHexNilKey(t *testing.T) {
	svc := &Service{}
	if got := svc.PublicKeyHex(); got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}


// TestRegisterNode handles the TestRegisterNode HTTP request.
func TestRegisterNode(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	node, err := svc.RegisterNode("pubkey1", "alpha", "addr1")
	if err != nil {
		t.Fatalf("RegisterNode failed: %v", err)
	}
	if node.Pubkey != "pubkey1" {
		t.Fatalf("bad pubkey: %s", node.Pubkey)
	}
	if node.Label != "alpha" {
		t.Fatalf("bad label: %s", node.Label)
	}
	if node.Addr != "addr1" {
		t.Fatalf("bad addr: %s", node.Addr)
	}
	if node.Type != NodeTypeSub {
		t.Fatalf("expected sub type, got %s", node.Type)
	}
	if node.Status != StatusOnline {
		t.Fatalf("expected online status, got %s", node.Status)
	}
	if node.Registered == "" {
		t.Fatal("expected registered timestamp")
	}
	if node.LastSeen == "" {
		t.Fatal("expected last_seen timestamp")
	}
}


// TestRegisterNodeDuplicate handles the TestRegisterNodeDuplicate HTTP request.
func TestRegisterNodeDuplicate(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	svc.RegisterNode("dupkey", "first", "addr")
	_, err := svc.RegisterNode("dupkey", "second", "otheraddr")
	if err == nil {
		t.Fatal("expected error for duplicate registration")
	}
	if !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("unexpected error: %v", err)
	}
}


// TestRegisterNodePersistence handles the TestRegisterNodePersistence HTTP request.
func TestRegisterNodePersistence(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	svc.RegisterNode("persist1", "persistent", "addr")
	svc.RegisterNode("persist2", "persistent2", "addr2")

	svc2 := NewService(dir)
	if len(svc2.Nodes) != 2 {
		t.Fatalf("expected 2 persisted nodes, got %d", len(svc2.Nodes))
	}
	if svc2.Nodes["persist1"].Label != "persistent" {
		t.Fatalf("bad label for persist1: %s", svc2.Nodes["persist1"].Label)
	}
	if svc2.Nodes["persist2"].Label != "persistent2" {
		t.Fatalf("bad label for persist2: %s", svc2.Nodes["persist2"].Label)
	}
}


// TestSignCommand handles the TestSignCommand HTTP request.
func TestSignCommand(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	sc, err := svc.SignCommand("restart", "target123")
	if err != nil {
		t.Fatalf("SignCommand failed: %v", err)
	}
	if sc.Command != "restart" {
		t.Fatalf("bad command: %s", sc.Command)
	}
	if sc.Target != "target123" {
		t.Fatalf("bad target: %s", sc.Target)
	}
	if sc.Nonce == "" {
		t.Fatal("expected nonce")
	}
	if sc.Timestamp == 0 {
		t.Fatal("expected timestamp")
	}
	if sc.Signature == "" {
		t.Fatal("expected signature")
	}
	if len(sc.Signature) != hex.EncodedLen(ed25519.SignatureSize) {
		t.Fatalf("bad signature hex length: %d", len(sc.Signature))
	}
}


// TestSignCommandNoKey handles the TestSignCommandNoKey HTTP request.
func TestSignCommandNoKey(t *testing.T) {
	svc := &Service{}
	_, err := svc.SignCommand("restart", "target")
	if err == nil {
		t.Fatal("expected error when no keypair loaded")
	}
	if !strings.Contains(err.Error(), "keypair not loaded") {
		t.Fatalf("unexpected error: %v", err)
	}
}


// TestVerifyCommand handles the TestVerifyCommand HTTP request.
func TestVerifyCommand(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	sc, _ := svc.SignCommand("reboot", "sub1")
	if !VerifyCommand(sc, svc.Pubkey) {
		t.Fatal("expected VerifyCommand to return true with correct key")
	}

	// wrong key
	wrongPub, _, _ := ed25519.GenerateKey(rand.Reader)
	if VerifyCommand(sc, wrongPub) {
		t.Fatal("expected VerifyCommand to return false with wrong key")
	}
}


// TestVerifyCommandTampered handles the TestVerifyCommandTampered HTTP request.
func TestVerifyCommandTampered(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	sc, _ := svc.SignCommand("reboot", "sub1")

	// tamper with command
	sc.Command = "shutdown"
	if VerifyCommand(sc, svc.Pubkey) {
		t.Fatal("expected VerifyCommand to return false for tampered command")
	}

	// tamper with target
	sc2, _ := svc.SignCommand("reboot", "sub1")
	sc2.Target = "sub2"
	if VerifyCommand(sc2, svc.Pubkey) {
		t.Fatal("expected VerifyCommand to return false for tampered target")
	}
}


// TestVerifyCommandBadSignature handles the TestVerifyCommandBadSignature HTTP request.
func TestVerifyCommandBadSignature(t *testing.T) {
	sc := &SignedCommand{
		Command:   "test",
		Target:    "target",
		Nonce:     "abc",
		Timestamp: 12345,
		Signature: "nothex",
	}
	pub, _, _ := ed25519.GenerateKey(rand.Reader)
	if VerifyCommand(sc, pub) {
		t.Fatal("expected VerifyCommand to return false with bad signature hex")
	}

	sc.Signature = "aabb"
	if VerifyCommand(sc, pub) {
		t.Fatal("expected VerifyCommand to return false with short signature")
	}
}


// TestHeartbeat handles the TestHeartbeat HTTP request.
func TestHeartbeat(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	svc.RegisterNode("hbkey", "hbnode", "addr")

	// set old last seen directly
	oldTime := time.Now().Add(-1 * time.Hour).UTC().Format(time.RFC3339)
	svc.Nodes["hbkey"].LastSeen = oldTime

	time.Sleep(10 * time.Millisecond)

	svc.Heartbeat("hbkey")
	if svc.Nodes["hbkey"].LastSeen == oldTime {
		t.Fatal("expected last_seen to be updated")
	}
	if svc.Nodes["hbkey"].Status != StatusOnline {
		t.Fatalf("expected online status, got %s", svc.Nodes["hbkey"].Status)
	}
}


// TestHeartbeatUnknownNode handles the TestHeartbeatUnknownNode HTTP request.
func TestHeartbeatUnknownNode(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	// should not panic
	svc.Heartbeat("nonexistent")
}


// TestCheckStale handles the TestCheckStale HTTP request.
func TestCheckStale(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	svc.RegisterNode("fresh", "fresh-node", "addr")
	svc.RegisterNode("stale", "stale-node", "addr")

	// set stale node's last seen to way back
	svc.Nodes["stale"].LastSeen = time.Now().Add(-2 * StaleThreshold).UTC().Format(time.RFC3339)

	svc.CheckStale()

	if svc.Nodes["fresh"].Status != StatusOnline {
		t.Fatalf("expected fresh node to be online, got %s", svc.Nodes["fresh"].Status)
	}
	if svc.Nodes["stale"].Status != StatusStale {
		t.Fatalf("expected stale node to be stale, got %s", svc.Nodes["stale"].Status)
	}
}


// TestCheckStaleBadTimestamp handles the TestCheckStaleBadTimestamp HTTP request.
func TestCheckStaleBadTimestamp(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	svc.RegisterNode("bad", "bad-ts", "addr")
	svc.Nodes["bad"].LastSeen = "invalid-date"

	svc.CheckStale()
	if svc.Nodes["bad"].Status != StatusStale {
		t.Fatalf("expected stale status for bad ts, got %s", svc.Nodes["bad"].Status)
	}
}


// TestListNodes handles the TestListNodes HTTP request.
func TestListNodes(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	svc.RegisterNode("k1", "node1", "addr1")
	svc.RegisterNode("k2", "node2", "addr2")
	svc.RegisterNode("k3", "node3", "addr3")

	nodes := svc.ListNodes()
	if len(nodes) != 3 {
		t.Fatalf("expected 3 nodes, got %d", len(nodes))
	}

	labels := make(map[string]bool)
	for _, n := range nodes {
		labels[n.Label] = true
	}
	if !labels["node1"] || !labels["node2"] || !labels["node3"] {
		t.Fatal("list missing some nodes")
	}
}


// TestGetNode handles the TestGetNode HTTP request.
func TestGetNode(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	svc.RegisterNode("getkey", "getnode", "getaddr")

	node := svc.GetNode("getkey")
	if node == nil {
		t.Fatal("expected non-nil node")
	}
	if node.Label != "getnode" {
		t.Fatalf("bad label: %s", node.Label)
	}

	missing := svc.GetNode("nonexistent")
	if missing != nil {
		t.Fatal("expected nil for missing node")
	}
}


// TestNodeInfoJSON handles the TestNodeInfoJSON HTTP request.
func TestNodeInfoJSON(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)
	n := &NodeInfo{
		Pubkey:     "pk1",
		Label:      "test",
		Type:       NodeTypeSub,
		Addr:       "onion:abc.onion",
		Status:     StatusOnline,
		LastSeen:   now,
		Registered: now,
	}
	b, err := json.Marshal(n)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var n2 NodeInfo
	if err := json.Unmarshal(b, &n2); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if n2.Pubkey != "pk1" || n2.Label != "test" || n2.Type != NodeTypeSub {
		t.Fatal("json round-trip failed")
	}
}


// TestSignedCommandJSON handles the TestSignedCommandJSON HTTP request.
func TestSignedCommandJSON(t *testing.T) {
	sc := &SignedCommand{
		Command:   "upgrade",
		Target:    "subkey1",
		Nonce:     hex.EncodeToString([]byte("nonce1234")),
		Timestamp: 1234567890,
		Signature: hex.EncodeToString([]byte("signaturehere")),
	}
	b, err := json.Marshal(sc)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var sc2 SignedCommand
	if err := json.Unmarshal(b, &sc2); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if sc2.Command != "upgrade" || sc2.Target != "subkey1" {
		t.Fatal("json round-trip failed")
	}
}


// TestCommandReceiptJSON handles the TestCommandReceiptJSON HTTP request.
func TestCommandReceiptJSON(t *testing.T) {
	cr := &CommandReceipt{
		Command:   "upgrade",
		Status:    "done",
		Result:    "success",
		Nonce:     "abcd",
		Timestamp: 1234567890,
		Signature: "deadbeef",
	}
	b, err := json.Marshal(cr)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}
	var cr2 CommandReceipt
	if err := json.Unmarshal(b, &cr2); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if cr2.Command != "upgrade" || cr2.Status != "done" || cr2.Result != "success" {
		t.Fatal("json round-trip failed")
	}
}


// TestNodeConstants handles the TestNodeConstants HTTP request.
func TestNodeConstants(t *testing.T) {
	if NodeTypeRoyal != "royal" {
		t.Fatalf("bad NodeTypeRoyal: %s", NodeTypeRoyal)
	}
	if NodeTypeSub != "sub" {
		t.Fatalf("bad NodeTypeSub: %s", NodeTypeSub)
	}
	if StatusOnline != "online" {
		t.Fatalf("bad StatusOnline: %s", StatusOnline)
	}
	if StatusOffline != "offline" {
		t.Fatalf("bad StatusOffline: %s", StatusOffline)
	}
	if StatusStale != "stale" {
		t.Fatalf("bad StatusStale: %s", StatusStale)
	}
	if HeartbeatInterval != 5*time.Minute {
		t.Fatal("bad HeartbeatInterval")
	}
	if StaleThreshold != 15*time.Minute {
		t.Fatal("bad StaleThreshold")
	}
}

// --- HTTP handler tests ---

func localRequest(method, target, body string) *http.Request {
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	req.RemoteAddr = "127.0.0.1:12345"
	return req
}


// TestRegisterHandler handles the TestRegisterHandler HTTP request.
func TestRegisterHandler(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	h := RegisterHandler(svc)

	body := `{"pubkey":"handlertest","label":"handler-node","addr":"onion:test.onion"}`
	req := localRequest("POST", "/api/royal/register", body)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["ok"] != true {
		t.Fatal("expected ok=true in response")
	}
	nodeMap, ok := resp["node"].(map[string]any)
	if !ok {
		t.Fatal("expected node in response")
	}
	if nodeMap["pubkey"] != "handlertest" {
		t.Fatalf("bad pubkey in response: %v", nodeMap["pubkey"])
	}
}


// TestRegisterHandlerForbidden handles the TestRegisterHandlerForbidden HTTP request.
func TestRegisterHandlerForbidden(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	h := RegisterHandler(svc)

	// non-local remote addr should be forbidden
	req := httptest.NewRequest("POST", "/api/royal/register", strings.NewReader(`{"pubkey":"x","label":"y","addr":"z"}`))
	req.RemoteAddr = "203.0.113.1:9999"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}


// TestRegisterHandlerBadJSON handles the TestRegisterHandlerBadJSON HTTP request.
func TestRegisterHandlerBadJSON(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	h := RegisterHandler(svc)

	req := localRequest("POST", "/api/royal/register", `not json`)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}


// TestRegisterHandlerMissingPubkey handles the TestRegisterHandlerMissingPubkey HTTP request.
func TestRegisterHandlerMissingPubkey(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	h := RegisterHandler(svc)

	req := localRequest("POST", "/api/royal/register", `{"label":"x","addr":"y"}`)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}


// TestRegisterHandlerDuplicate handles the TestRegisterHandlerDuplicate HTTP request.
func TestRegisterHandlerDuplicate(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	h := RegisterHandler(svc)

	body := `{"pubkey":"dup","label":"first","addr":"addr1"}`
	req1 := localRequest("POST", "/api/royal/register", body)
	w1 := httptest.NewRecorder()
	h.ServeHTTP(w1, req1)
	if w1.Code != http.StatusOK {
		t.Fatalf("first register: expected 200, got %d", w1.Code)
	}

	req2 := localRequest("POST", "/api/royal/register", body)
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)
	if w2.Code != 409 {
		t.Fatalf("expected 409 for duplicate, got %d", w2.Code)
	}
}


// TestNodesHandler handles the TestNodesHandler HTTP request.
func TestNodesHandler(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	svc.RegisterNode("n1", "node1", "addr1")
	svc.RegisterNode("n2", "node2", "addr2")

	h := NodesHandler(svc)
	req := localRequest("GET", "/api/royal/nodes", "")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["royal_pubkey"] != svc.PublicKeyHex() {
		t.Fatal("royal_pubkey mismatch")
	}

	nodes, ok := resp["nodes"].([]any)
	if !ok {
		t.Fatal("expected nodes array")
	}
	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}
}


// TestNodesHandlerContactLink handles the TestNodesHandlerContactLink HTTP request.
func TestNodesHandlerContactLink(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)

	// write contact link file
	link := "simplex://contact/abc123"
	if err := writeFile(dir+"/island_contact_link.txt", link); err != nil {
		t.Fatalf("write contact link: %v", err)
	}

	h := NodesHandler(svc)
	req := localRequest("GET", "/api/royal/nodes", "")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["contact_link"] != link {
		t.Fatalf("expected contact_link=%q, got %q", link, resp["contact_link"])
	}
}


// TestCommandHandler handles the TestCommandHandler HTTP request.
func TestCommandHandler(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	h := CommandHandler(svc)

	body := `{"command":"restart","target":"subkey1"}`
	req := localRequest("POST", "/api/royal/command", body)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["ok"] != true {
		t.Fatal("expected ok=true")
	}

	sc, ok := resp["signed_command"].(map[string]any)
	if !ok {
		t.Fatal("expected signed_command in response")
	}
	if sc["command"] != "restart" {
		t.Fatalf("bad command: %v", sc["command"])
	}
	if sc["target"] != "subkey1" {
		t.Fatalf("bad target: %v", sc["target"])
	}
	if sc["signature"] == "" {
		t.Fatal("expected signature")
	}
}


// TestCommandHandlerForbidden handles the TestCommandHandlerForbidden HTTP request.
func TestCommandHandlerForbidden(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	h := CommandHandler(svc)

	req := httptest.NewRequest("POST", "/api/royal/command", strings.NewReader(`{"command":"x","target":"y"}`))
	req.RemoteAddr = "10.0.0.1:9999"
	// 10.x is private so this will actually pass due to isPrivateIP...
	// Use a public IP instead
	req.RemoteAddr = "203.0.113.1:9999"
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}


// TestCommandHandlerMissingFields handles the TestCommandHandlerMissingFields HTTP request.
func TestCommandHandlerMissingFields(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	h := CommandHandler(svc)

	// missing target
	req := localRequest("POST", "/api/royal/command", `{"command":"x"}`)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}

	// missing command
	req2 := localRequest("POST", "/api/royal/command", `{"target":"x"}`)
	w2 := httptest.NewRecorder()
	h.ServeHTTP(w2, req2)
	if w2.Code != 400 {
		t.Fatalf("expected 400, got %d", w2.Code)
	}
}


// TestCommandHandlerBadJSON handles the TestCommandHandlerBadJSON HTTP request.
func TestCommandHandlerBadJSON(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	h := CommandHandler(svc)

	req := localRequest("POST", "/api/royal/command", `not json`)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}


// TestHeartbeatHandler handles the TestHeartbeatHandler HTTP request.
func TestHeartbeatHandler(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	svc.RegisterNode("hbhandler", "hbnode", "addr")

	h := HeartbeatHandler(svc)
	req := localRequest("POST", "/api/royal/heartbeat?pubkey=hbhandler", "")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["ok"] != true {
		t.Fatal("expected ok=true")
	}
	if resp["pubkey"] != "hbhandler" {
		t.Fatalf("bad pubkey: %v", resp["pubkey"])
	}
}


// TestHeartbeatHandlerMissingPubkey handles the TestHeartbeatHandlerMissingPubkey HTTP request.
func TestHeartbeatHandlerMissingPubkey(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	h := HeartbeatHandler(svc)

	req := localRequest("POST", "/api/royal/heartbeat", "")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}


// TestKeyHandler handles the TestKeyHandler HTTP request.
func TestKeyHandler(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(dir)
	h := KeyHandler(svc)

	req := localRequest("GET", "/api/royal/key", "")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["pubkey"] != svc.PublicKeyHex() {
		t.Fatalf("pubkey mismatch: %v vs %s", resp["pubkey"], svc.PublicKeyHex())
	}
}

// writeFile is a small helper to write a string to a file for testing.
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
