// Package channels provides SimpleX channel management with local persistence.
//
// A SimpleX channel is a broadcast medium where a creator publishes messages
// to subscribers. This package manages channel metadata, subscriptions, and
// message caching using JSON file persistence.
//
// Types:
//   - Channel: local channel metadata (id, name, link, role, unread count)
//   - ChannelMessage: cached channel message (text, sender, timestamp)
//   - Manager: lifecycle manager with AddChannel, ListChannels, AddMessage, etc.
//
// The Manager persists to <dataDir>/channels/channels.json and is safe for
// concurrent access via sync.RWMutex.
package channels
