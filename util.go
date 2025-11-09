package main

import (
	"net"
	"net/http"
	"strings"
)

func clientIP(r *http.Request) string {
	if ip := strings.TrimSpace(r.Header.Get("CF-Connecting-IP")); ip != "" {
		return ip
	}
	if ip := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); ip != "" {
		if comma := strings.Index(ip, ","); comma >= 0 {
			ip = ip[:comma]
		}
		return strings.TrimSpace(ip)
	}
	host := strings.TrimSpace(r.RemoteAddr)
	if host == "" {
		return "-"
	}
	if parsed, _, err := net.SplitHostPort(host); err == nil && parsed != "" {
		return parsed
	}
	return host
}

func addressForLog(addr string) string {
	if strings.HasPrefix(addr, ":") {
		return "127.0.0.1" + addr
	}
	return addr
}
