// Package radio provides the radio streaming and scheduling system
package radio

import (
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
)

// ffprobeDuration attempts to detect audio duration using ffprobe.
// Returns duration in seconds, or 0 if detection fails.
func ffprobeDuration(filePath string) int {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		filePath,
	)
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	var info struct {
		Format struct {
			Duration string `json:"duration"`
		} `json:"format"`
	}
	if err := json.Unmarshal(out, &info); err != nil {
		return 0
	}
	if info.Format.Duration == "" {
		return 0
	}
	sec, err := strconv.ParseFloat(info.Format.Duration, 64)
	if err != nil {
		return 0
	}
	return int(sec)
}

// detectDuration tries ffprobe, falls back to hardcoded default.
func detectDuration(filePath string, defaultSec int) int {
	if d := ffprobeDuration(filePath); d > 0 {
		return d
	}
	return defaultSec
}

// TrackMetadata holds extended track info detected by ffprobe.
type TrackMetadata struct {
	Duration int    `json:"duration_sec"`
	Bitrate  string `json:"bitrate,omitempty"`
	Codec    string `json:"codec,omitempty"`
	Samplerate string `json:"sample_rate,omitempty"`
	Channels int    `json:"channels,omitempty"`
}

// ProbeTrackMetadata runs ffprobe and returns extended metadata.
func ProbeTrackMetadata(filePath string) *TrackMetadata {
	cmd := exec.Command("ffprobe",
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		filePath,
	)
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var info struct {
		Format struct {
			Duration string `json:"duration"`
			BitRate  string `json:"bit_rate"`
		} `json:"format"`
		Streams []struct {
			CodecType string `json:"codec_type"`
			CodecName string `json:"codec_name"`
			SampleRate string `json:"sample_rate"`
			Channels  int    `json:"channels"`
		} `json:"streams"`
	}
	if err := json.Unmarshal(out, &info); err != nil {
		return nil
	}

	meta := &TrackMetadata{}
	if d, err := strconv.ParseFloat(info.Format.Duration, 64); err == nil {
		meta.Duration = int(d)
	}
	meta.Bitrate = info.Format.BitRate

	for _, s := range info.Streams {
		if s.CodecType == "audio" {
			meta.Codec = s.CodecName
			meta.Samplerate = s.SampleRate
			meta.Channels = s.Channels
			break
		}
	}
	return meta
}

// defaultDuration returns the best-guess duration for a track kind.
func defaultDuration(kind TrackKind) int {
	switch kind {
	case KindAd, KindAIGospel:
		return 30
	case KindNews, KindAINews:
		return 20
	case KindKing, KindAIKing:
		return 40
	case KindAIMusic:
		return 120
	default:
		return 60
	}
}

// UpdateTrackDuration patches a Track with real ffprobe duration.
func UpdateTrackDuration(t *Track) {
	if d := ffprobeDuration(t.FilePath); d > 0 {
		t.Duration = d
	} else {
		t.Duration = defaultDuration(t.Kind)
	}
}

// HasFFProbe returns true if ffprobe is available on the system.
func HasFFProbe() bool {
	_, err := exec.LookPath("ffprobe")
	return err == nil
}

// SafeTrackTitle strips numbered prefixes and underscores.
func SafeTrackTitle(name string) string {
	base := strings.TrimSuffix(name, ".mp3")
	base = strings.TrimSuffix(base, ".ogg")
	base = strings.TrimSuffix(base, ".wav")
	// Strip leading numbers like "01_", "02-"
	for len(base) > 0 && base[0] >= '0' && base[0] <= '9' {
		base = base[1:]
	}
	if len(base) > 0 && (base[0] == '_' || base[0] == '-' || base[0] == ' ') {
		base = base[1:]
	}
	return strings.ReplaceAll(base, "_", " ")
}
