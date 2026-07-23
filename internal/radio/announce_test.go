// Package radio provides the radio streaming and scheduling system
package radio

import (
	"os"
	"testing"
	"time"
)


// TestNewAnnouncementStore handles the TestNewAnnouncementStore HTTP request.
func TestNewAnnouncementStore(t *testing.T) {
	dir := t.TempDir()
	as := NewAnnouncementStore(dir)
	if as == nil {
		t.Fatal("expected non-nil AnnouncementStore")
	}
	if _, err := os.Stat(as.DataDir); err != nil {
		t.Fatalf("expected data dir to exist: %v", err)
	}
}


// TestAnnouncementStoreAddAndPending handles the TestAnnouncementStoreAddAndPending HTTP request.
func TestAnnouncementStoreAddAndPending(t *testing.T) {
	dir := t.TempDir()
	as := NewAnnouncementStore(dir)

	a := Announcement{
		ID: "test-1", Announcer: AnnouncerKing,
		Title: "Royal Decree", Body: "Test announcement",
		Lang: LangEN, Priority: PriorityHigh,
		CreatedAt: time.Now(),
	}
	as.Add(a)

	pending := as.Pending()
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending, got %d", len(pending))
	}
	if pending[0].Title != "Royal Decree" {
		t.Fatalf("expected 'Royal Decree', got '%s'", pending[0].Title)
	}
}


// TestAnnouncementStoreMarkPlayed handles the TestAnnouncementStoreMarkPlayed HTTP request.
func TestAnnouncementStoreMarkPlayed(t *testing.T) {
	dir := t.TempDir()
	as := NewAnnouncementStore(dir)

	as.Add(Announcement{
		ID: "play-test", Announcer: AnnouncerSteward,
		Title: "Play Me", Body: "test",
		CreatedAt: time.Now(),
	})

	if len(as.Pending()) != 1 {
		t.Fatal("expected 1 pending before mark")
	}

	as.MarkPlayed("play-test")
	if len(as.Pending()) != 0 {
		t.Fatal("expected 0 pending after mark")
	}
}


// TestAnnouncementStoreHistory handles the TestAnnouncementStoreHistory HTTP request.
func TestAnnouncementStoreHistory(t *testing.T) {
	dir := t.TempDir()
	as := NewAnnouncementStore(dir)

	for i := 0; i < 5; i++ {
		as.Add(Announcement{
			ID: "h-test", Announcer: AnnouncerKing,
			Title: "History", Body: "test",
			CreatedAt: time.Now(),
		})
	}

	hist := as.History(3)
	if len(hist) != 3 {
		t.Fatalf("expected 3 history items, got %d", len(hist))
	}
}


// TestAnnouncementStorePersistence handles the TestAnnouncementStorePersistence HTTP request.
func TestAnnouncementStorePersistence(t *testing.T) {
	dir := t.TempDir()
	as := NewAnnouncementStore(dir)
	as.Add(Announcement{
		ID: "persist-ann", Announcer: AnnouncerKing,
		Title: "Persisted", Body: "test persistence",
		CreatedAt: time.Now(),
	})

	as2 := NewAnnouncementStore(dir)
	pending := as2.Pending()
	if len(pending) != 1 {
		t.Fatalf("expected 1 pending in reloaded store, got %d", len(pending))
	}
	if pending[0].Body != "test persistence" {
		t.Fatalf("expected 'test persistence', got '%s'", pending[0].Body)
	}
}


// TestAnnouncementScheduledNotPending handles the TestAnnouncementScheduledNotPending HTTP request.
func TestAnnouncementScheduledNotPending(t *testing.T) {
	dir := t.TempDir()
	as := NewAnnouncementStore(dir)
	future := time.Now().Add(24 * time.Hour)
	as.Add(Announcement{
		ID: "future-ann", Announcer: AnnouncerTorquemada,
		Title: "Future", Body: "not yet", Priority: PriorityLow,
		ScheduledAt: &future, CreatedAt: time.Now(),
	})

	pending := as.Pending()
	if len(pending) != 0 {
		t.Fatal("expected 0 pending for future-scheduled announcement")
	}
}


// TestAnnouncerLabel handles the TestAnnouncerLabel HTTP request.
func TestAnnouncerLabel(t *testing.T) {
	tests := []struct {
		a    Announcer
		want string
	}{
		{AnnouncerKing, "👑 Король"},
		{AnnouncerTorquemada, "🔧 Торквемада"},
		{AnnouncerSteward, "🤖 Стюард"},
		{Announcer("unknown"), "unknown"},
	}
	for _, tc := range tests {
		got := announcerLabel(tc.a)
		if got != tc.want {
			t.Errorf("announcerLabel(%q) = %q, want %q", tc.a, got, tc.want)
		}
	}
}
