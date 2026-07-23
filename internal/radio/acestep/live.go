// Package acestep provides Acestream integration for P2P streaming
package acestep

import (
	"context"
	"log"
	"time"
)

// LiveBroadcast manages a continuous AI-generated radio stream.
// It pre-generates tracks on a schedule and makes them available for streaming.
type LiveBroadcast struct {
	gen      *Generator
	queue    chan string // local file paths ready for playback
	interval time.Duration
	ctx      context.Context
	cancel   context.CancelFunc
	style    Style
}

// NewLiveBroadcast creates a continuous broadcast that generates a new track every interval.
func NewLiveBroadcast(gen *Generator, interval time.Duration) *LiveBroadcast {
	ctx, cancel := context.WithCancel(context.Background())
	return &LiveBroadcast{
		gen:      gen,
		queue:    make(chan string, 10),
		interval: interval,
		ctx:      ctx,
		cancel:   cancel,
		style:    StyleMusic,
	}
}

// Start begins the generation loop. Generates on a rotating topic schedule.
func (b *LiveBroadcast) Start() {
	// Pre-generate first tracks
	for i := 0; i < 3; i++ {
		path, _ := b.gen.GenerateMusic("gentle waves and palm trees")
		if path != "" {
			b.queue <- path
		}
	}

	go func() {
		ticker := time.NewTicker(b.interval)
		defer ticker.Stop()

		topics := []string{
			"gentle waves and palm trees — island vibes",
			"happy children playing in the sun",
			"sunset meditation with ocean breeze",
			"morning birds and fresh coconut water",
			"stars over Saint Mary Island at midnight",
		}
		idx := 0

		for {
			select {
			case <-b.ctx.Done():
				return
			case <-ticker.C:
				// Rotate through different styles
				switch idx % 4 {
				case 0:
					path, _ := b.gen.GenerateMusic(topics[idx%len(topics)])
					if path != "" {
						b.queue <- path
					}
				case 1:
					path, _ := b.gen.GenerateAdvert("Остров Святой Марии — рай на земле!")
					if path != "" {
						b.queue <- path
					}
				case 2:
					path, _ := b.gen.GenerateNews("Жизнь на Острове процветает!")
					if path != "" {
						b.queue <- path
					}
				case 3:
					path, _ := b.gen.GenerateDecree("Да будет так на Острове Святой Марии!")
					if path != "" {
						b.queue <- path
					}
				}
				idx++
			}
		}
	}()
	log.Printf("[acestep live] broadcast started, interval=%s", b.interval)
}

// Stop halts the broadcast.
func (b *LiveBroadcast) Stop() {
	b.cancel()
	log.Println("[acestep live] broadcast stopped")
}

// NextTrack blocks until the next track is ready, or returns "" if stopped.
func (b *LiveBroadcast) NextTrack() string {
	select {
	case path := <-b.queue:
		return path
	case <-b.ctx.Done():
		return ""
	}
}

// QueueLen returns how many pre-generated tracks are queued.
func (b *LiveBroadcast) QueueLen() int {
	return len(b.queue)
}

// Status returns broadcast status info.
func (b *LiveBroadcast) Status() map[string]any {
	return map[string]any{
		"running":  b.ctx.Err() == nil,
		"queued":   len(b.queue),
		"interval": b.interval.String(),
		"style":    "rotating (music/ad/news/decree)",
	}
}
