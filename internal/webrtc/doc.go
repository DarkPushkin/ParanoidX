// Package webrtc provides WebRTC signaling and TURN/ICE configuration for
// peer-to-peer connections on the simplex-node network. It implements SignalState
// for room-based SDP offer/answer/ICE candidate exchange and ICEConfig for generating
// HMAC-SHA1 TURN credentials (12-hour validity) for the coturn relay server.
// The GenerateConfig method returns full ICE server configuration with turn:// and turns:// URLs.
package webrtc
