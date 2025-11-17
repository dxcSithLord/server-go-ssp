package ssp

import (
	"runtime"
	"unsafe"
)

// ClearBytes securely clears a byte slice by overwriting with zeros.
// ClearBytes overwrites every element of b with zero.
// If b is empty, ClearBytes does nothing.
// It calls runtime.KeepAlive(b) to prevent the compiler from eliding the overwrite.
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

// ClearString securely clears a string's underlying data by directly accessing
// the backing memory using unsafe operations.
//
// IMPORTANT LIMITATION: Go strings are immutable by design. This function uses
// unsafe operations to overwrite the backing memory, but:
// 1. If the string was interned or shared, other references may still see data
// 2. String copies created via string() conversion cannot be cleared
// 3. Small strings may be stored inline and not in separate backing memory
//
// For sensitive data, prefer []byte over string where possible, as byte slices
// can be reliably cleared with ClearBytes.
func ClearString(s *string) {
	if s == nil || *s == "" {
		return
	}

	// Use unsafe.StringData to get pointer to backing array (Go 1.20+)
	// This accesses the actual memory backing the string
	ptr := unsafe.StringData(*s)
	if ptr != nil {
		// Create a mutable byte slice over the string's backing memory
		// WARNING: This violates Go's immutability guarantees for strings
		b := unsafe.Slice(ptr, len(*s))
		// Clear the backing memory
		for i := range b {
			b[i] = 0
		}
		runtime.KeepAlive(b)
	}

	// Set string pointer to empty to prevent further access
	*s = ""
	runtime.KeepAlive(s)
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
// ClearBytesSecure overwrites the provided byte slice three times (zero, 0xFF, zero) to reduce the risk that compiler optimizations leave sensitive data in memory.
// If the slice is empty the function returns immediately. It uses only safe Go operations (no unsafe pointers) and calls runtime.KeepAlive to ensure the slice is retained until the clears complete.
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
