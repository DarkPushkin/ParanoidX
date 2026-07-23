// Package radio implements radio streaming, playlist management, and station handling.
//
// Features:
//   - Station management (create, list, delete per genre)
//   - Playlist generation from station directories
//   - HTTP audio streaming (direct + onion-routed via Tor)
//   - AI-generated content descriptions
//   - M3U8 playlist export
//   - Track upload via API
//   - Audio file detection and validation (ffprobe)
//   - Genre-based station organization with auto-symlink
//   - OnionStream: Tor-routed stream handler for remote listeners
//
// File organization:
//   - radio/ (root):       Shared audio files
//   - radio/stations/:     Station-specific directories with symlinked content
//   - playlists/:          Generated M3U8 playlists
package radio
