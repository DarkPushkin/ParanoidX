// Package middleware provides HTTP middleware for security and rate limiting
package middleware

import (
	"net"
	"net/http"
	"strings"
)

var localSubnets []*net.IPNet

func init() {
	for _, cidr := range []string{"127.0.0.0/8", "::1/128", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"} {
		_, n, err := net.ParseCIDR(cidr)
		if err == nil {
			localSubnets = append(localSubnets, n)
		}
	}
}

func isPrivateIP(ip net.IP) bool {
	for _, n := range localSubnets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}


// IsLocalOrOnionAccess handles the IsLocalOrOnionAccess HTTP request.
func IsLocalOrOnionAccess(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip != nil && isPrivateIP(ip) {
		return true
	}
	host = strings.ToLower(r.Host)
	onion := strings.HasSuffix(host, ".onion")
	if onion {
		return true
	}
	return strings.HasSuffix(host, ".local")
}


// IsStrictLocalOnly handles the IsStrictLocalOnly HTTP request.
func IsStrictLocalOnly(r *http.Request) bool {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		host = r.RemoteAddr
	}
	ip := net.ParseIP(host)
	if ip != nil && ip.IsLoopback() {
		return true
	}
	return false
}


// DenyIfNotLocalOrOnion handles the DenyIfNotLocalOrOnion HTTP request.
func DenyIfNotLocalOrOnion(w http.ResponseWriter, r *http.Request) bool {
	if !IsLocalOrOnionAccess(r) {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return true
	}
	return false
}
