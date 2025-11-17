package ssp

import (
	"log"
	"strings"
)

// SafeLogRequest logs request information without exposing sensitive cryptographic data
func SafeLogRequest(req *CliRequest) {
	if req == nil {
		log.Printf("Request: nil")
		return
	}
	if req.Client == nil {
		log.Printf("Request: cmd=unknown, idk=nil, ip=%s", req.IPAddress)
		return
	}
	safeCmd := strings.ReplaceAll(strings.ReplaceAll(req.Client.Cmd, "\n", ""), "\r", "")
	log.Printf("Request: cmd=%s, idk=%s..., ip=%s",
		safeCmd,
		truncateKey(req.Client.Idk, 8),
		maskIP(req.IPAddress))
}

// SafeLogIdentity logs identity information without exposing full keys
func SafeLogIdentity(identity *SqrlIdentity) {
	if identity == nil {
		log.Printf("Identity: nil")
		return
	}
	log.Printf("Identity: idk=%s..., disabled=%v, rekeyed=%v",
		truncateKey(identity.Idk, 8),
		identity.Disabled,
		identity.Rekeyed != "")
}

// SafeLogResponse logs response information without exposing sensitive data
func SafeLogResponse(resp *CliResponse) {
	if resp == nil {
		log.Printf("Response: nil")
		return
	}
	log.Printf("Response: nut=%s..., tif=0x%X",
		truncateKey(string(resp.Nut), 8),
		resp.TIF)
}

// SafeLogError logs error information safely
func SafeLogError(context string, err error) {
	if err == nil {
		return
	}
	log.Printf("Error [%s]: %v", context, err)
}

// SafeLogAuth logs authentication events without sensitive data
func SafeLogAuth(event string, idk string, success bool) {
	log.Printf("Auth [%s]: idk=%s..., success=%v",
		event,
		truncateKey(idk, 8),
		success)
}

// truncateKey safely truncates a key for logging purposes and sanitizes control characters
// to prevent log injection attacks
func truncateKey(key string, maxLen int) string {
	if key == "" {
		return "(empty)"
	}

	// Sanitize control characters first to prevent log injection
	sanitized := sanitizeControlChars(key)

	// Always truncate to maxLen
	if len(sanitized) > maxLen {
		return sanitized[:maxLen]
	}
	return sanitized
}

// sanitizeControlChars replaces control characters with safe visible tokens
// to prevent log injection attacks
func sanitizeControlChars(s string) string {
	result := make([]byte, 0, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch c {
		case '\n':
			result = append(result, ' ')
		case '\r':
			result = append(result, ' ')
		case '\t':
			result = append(result, ' ')
		case '\x00':
			result = append(result, ' ')
		case '\x1b': // Escape character
			result = append(result, ' ')
		default:
			// Replace any other control characters (ASCII 0-31, 127)
			if c < 32 || c == 127 {
				result = append(result, ' ')
			} else {
				result = append(result, c)
			}
		}
	}
	return string(result)
}

// maskIP partially masks an IP address for privacy
func maskIP(ip string) string {
	if ip == "" {
		return "(no-ip)"
	}
	// Sanitize control characters to prevent log injection
	ip = strings.ReplaceAll(ip, "\n", "")
	ip = strings.ReplaceAll(ip, "\r", "")
	ip = strings.ReplaceAll(ip, "\t", "")
	
	// Mask based on IP structure
	if strings.Contains(ip, ":") {
		// IPv6: show only first segment
		parts := strings.Split(ip, ":")
		if len(parts) > 0 {
			return parts[0] + ":***"
		}
	}
	// IPv4: mask last two octets
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		return parts[0] + "." + parts[1] + ".*.*"
	}
	// Fallback for unknown format
	if len(ip) > 8 {
		return ip[:4] + "***"
	}
	return ip
}
