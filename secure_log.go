package ssp

import (
	"log"
	"strings"
)

// SafeLogRequest logs a request while redacting sensitive cryptographic fields.
//
// If req is nil it logs "Request: nil". If req.Client is nil it logs the request
// with "cmd=unknown" and "idk=nil" using the provided IP address. Otherwise it
// logs the client command with newlines removed, the Idk truncated for display,
// and a masked IP address to avoid exposing full sensitive values.
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

// SafeLogIdentity logs a short, privacy-preserving summary of an identity.
// If identity is nil it logs "Identity: nil". Otherwise it logs the identity
// key truncated to 8 characters followed by "..." , the Disabled flag, and
// whether the Rekeyed field is non-empty.
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

// SafeLogResponse logs a response's nut and TIF while redacting sensitive data.
// If resp is nil it logs "Response: nil". Otherwise it logs the nut truncated to at most 8 characters followed by "..." and the TIF formatted in hexadecimal (prefixed with 0x).
func SafeLogResponse(resp *CliResponse) {
	if resp == nil {
		log.Printf("Response: nil")
		return
	}
	log.Printf("Response: nut=%s..., tif=0x%X",
		truncateKey(string(resp.Nut), 8),
		resp.TIF)
}

// SafeLogError logs the provided error with the given context if err is non-nil.
// If err is nil, it performs no action. The context parameter is sanitized to
// prevent log injection attacks.
func SafeLogError(context string, err error) {
	if err == nil {
		return
	}
	// Sanitize context to prevent log injection
	safeContext := sanitizeControlChars(context)
	log.Printf("Error [%s]: %v", safeContext, err)
}

// SafeLogAuth logs an authentication event with the identity key truncated for privacy.
// It records the event name, the idk truncated to at most 8 characters (followed by an ellipsis), and whether the authentication succeeded.
// The event parameter is sanitized to prevent log injection attacks.
func SafeLogAuth(event string, idk string, success bool) {
	// Sanitize event to prevent log injection
	safeEvent := sanitizeControlChars(event)
	log.Printf("Auth [%s]: idk=%s..., success=%v",
		safeEvent,
		truncateKey(idk, 8),
		success)
}

// truncateKey safely truncates a key for logging purposes and fully removes newline, carriage return, and other control characters.
// It truncates key to at most maxLen characters, then removes any log injection risks by deleting all \n, \r, NUL, and other ASCII controls.
// If key is empty it returns "(empty)". This function is robust for plain text log output.
func truncateKey(key string, maxLen int) string {
	if key == "" {
		return "(empty)"
	}

	// Truncate to maxLen first
	if len(key) > maxLen {
		key = key[:maxLen]
	}
	// Remove all dangerous control characters, not just replace with spaces
	var b strings.Builder
	for i := 0; i < len(key); i++ {
		c := key[i]
		// Remove \n, \r, NUL, ESC, TAB, DEL, any ASCII <32, 127
		if c == '\n' || c == '\r' || c == '\t' || c == 0x1b || c == 0x00 || c < 32 || c == 127 {
			continue
		}
		b.WriteByte(c)
	}
	safe := b.String()
	if safe == "" {
		return "(empty)"
	}
	return safe
}

// sanitizeControlChars replaces control characters in s with spaces to prevent log injection.
// It substitutes newline, carriage return, tab, NUL, escape, and any ASCII code less than 32 or 127 with a space and returns the resulting string.
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

// sanitizeForLog sanitizes a string for safe logging by removing control characters.
// This is primarily used for URLs and other user-influenced data that could contain
// log injection attacks. Returns a safe string with all control characters removed.
func sanitizeForLog(s string) string {
	if s == "" {
		return "(empty)"
	}
	// Remove all dangerous control characters
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		// Remove newlines, carriage returns, and other control characters
		if c == '\n' || c == '\r' || c == '\t' || c == 0x1b || c == 0x00 || c < 32 || c == 127 {
			continue
		}
		b.WriteByte(c)
	}
	result := b.String()
	if result == "" {
		return "(empty)"
	}
	return result
}

// maskIP partially masks an IP address for privacy.
// If ip is empty it returns "(no-ip)". For IPv6 addresses it returns the first
// colon-separated segment followed by ":***". For IPv4 addresses it masks the
// last two octets as "a.b.*.*". For other formats, if the input is longer than
// eight characters it returns the first four characters plus "***", otherwise
// it returns the input unchanged.
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
