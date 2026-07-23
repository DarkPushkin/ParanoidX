// Package channels provides SimpleX channel management
package channels

import (
	"os"
	"testing"
)


// TestNewManager handles the TestNewManager HTTP request.
func TestNewManager(t *testing.T) {
	dir, _ := os.MkdirTemp("", "channels-test")
	defer os.RemoveAll(dir)
	m := NewManager(dir)
	if m == nil {
		t.Fatal("expected manager")
	}
	if len(m.ListChannels()) != 0 {
		t.Fatal("expected empty")
	}
}


// TestAddAndList handles the TestAddAndList HTTP request.
func TestAddAndList(t *testing.T) {
	dir, _ := os.MkdirTemp("", "channels-test")
	defer os.RemoveAll(dir)
	m := NewManager(dir)
	m.AddChannel("ch1", "general", "https://simplex.chat/...", "creator")
	m.AddChannel("ch2", "random", "https://simplex.chat/...", "subscriber")
	list := m.ListChannels()
	if len(list) != 2 {
		t.Fatalf("expected 2, got %d", len(list))
	}
}


// TestGetChannel handles the TestGetChannel HTTP request.
func TestGetChannel(t *testing.T) {
	dir, _ := os.MkdirTemp("", "channels-test")
	defer os.RemoveAll(dir)
	m := NewManager(dir)
	m.AddChannel("ch1", "general", "", "creator")
	ch := m.GetChannel("ch1")
	if ch == nil || ch.Name != "general" {
		t.Fatal("expected channel")
	}
	if m.GetChannel("nonexistent") != nil {
		t.Fatal("expected nil")
	}
}


// TestAddMessage handles the TestAddMessage HTTP request.
func TestAddMessage(t *testing.T) {
	dir, _ := os.MkdirTemp("", "channels-test")
	defer os.RemoveAll(dir)
	m := NewManager(dir)
	m.AddChannel("ch1", "general", "", "creator")
	m.AddMessage("ch1", "hello world", "alice")
	msgs := m.GetMessages("ch1", 0)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Text != "hello world" {
		t.Fatalf("expected 'hello world', got '%s'", msgs[0].Text)
	}
}


// TestMarkRead handles the TestMarkRead HTTP request.
func TestMarkRead(t *testing.T) {
	dir, _ := os.MkdirTemp("", "channels-test")
	defer os.RemoveAll(dir)
	m := NewManager(dir)
	m.AddChannel("ch1", "general", "", "creator")
	m.AddMessage("ch1", "msg1", "alice")
	m.AddMessage("ch1", "msg2", "bob")
	ch := m.GetChannel("ch1")
	if ch.Unread != 2 {
		t.Fatalf("expected 2 unread, got %d", ch.Unread)
	}
	m.MarkRead("ch1")
	if ch.Unread != 0 {
		t.Fatalf("expected 0 unread after mark read, got %d", ch.Unread)
	}
}


// TestRemoveChannel handles the TestRemoveChannel HTTP request.
func TestRemoveChannel(t *testing.T) {
	dir, _ := os.MkdirTemp("", "channels-test")
	defer os.RemoveAll(dir)
	m := NewManager(dir)
	m.AddChannel("ch1", "general", "", "creator")
	m.RemoveChannel("ch1")
	if m.GetChannel("ch1") != nil {
		t.Fatal("expected nil after remove")
	}
}


// TestSaveAndLoad handles the TestSaveAndLoad HTTP request.
func TestSaveAndLoad(t *testing.T) {
	dir, _ := os.MkdirTemp("", "channels-test")
	defer os.RemoveAll(dir)
	m := NewManager(dir)
	m.AddChannel("ch1", "general", "", "creator")
	m.AddMessage("ch1", "persisted message", "system")

	m2 := NewManager(dir)
	if len(m2.ListChannels()) != 1 {
		t.Fatal("expected 1 channel after reload")
	}
	msgs := m2.GetMessages("ch1", 0)
	if len(msgs) != 1 || msgs[0].Text != "persisted message" {
		t.Fatal("expected persisted message after reload")
	}
}


// TestGetMessagesLimit handles the TestGetMessagesLimit HTTP request.
func TestGetMessagesLimit(t *testing.T) {
	dir, _ := os.MkdirTemp("", "channels-test")
	defer os.RemoveAll(dir)
	m := NewManager(dir)
	m.AddChannel("ch1", "general", "", "creator")
	m.AddMessage("ch1", "msg1", "a")
	m.AddMessage("ch1", "msg2", "b")
	m.AddMessage("ch1", "msg3", "c")
	msgs := m.GetMessages("ch1", 2)
	if len(msgs) != 2 {
		t.Fatalf("expected 2, got %d", len(msgs))
	}
	if msgs[0].Text != "msg2" || msgs[1].Text != "msg3" {
		t.Fatal("expected last 2 messages")
	}
}
