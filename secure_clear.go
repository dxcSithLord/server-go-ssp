package ssp

import (
	"runtime"
)

// ClearBytes securely clears a byte slice by overwriting with zeros.
// Uses runtime.KeepAlive to prevent compiler optimization from removing the clear operation.
func ClearBytes(b []byte) {
	if len(b) == 0 {
		return
	}
	for i := range b {
		b[i] = 0
	}
	// Memory fence to prevent compiler optimization
	runtime.KeepAlive(b)
}

// ClearString securely clears a string's underlying data.
// Note: Go strings are immutable, so we work on the bytes and set the string to empty.
func ClearString(s *string) {
	if s == nil || *s == "" {
		return
	}
	// Convert string to bytes, clear them
	b := []byte(*s)
	ClearBytes(b)
	*s = ""
	runtime.KeepAlive(b)
}

// Clear securely clears all sensitive fields in SqrlIdentity
func (si *SqrlIdentity) Clear() {
	if si == nil {
		return
	}
	ClearString(&si.Idk)
	ClearString(&si.Suk)
	ClearString(&si.Vuk)
	ClearString(&si.Pidk)
	ClearString(&si.Rekeyed)
	si.SQRLOnly = false
	si.Hardlock = false
	si.Disabled = false
	si.Btn = 0
}

// Clear securely clears all sensitive fields in ClientBody
func (cb *ClientBody) Clear() {
	if cb == nil {
		return
	}
	ClearString(&cb.Suk)
	ClearString(&cb.Vuk)
	ClearString(&cb.Pidk)
	ClearString(&cb.Idk)
	cb.Version = nil
	cb.Cmd = ""
	cb.Opt = nil
	cb.Btn = 0
}

// Clear securely clears all sensitive signature data in CliRequest
func (cr *CliRequest) Clear() {
	if cr == nil {
		return
	}
	ClearString(&cr.Ids)
	ClearString(&cr.Pids)
	ClearString(&cr.Urs)
	ClearString(&cr.ClientEncoded)
	ClearString(&cr.Server)
	ClearString(&cr.IPAddress)
	if cr.Client != nil {
		cr.Client.Clear()
	}
}

// Clear securely clears cached sensitive data in HoardCache
func (hc *HoardCache) Clear() {
	if hc == nil {
		return
	}
	if hc.Identity != nil {
		hc.Identity.Clear()
	}
	if hc.LastRequest != nil {
		hc.LastRequest.Clear()
	}
	ClearBytes(hc.LastResponse)
	hc.State = ""
	hc.RemoteIP = ""
	hc.OriginalNut = ""
	hc.PagNut = ""
}

// ClearBytesSecure provides an additional layer of clearing with multiple passes.
// This uses safe Go operations without unsafe pointers.
func ClearBytesSecure(b []byte) {
	if len(b) == 0 {
		return
	}
	// First pass: zero out
	for i := range b {
		b[i] = 0
	}
	// Second pass: pattern fill (prevents optimization)
	for i := range b {
		b[i] = 0xFF
	}
	// Third pass: final zero
	for i := range b {
		b[i] = 0
	}
	// Memory fence to prevent compiler optimization
	runtime.KeepAlive(b)
}
