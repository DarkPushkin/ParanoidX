// Package radio provides the radio streaming and scheduling system
package radio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)


// TestPlaylistBuilderEmptyStation handles the TestPlaylistBuilderEmptyStation HTTP request.
func TestPlaylistBuilderEmptyStation(t *testing.T) {
	dir := t.TempDir()
	as := NewAnnouncementStore(dir)
	pb := NewPlaylistBuilder(dir, as)

	tracks := pb.Build("nonexistent-station")
	if len(tracks) == 0 {
		t.Fatal("expected at least a placeholder track")
	}
	if !tracks[0].IsAnnounce {
		t.Fatal("expected placeholder to be marked as announcement")
	}
}


// TestPlaylistBuilderWithMusic handles the TestPlaylistBuilderWithMusic HTTP request.
func TestPlaylistBuilderWithMusic(t *testing.T) {
	dir := t.TempDir()
	as := NewAnnouncementStore(dir)
	pb := NewPlaylistBuilder(dir, as)

	stationDir := filepath.Join(dir, "radio", "stations", "test-station")
	os.MkdirAll(stationDir, 0755)
	os.WriteFile(filepath.Join(stationDir, "track1.mp3"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(stationDir, "track2.ogg"), []byte("data"), 0644)
	os.WriteFile(filepath.Join(stationDir, "readme.txt"), []byte("not audio"), 0644)

	tracks := pb.Build("test-station")
	if len(tracks) != 2 {
		t.Fatalf("expected 2 audio tracks, got %d", len(tracks))
	}
	if tracks[0].ID != "test-station-track1.mp3" {
		t.Fatalf("expected first track ID 'test-station-track1.mp3', got '%s'", tracks[0].ID)
	}
}


// TestPlaylistBuilderInterleavesAnnouncements handles the TestPlaylistBuilderInterleavesAnnouncements HTTP request.
func TestPlaylistBuilderInterleavesAnnouncements(t *testing.T) {
	dir := t.TempDir()
	as := NewAnnouncementStore(dir)
	pb := NewPlaylistBuilder(dir, as)

	stationDir := filepath.Join(dir, "radio", "stations", "test-station")
	os.MkdirAll(stationDir, 0755)
	for i := 0; i < 5; i++ {
		os.WriteFile(filepath.Join(stationDir, fmt.Sprintf("track%d.mp3", i)), []byte("data"), 0644)
	}

	as.Add(Announcement{
		ID: "ann-1", Announcer: AnnouncerKing,
		Title: "Test Announce", Body: "body",
		Stations: []string{"test-station"},
	})

	tracks := pb.Build("test-station")
	// Should have 5 music tracks + 1 announcement = 6 total
	if len(tracks) != 6 {
		t.Fatalf("expected 6 tracks (5 music + 1 announcement), got %d", len(tracks))
	}

	// Announcement should be inserted after the 4th track (at index 4)
	if !tracks[4].IsAnnounce {
		t.Fatal("expected announcement at index 4")
	}
}


// TestCurrentPlaylistJSON handles the TestCurrentPlaylistJSON HTTP request.
func TestCurrentPlaylistJSON(t *testing.T) {
	dir := t.TempDir()
	as := NewAnnouncementStore(dir)
	pb := NewPlaylistBuilder(dir, as)

	stationDir := filepath.Join(dir, "radio", "stations", "test-json")
	os.MkdirAll(stationDir, 0755)
	os.WriteFile(filepath.Join(stationDir, "song.mp3"), []byte("data"), 0644)

	result := pb.CurrentPlaylistJSON("test-json")
	if len(result) != 1 {
		t.Fatalf("expected 1 track in JSON, got %d", len(result))
	}
	if result[0]["id"] != "test-json-song.mp3" {
		t.Fatalf("unexpected id: %v", result[0]["id"])
	}
	url, ok := result[0]["stream_url"]
	if !ok {
		t.Fatal("expected stream_url in JSON output")
	}
	urlStr := url.(string)
	if !strings.Contains(urlStr, "/api/radio/track") {
		t.Fatalf("expected stream_url to contain /api/radio/track, got %s", urlStr)
	}
}


// TestPlaylistBuilderFiltersAnnouncementsByStation handles the TestPlaylistBuilderFiltersAnnouncementsByStation HTTP request.
func TestPlaylistBuilderFiltersAnnouncementsByStation(t *testing.T) {
	dir := t.TempDir()
	as := NewAnnouncementStore(dir)
	pb := NewPlaylistBuilder(dir, as)

	stationDir := filepath.Join(dir, "radio", "stations", "station-a")
	os.MkdirAll(stationDir, 0755)
	os.WriteFile(filepath.Join(stationDir, "track.mp3"), []byte("data"), 0644)

	as.Add(Announcement{
		ID: "ann-for-b", Announcer: AnnouncerSteward,
		Title: "Only B", Body: "body",
		Stations: []string{"station-b"},
	})

	tracks := pb.Build("station-a")
	for _, tr := range tracks {
		if tr.ID == "ann-ann-for-b" {
			t.Fatal("announcement for station-b leaked into station-a")
		}
	}
}
