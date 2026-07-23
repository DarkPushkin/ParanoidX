// Package radio provides the radio streaming and scheduling system
package radio

import (
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"
)

type RotationMode string

const (
	RotationSequential RotationMode = "sequential"
	RotationShuffle    RotationMode = "shuffle"
	RotationWeighted   RotationMode = "weighted"
)

type TimeSlotSched struct {
	Name    string `json:"name"`    // slot name
	Content string `json:"content"` // content type to play at this slot
	Minute  int    `json:"minute"`  // minute of hour (0-59)
	Enabled bool   `json:"enabled"`
}

type ContentScheduleEntry struct {
	ID        string `json:"id"`
	Type      string `json:"type"`       // ad, announce, music, news
	CronMin   int    `json:"cron_min"`   // minute interval (every N minutes)
	Prompt    string `json:"prompt"`
	Priority  int    `json:"priority"`   // higher = plays first
	LastRun   string `json:"last_run,omitempty"`
	NextRun   string `json:"next_run,omitempty"`
	CreatedAt string `json:"created_at"`
	Enabled   bool   `json:"enabled"`
	Weight    int    `json:"weight,omitempty"` // for weighted rotation
}

type EnhancedScheduler struct {
	mu             sync.RWMutex
	entries        []ContentScheduleEntry
	filePath       string
	ticker         *time.Ticker
	stopCh         chan struct{}
	onTick         func(entry ContentScheduleEntry)
	rotationMode   RotationMode
	timeSlots      []TimeSlotSched
	currentIndex   int
	rotationOrder  []int // shuffled index for shuffle mode
}


// NewEnhancedScheduler handles the NewEnhancedScheduler HTTP request.
func NewEnhancedScheduler(dataDir string) *EnhancedScheduler {
	es := &EnhancedScheduler{
		filePath:     filepath.Join(dataDir, "radio", "radio_schedule.json"),
		stopCh:       make(chan struct{}),
		rotationMode: RotationSequential,
		timeSlots: []TimeSlotSched{
			{Name: "news", Content: "news", Minute: 0, Enabled: true},
			{Name: "music", Content: "music", Minute: 15, Enabled: true},
			{Name: "ads", Content: "ad", Minute: 30, Enabled: true},
			{Name: "announce", Content: "announce", Minute: 45, Enabled: true},
		},
	}
	os.MkdirAll(filepath.Dir(es.filePath), 0755)
	es.load()
	return es
}


// SetOnTick handles the SetOnTick HTTP request.
func (es *EnhancedScheduler) SetOnTick(fn func(entry ContentScheduleEntry)) {
	es.mu.Lock()
	es.onTick = fn
	es.mu.Unlock()
}


// Start handles the Start HTTP request.
func (es *EnhancedScheduler) Start() {
	es.ticker = time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-es.ticker.C:
				es.processTick()
			case <-es.stopCh:
				return
			}
		}
	}()
}


// Stop handles the Stop HTTP request.
func (es *EnhancedScheduler) Stop() {
	if es.ticker != nil {
		es.ticker.Stop()
	}
	close(es.stopCh)
}

func (es *EnhancedScheduler) processTick() {
	es.mu.Lock()
	entries := make([]ContentScheduleEntry, len(es.entries))
	copy(entries, es.entries)
	onTick := es.onTick
	es.mu.Unlock()

	now := time.Now()
	changed := false

	// Check time-of-day slots
	for _, slot := range es.timeSlots {
		if !slot.Enabled {
			continue
		}
		if now.Minute() == slot.Minute {
			log.Printf("[radio scheduler] time slot: %s (content: %s)", slot.Name, slot.Content)
		}
	}

	for i, e := range entries {
		if !e.Enabled || e.CronMin <= 0 {
			continue
		}
		run := true
		if e.LastRun != "" {
			last, err := time.Parse(time.RFC3339, e.LastRun)
			if err == nil {
				if now.Sub(last) < time.Duration(e.CronMin)*time.Minute {
					run = false
				}
			}
		}
		if run && onTick != nil {
			onTick(e)
			entries[i].LastRun = now.Format(time.RFC3339)
			entries[i].NextRun = now.Add(time.Duration(e.CronMin) * time.Minute).Format(time.RFC3339)
			changed = true
		}
	}

	if changed {
		es.mu.Lock()
		es.entries = entries
		es.save()
		es.mu.Unlock()
	}
}


// GetRotationOrder handles the GetRotationOrder HTTP request.
func (es *EnhancedScheduler) GetRotationOrder() []ContentScheduleEntry {
	es.mu.RLock()
	defer es.mu.RUnlock()

	entries := make([]ContentScheduleEntry, len(es.entries))
	copy(entries, es.entries)

	// Filter enabled only
	enabled := make([]ContentScheduleEntry, 0)
	for _, e := range entries {
		if e.Enabled {
			enabled = append(enabled, e)
		}
	}

	// Sort by priority (higher first)
	sort.Slice(enabled, func(i, j int) bool {
		return enabled[i].Priority > enabled[j].Priority
	})

	switch es.rotationMode {
	case RotationSequential:
		es.currentIndex = (es.currentIndex + 1) % len(enabled)
	case RotationShuffle:
		rand.Shuffle(len(enabled), func(i, j int) {
			enabled[i], enabled[j] = enabled[j], enabled[i]
		})
	case RotationWeighted:
		totalWeight := 0
		for _, e := range enabled {
			w := e.Weight
			if w <= 0 {
				w = 1
			}
			totalWeight += w
		}
		r := rand.Intn(totalWeight)
		cumulative := 0
		for i, e := range enabled {
			w := e.Weight
			if w <= 0 {
				w = 1
			}
			cumulative += w
			if r < cumulative {
				enabled[0], enabled[i] = enabled[i], enabled[0]
				break
			}
		}
	}

	return enabled
}


// Optimize handles the Optimize HTTP request.
func (es *EnhancedScheduler) Optimize() {
	es.mu.Lock()
	defer es.mu.Unlock()

	// Reorder entries: higher priority first, then shuffle within same priority
	sort.Slice(es.entries, func(i, j int) bool {
		if es.entries[i].Priority != es.entries[j].Priority {
			return es.entries[i].Priority > es.entries[j].Priority
		}
		// Alternate types for variety
		return es.entries[i].Type < es.entries[j].Type
	})

	// Stagger run times
	now := time.Now()
	for i := range es.entries {
		if es.entries[i].Enabled && es.entries[i].CronMin > 0 {
			stagger := i * 2
			es.entries[i].NextRun = now.Add(time.Duration(stagger) * time.Minute).Format(time.RFC3339)
		}
	}
	es.save()
	log.Printf("[radio scheduler] optimized %d entries", len(es.entries))
}


// Stats handles the Stats HTTP request.
func (es *EnhancedScheduler) Stats() map[string]int {
	es.mu.RLock()
	defer es.mu.RUnlock()

	stats := map[string]int{}
	for _, e := range es.entries {
		stats[e.Type]++
	}
	stats["total"] = len(es.entries)

	enabled := 0
	for _, e := range es.entries {
		if e.Enabled {
			enabled++
		}
	}
	stats["enabled"] = enabled
	return stats
}


// SetRotationMode handles the SetRotationMode HTTP request.
func (es *EnhancedScheduler) SetRotationMode(mode RotationMode) {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.rotationMode = mode
	es.save()
}


// GetRotationMode handles the GetRotationMode HTTP request.
func (es *EnhancedScheduler) GetRotationMode() RotationMode {
	es.mu.RLock()
	defer es.mu.RUnlock()
	return es.rotationMode
}


// SetTimeSlots handles the SetTimeSlots HTTP request.
func (es *EnhancedScheduler) SetTimeSlots(slots []TimeSlotSched) {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.timeSlots = slots
	es.save()
}


// GetTimeSlots handles the GetTimeSlots HTTP request.
func (es *EnhancedScheduler) GetTimeSlots() []TimeSlotSched {
	es.mu.RLock()
	defer es.mu.RUnlock()
	out := make([]TimeSlotSched, len(es.timeSlots))
	copy(out, es.timeSlots)
	return out
}


// Add handles the Add HTTP request.
func (es *EnhancedScheduler) Add(entry ContentScheduleEntry) {
	es.mu.Lock()
	entry.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	entry.NextRun = time.Now().Add(time.Duration(entry.CronMin) * time.Minute).Format(time.RFC3339)
	entry.Enabled = true
	es.entries = append(es.entries, entry)
	es.save()
	es.mu.Unlock()
}


// List handles the List HTTP request.
func (es *EnhancedScheduler) List() []ContentScheduleEntry {
	es.mu.RLock()
	defer es.mu.RUnlock()
	out := make([]ContentScheduleEntry, len(es.entries))
	copy(out, es.entries)
	return out
}


// Remove handles the Remove HTTP request.
func (es *EnhancedScheduler) Remove(id string) bool {
	es.mu.Lock()
	defer es.mu.Unlock()
	for i, e := range es.entries {
		if e.ID == id {
			es.entries = append(es.entries[:i], es.entries[i+1:]...)
			es.save()
			return true
		}
	}
	return false
}

type SchedPersist struct {
	RotationMode RotationMode       `json:"rotation_mode"`
	TimeSlots    []TimeSlotSched    `json:"time_slots"`
	Entries      []ContentScheduleEntry `json:"entries"`
}

func (es *EnhancedScheduler) load() {
	b, err := os.ReadFile(es.filePath)
	if err != nil {
		return
	}
	var sp SchedPersist
	if json.Unmarshal(b, &sp) != nil {
		json.Unmarshal(b, &es.entries)
		return
	}
	es.rotationMode = sp.RotationMode
	if sp.TimeSlots != nil {
		es.timeSlots = sp.TimeSlots
	}
	es.entries = sp.Entries
	if es.entries == nil {
		es.entries = make([]ContentScheduleEntry, 0)
	}
}

func (es *EnhancedScheduler) save() {
	sp := SchedPersist{
		RotationMode: es.rotationMode,
		TimeSlots:    es.timeSlots,
		Entries:      es.entries,
	}
	b, _ := json.MarshalIndent(sp, "", "  ")
	os.WriteFile(es.filePath, b, 0644)
}

// DefaultSlots and other items for backward compatibility
type TimeSlot struct {
	Name     string
	StartH   int
	EndH     int
	AdRatio  int
	Vibe     string
}

var DefaultSlots = []TimeSlot{
	{Name: "morning", StartH: 6, EndH: 12, AdRatio: 2, Vibe: "energetic — доброе утро, Остров!"},
	{Name: "afternoon", StartH: 12, EndH: 18, AdRatio: 3, Vibe: "relaxed — солнечный день на Острове"},
	{Name: "evening", StartH: 18, EndH: 23, AdRatio: 2, Vibe: "party — вечерняя свобода!"},
	{Name: "night", StartH: 23, EndH: 6, AdRatio: 1, Vibe: "chill — спокойной ночи, Остров"},
}

type Scheduler struct {
	mu         sync.RWMutex
	Slots      []TimeSlot
	lastSlot   string
	onSlotChange func(old, new string)
}


// NewScheduler handles the NewScheduler HTTP request.
func NewScheduler() *Scheduler {
	return &Scheduler{
		Slots: DefaultSlots,
	}
}


// OnSlotChange handles the OnSlotChange HTTP request.
func (s *Scheduler) OnSlotChange(fn func(old, new string)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.onSlotChange = fn
}


// CurrentSlot handles the CurrentSlot HTTP request.
func (s *Scheduler) CurrentSlot() TimeSlot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	h := time.Now().Hour()
	for _, slot := range s.Slots {
		if slot.StartH <= slot.EndH {
			if h >= slot.StartH && h < slot.EndH {
				return slot
			}
		} else {
			if h >= slot.StartH || h < slot.EndH {
				return slot
			}
		}
	}
	return s.Slots[0]
}


// CurrentSlotName handles the CurrentSlotName HTTP request.
func (s *Scheduler) CurrentSlotName() string {
	return s.CurrentSlot().Name
}


// AdRatio handles the AdRatio HTTP request.
func (s *Scheduler) AdRatio() int {
	return s.CurrentSlot().AdRatio
}


// Vibe handles the Vibe HTTP request.
func (s *Scheduler) Vibe() string {
	return s.CurrentSlot().Vibe
}


// Start handles the Start HTTP request.
func (s *Scheduler) Start() {
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			slot := s.CurrentSlotName()
			s.mu.RLock()
			old := s.lastSlot
			s.mu.RUnlock()
			if slot != old {
				s.mu.Lock()
				s.lastSlot = slot
				fn := s.onSlotChange
				s.mu.Unlock()
				if fn != nil {
					fn(old, slot)
				}
				log.Printf("[radio scheduler] slot changed: %s → %s (%s)", old, slot, s.Vibe())
			}
		}
	}()
}
