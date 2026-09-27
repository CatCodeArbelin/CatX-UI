package audit

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net"
	"net/url"
	"strings"
)

var forbiddenKeys = map[string]struct{}{
	"password": {}, "passwd": {}, "token": {}, "secret": {}, "authorization": {},
	"cookie": {}, "set-cookie": {}, "body": {}, "payload": {}, "content": {},
}

func sanitizeMetadata(input map[string]any) string {
	allowed := make(map[string]any)
	for key, value := range input {
		key = strings.ToLower(strings.TrimSpace(key))
		if key == "" {
			continue
		}
		if _, forbidden := forbiddenKeys[key]; forbidden || strings.Contains(key, "auth") || strings.Contains(key, "pass") {
			continue
		}
		switch typed := value.(type) {
		case string:
			if len(typed) <= 256 {
				allowed[key] = typed
			}
		case bool, int, int64, float64:
			allowed[key] = typed
		}
	}
	data, _ := json.Marshal(allowed)
	return string(data)
}

func targetRef(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:8])
}

func publicHTTPS(raw string) error {
	u, err := url.ParseRequestURI(raw)
	if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil {
		return errInvalidWebhookURL
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return errInvalidWebhookURL
	}
	host := u.Hostname()
	if host == "" || strings.Contains(host, "\\") {
		return errInvalidWebhookURL
	}
	if ip := net.ParseIP(host); ip != nil {
		if !isPublicIP(ip) {
			return errPrivateWebhookAddress
		}
		return nil
	}
	return nil
}

func isPublicIP(ip net.IP) bool {
	return ip != nil && ip.IsGlobalUnicast() && !ip.IsPrivate() && !ip.IsLoopback() &&
		!ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast() && !ip.IsMulticast() && !ip.IsUnspecified()
}
