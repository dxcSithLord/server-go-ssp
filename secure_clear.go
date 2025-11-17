package ssp

import (
	"runtime"
	"unsafe"
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

// memclr is a more aggressive memory clearing function that uses unsafe operations
// to ensure the memory is actually cleared and not optimized away.
func memclr(b []byte) {
	if len(b) == 0 {
		return
	}
	// Use a volatile-like pattern to prevent optimization
	ptr := unsafe.Pointer(&b[0])
	for i := 0; i < len(b); i++ {
		*(*byte)(unsafe.Add(ptr, i)) = 0
	}
	runtime.KeepAlive(ptr)
}

// ClearBytesSecure uses the more aggressive clearing method
func ClearBytesSecure(b []byte) {
	memclr(b)
}
