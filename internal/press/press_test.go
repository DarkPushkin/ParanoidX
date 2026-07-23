// Package press provides press release and announcement template management
package press

import (
	"os"
	"path/filepath"
	"testing"
)


// TestNewTemplateManager handles the TestNewTemplateManager HTTP request.
func TestNewTemplateManager(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "banknote-templates"), 0700)

	tm := NewTemplateManager(dir)
	if tm == nil {
		t.Fatal("expected non-nil manager")
	}
	if tm.Manifest == nil {
		t.Fatal("expected manifest loaded")
	}
	if len(tm.Manifest.Denominations) == 0 {
		t.Fatal("expected denominations")
	}
	if len(tm.Manifest.Rarities) == 0 {
		t.Fatal("expected rarities")
	}
}


// TestNewTemplateManagerCreatesDir handles the TestNewTemplateManagerCreatesDir HTTP request.
func TestNewTemplateManagerCreatesDir(t *testing.T) {
	dir := t.TempDir()
	tm := NewTemplateManager(dir)
	if tm.Manifest == nil {
		t.Fatal("expected manifest even without existing templates dir")
	}
}


// TestValidateSerialValid handles the TestValidateSerialValid HTTP request.
func TestValidateSerialValid(t *testing.T) {
	if !ValidateSerial("MB-rare-202505-001") {
		t.Fatal("expected valid serial")
	}
}


// TestValidateSerialInvalid handles the TestValidateSerialInvalid HTTP request.
func TestValidateSerialInvalid(t *testing.T) {
	if ValidateSerial("") {
		t.Fatal("expected invalid empty serial")
	}
	if ValidateSerial("ABC-123") {
		t.Fatal("expected invalid serial without MB prefix")
	}
}


// TestSaveAndLoadManifest handles the TestSaveAndLoadManifest HTTP request.
func TestSaveAndLoadManifest(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "banknote-templates"), 0700)

	tm := NewTemplateManager(dir)
	tm.Manifest.KingPubkey = "custom_pubkey"
	tm.SaveManifest()

	tm2 := NewTemplateManager(dir)
	if tm2.Manifest.KingPubkey != "custom_pubkey" {
		t.Fatalf("expected custom_pubkey, got %s", tm2.Manifest.KingPubkey)
	}
}


// TestListTemplates handles the TestListTemplates HTTP request.
func TestListTemplates(t *testing.T) {
	dir := t.TempDir()
	base := filepath.Join(dir, "banknote-templates")
	os.MkdirAll(filepath.Join(base, "100"), 0700)
	os.WriteFile(filepath.Join(base, "100", "common-front.svg"), []byte("<svg/>"), 0600)

	tm := NewTemplateManager(dir)
	templates := tm.ListTemplates()
	if len(templates) == 0 {
		t.Fatal("expected at least one template")
	}
}


// TestSetKingKey handles the TestSetKingKey HTTP request.
func TestSetKingKey(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "banknote-templates"), 0700)

	tm := NewTemplateManager(dir)
	tm.SetKingKey("abcdef1234567890")

	tm2 := NewTemplateManager(dir)
	if tm2.Manifest.KingPubkey != "abcdef1234567890" {
		t.Fatalf("expected preserved king key, got %s", tm2.Manifest.KingPubkey)
	}
}


// TestRenderBanknoteGeneratesFiles handles the TestRenderBanknoteGeneratesFiles HTTP request.
func TestRenderBanknoteGeneratesFiles(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "banknote-templates"), 0700)
	outputDir := filepath.Join(dir, "output")
	os.MkdirAll(outputDir, 0700)

	tm := NewTemplateManager(dir)
	pdfPath, sigPath, err := tm.RenderBanknote(100, "common", "MB-common-202506-001", outputDir)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := os.Stat(pdfPath); err != nil {
		t.Fatalf("expected pdf file: %v", err)
	}
	if _, err := os.Stat(sigPath); err != nil {
		t.Fatalf("expected sig file: %v", err)
	}
}


// TestVerifySignature handles the TestVerifySignature HTTP request.
func TestVerifySignature(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "banknote-templates"), 0700)
	outputDir := filepath.Join(dir, "output")
	os.MkdirAll(outputDir, 0700)

	tm := NewTemplateManager(dir)
	tm.SetKingKey("deadbeefdeadbeefdeadbeefdeadbeef")
	pdfPath, sigPath, err := tm.RenderBanknote(500, "rare", "MB-rare-202506-002", outputDir)
	if err != nil {
		t.Fatal(err)
	}

	valid, err := tm.VerifySignature(pdfPath, sigPath)
	if err != nil {
		t.Fatal(err)
	}
	if !valid {
		t.Fatal("expected valid signature (SHA256 stub)")
	}
}


// TestVerifySignatureKingKeyNotSet handles the TestVerifySignatureKingKeyNotSet HTTP request.
func TestVerifySignatureKingKeyNotSet(t *testing.T) {
	dir := t.TempDir()
	outputDir := filepath.Join(dir, "output")
	os.MkdirAll(outputDir, 0700)

	tm := NewTemplateManager(dir)
	tm.Manifest.KingPubkey = ""
	pdfPath, sigPath, err := tm.RenderBanknote(100, "common", "MB-common-202506-003", outputDir)
	if err != nil {
		t.Fatal(err)
	}

	_, err = tm.VerifySignature(pdfPath, sigPath)
	if err == nil {
		t.Fatal("expected error when king key not set")
	}
}


// TestRenderBanknoteInOutputDir handles the TestRenderBanknoteInOutputDir HTTP request.
func TestRenderBanknoteInOutputDir(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "banknote-templates"), 0700)
	outputDir := filepath.Join(dir, "banknotes")
	os.MkdirAll(outputDir, 0700)

	tm := NewTemplateManager(dir)
	pdfPath, sigPath, err := tm.RenderBanknote(1000, "legendary", "MB-legendary-202506-099", outputDir)
	if err != nil {
		t.Fatal(err)
	}

	files, _ := os.ReadDir(outputDir)
	found := 0
	for _, f := range files {
		if !f.IsDir() {
			found++
		}
	}
	if found < 2 {
		t.Fatalf("expected at least 2 files in output, got %d", found)
	}
	_ = pdfPath
	_ = sigPath
}
