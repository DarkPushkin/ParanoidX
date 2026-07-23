// Package radio provides the radio streaming and scheduling system
package radio

import (
	"os"
	"testing"
	"time"
)


// TestNewRadioService handles the TestNewRadioService HTTP request.
func TestNewRadioService(t *testing.T) {
	dir := t.TempDir()
	rs := NewRadioService(dir)
	if rs.DataDir == "" {
		t.Fatal("expected non-empty DataDir")
	}
	if _, err := os.Stat(rs.DataDir); err != nil {
		t.Fatalf("expected radio dir to exist: %v", err)
	}
}


// TestRadioServiceSeedDefaults handles the TestRadioServiceSeedDefaults HTTP request.
func TestRadioServiceSeedDefaults(t *testing.T) {
	dir := t.TempDir()
	rs := NewRadioService(dir)
	stations := rs.List()
	if len(stations) == 0 {
		t.Fatal("expected seeded stations")
	}
	// Verify one of the expected defaults
	found := false
	for _, s := range stations {
		if s.ID == "liberty-voice-en" {
			found = true
			if !s.Enabled {
				t.Fatal("expected liberty-voice-en to be enabled")
			}
			break
		}
	}
	if !found {
		t.Fatal("expected liberty-voice-en station")
	}
}


// TestRadioServiceGet handles the TestRadioServiceGet HTTP request.
func TestRadioServiceGet(t *testing.T) {
	dir := t.TempDir()
	rs := NewRadioService(dir)
	s := rs.Get("liberty-voice-en")
	if s == nil {
		t.Fatal("expected to find liberty-voice-en")
	}
	if s.Name != "Liberty Voice" {
		t.Fatalf("expected 'Liberty Voice', got '%s'", s.Name)
	}
	missing := rs.Get("nonexistent")
	if missing != nil {
		t.Fatal("expected nil for nonexistent station")
	}
}


// TestRadioServiceSave handles the TestRadioServiceSave HTTP request.
func TestRadioServiceSave(t *testing.T) {
	dir := t.TempDir()
	rs := NewRadioService(dir)
	s := &RadioStation{
		ID: "test-station", Name: "Test Station", Type: StationMusic,
		Lang: LangEN, Description: "test", Enabled: true, CreatedAt: time.Now(),
	}
	rs.Save(s)
	got := rs.Get("test-station")
	if got == nil {
		t.Fatal("expected to get saved station")
	}
	if got.Name != "Test Station" {
		t.Fatalf("expected 'Test Station', got '%s'", got.Name)
	}
}


// TestRadioServicePersistence handles the TestRadioServicePersistence HTTP request.
func TestRadioServicePersistence(t *testing.T) {
	dir := t.TempDir()
	rs := NewRadioService(dir)
	rs.Save(&RadioStation{
		ID: "persist-test", Name: "Persist Test", Type: StationNews,
		Lang: LangRU, Description: "persistence test", Enabled: true, CreatedAt: time.Now(),
	})
	// Create a new service instance loading the same directory
	rs2 := NewRadioService(dir)
	got := rs2.Get("persist-test")
	if got == nil {
		t.Fatal("expected persisted station to load")
	}
	if got.Name != "Persist Test" {
		t.Fatalf("expected 'Persist Test', got '%s'", got.Name)
	}
}


// TestStationLang handles the TestStationLang HTTP request.
func TestStationLang(t *testing.T) {
	tests := []struct {
		id   string
		want Lang
	}{
		{"liberty-voice-en", LangEN},
		{"liberty-voice-ru", LangRU},
		{"liberty-voice-es", LangES},
		{"torquemada-monitor", LangEN},
		{"unknown-station", LangEN},
	}
	for _, tc := range tests {
		got := stationLang(tc.id)
		if got != tc.want {
			t.Errorf("stationLang(%q) = %q, want %q", tc.id, got, tc.want)
		}
	}
}


// TestIsAudioFile handles the TestIsAudioFile HTTP request.
func TestIsAudioFile(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"track.mp3", true},
		{"track.ogg", true},
		{"track.wav", true},
		{"track.flac", true},
		{"track.aac", true},
		{"track.m4a", true},
		{"track.webm", true},
		{"track.txt", false},
		{"track.mp4", false},
		{"track", false},
	}
	for _, tc := range tests {
		got := isAudioFile(tc.name)
		if got != tc.want {
			t.Errorf("isAudioFile(%q) = %v, want %v", tc.name, got, tc.want)
		}
	}
}


// TestTrackTitle handles the TestTrackTitle HTTP request.
func TestTrackTitle(t *testing.T) {
	tests := []struct {
		name string
		want string
	}{
		{"hello-world.mp3", "hello world"},
		{"announcement-001.ogg", "announcement 001"},
		{"track.flac", "track"},
	}
	for _, tc := range tests {
		got := trackTitle(tc.name)
		if got != tc.want {
			t.Errorf("trackTitle(%q) = %q, want %q", tc.name, got, tc.want)
		}
	}
}


// TestContains handles the TestContains HTTP request.
func TestContains(t *testing.T) {
	slice := []string{"a", "b", "c"}
	if !contains(slice, "a") {
		t.Error("expected contains(a) = true")
	}
	if contains(slice, "z") {
		t.Error("expected contains(z) = false")
	}
}


// TestSortTracksByAdded handles the TestSortTracksByAdded HTTP request.
func TestSortTracksByAdded(t *testing.T) {
	now := time.Now()
	tracks := []Track{
		{ID: "old", AddedAt: now.Add(-time.Hour)},
		{ID: "new", AddedAt: now},
		{ID: "mid", AddedAt: now.Add(-30 * time.Minute)},
	}
	SortTracksByAdded(tracks)
	if tracks[0].ID != "new" {
		t.Fatalf("expected newest first, got %s", tracks[0].ID)
	}
	if tracks[2].ID != "old" {
		t.Fatalf("expected oldest last, got %s", tracks[2].ID)
	}
}
