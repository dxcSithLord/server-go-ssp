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

// truncateKey safely truncates a key for logging purposes
func truncateKey(key string, maxLen int) string {
	if key == "" {
		return "(empty)"
	}
	if len(key) <= maxLen {
		return key
	}
	return key[:maxLen]
}

// maskIP partially masks an IP address for privacy
func maskIP(ip string) string {
	if ip == "" {
		return "(no-ip)"
	}
	if len(ip) > 10 {
		// Show only first few characters for identification
		return ip[:len(ip)/2] + "..."
	}
	return ip
}
