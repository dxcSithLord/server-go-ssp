package ssp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/yeqown/go-qrcode/v2"
	"github.com/yeqown/go-qrcode/writer/standard"
)

type nutJSON struct {
	Nut        Nut `json:"nut"`
	Pagnut     Nut `json:"pag"`
	Expiration int `json:"exp"`
}

// Nut implements the /nut.sqrl endpoint
// TODO sin, ask and 1-9 params
func (api *SqrlSspAPI) Nut(w http.ResponseWriter, r *http.Request) {
	hoardCache, err := api.createAndSaveNut(r)
	if err != nil {
		log.Print(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Add("Content-Type", "application/json")
		respObj := &nutJSON{
			Nut:        hoardCache.OriginalNut,
			Pagnut:     hoardCache.PagNut,
			Expiration: api.NutExpirationSeconds(),
		}
		enc, err := json.Marshal(respObj)
		if err != nil {
			SafeLogError("json_encode_nut", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(enc)
		return
	}
	w.Header().Add("Content-Type", "application/x-www-form-urlencoded")
	values := make(url.Values)
	values.Add("nut", string(hoardCache.OriginalNut))
	values.Add("pag", string(hoardCache.PagNut))
	values.Add("exp", fmt.Sprintf("%d", api.NutExpirationSeconds()))

	if referer := r.Header.Get("Referer"); referer != "" {
		values.Add("can", Sqrl64.EncodeToString([]byte(referer)))
	}

	_, err = w.Write([]byte(values.Encode()))
	if err != nil {
		SafeLogError("nut_response_write", err)
	}
}

func (api *SqrlSspAPI) createAndSaveNut(r *http.Request) (*HoardCache, error) {
	nut, err := api.tree.Nut()
	if err != nil {
		return nil, fmt.Errorf("failed generating nut: %v", err)
	}
	pagnut, err := api.tree.Nut()
	if err != nil {
		return nil, fmt.Errorf("failed generating nut: %v", err)
	}

	hoardCache := &HoardCache{
		State:       "issued",
		RemoteIP:    api.RemoteIP(r),
		OriginalNut: nut,
		PagNut:      pagnut,
	}
	// Store the nut in the hoard under OriginalNut
	err = api.hoard.Save(nut, hoardCache, api.NutExpiration)
	if err != nil {
		return nil, fmt.Errorf("Failed to save a nut: %v", err)
	}
	// Also store under PagNut for polling (initially in "pending" state)
	// This allows /pag.sqrl to distinguish between invalid pagnuts (404) and pending auth (200 empty)
	pendingCache := &HoardCache{
		State:       "pending",
		RemoteIP:    hoardCache.RemoteIP,
		OriginalNut: nut,
		PagNut:      pagnut,
		Identity:    nil, // No identity until authentication completes
	}
	err = api.hoard.Save(pagnut, pendingCache, api.NutExpiration)
	if err != nil {
		return nil, fmt.Errorf("Failed to save pagnut: %v", err)
	}
	// SECURITY: Sanitize nut and mask IP to prevent log injection
	SafeLogInfo("Saved nut %s in hoard from %s", sanitizeForLog(string(nut)), maskIP(hoardCache.RemoteIP))
	return hoardCache, nil
}

func (api *SqrlSspAPI) getAndDelete(nut Nut) (*HoardCache, error) {
	hoardCache, err := api.hoard.GetAndDelete(Nut(nut))
	if err != nil {
		return nil, err
	}
	return hoardCache, nil
}

// PNG implements the /png.sqrl endpoint
func (api *SqrlSspAPI) PNG(w http.ResponseWriter, r *http.Request) {
	nut := r.URL.Query().Get("nut")
	var hoardCache *HoardCache
	var err error
	nutWasProvided := nut != ""

	if nut == "" {
		// create a nut
		hoardCache, err = api.createAndSaveNut(r)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		nut = string(hoardCache.OriginalNut)
	} else {
		// Validate provided nut exists in hoard
		_, err = api.hoard.Get(Nut(nut))
		if err != nil {
			if err == ErrNotFound {
				// SECURITY: Sanitize nut value to prevent log injection
				SafeLogInfo("PNG requested with invalid nut: %s", sanitizeForLog(nut))
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte("Invalid or expired nut"))
				return
			}
			SafeLogError("png_nut_lookup", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	params := make(url.Values)
	params.Add("nut", nut)
	sqrlURL := &url.URL{
		Scheme:   SqrlScheme,
		Host:     api.Host(r),
		Path:     fmt.Sprintf("%v/cli.sqrl", api.RootPath),
		RawQuery: params.Encode(),
	}

	value := sqrlURL.String()

	// Create QR code with medium error correction
	qrc, err := qrcode.NewWith(value,
		qrcode.WithEncodingMode(qrcode.EncModeByte),
		qrcode.WithErrorCorrectionLevel(qrcode.ErrorCorrectionMedium),
	)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Failed to create QR code"))
		return
	}

	// Generate PNG to bytes buffer
	buf := bytes.NewBuffer(nil)
	qrWriter := standard.NewWithWriter(&nopCloser{Writer: buf},
		standard.WithQRWidth(10),
		standard.WithBuiltinImageEncoder(standard.PNG_FORMAT),
	)
	if err = qrc.Save(qrWriter); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Failed to encode PNG"))
		return
	}
	// Close the writer to finalize the output
	if err = qrWriter.Close(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Failed to finalize PNG"))
		return
	}
	png := buf.Bytes()

	// Only add nut headers when we generated a new nut (not when client provided one)
	if !nutWasProvided && hoardCache != nil {
		w.Header().Add("Sqrl-Nut", string(hoardCache.OriginalNut))
		w.Header().Add("Sqrl-Pag", string(hoardCache.PagNut))
		w.Header().Add("Sqrl-Exp", fmt.Sprintf("%d", api.NutExpirationSeconds()))
	}
	w.Header().Add("Content-Type", "image/png")
	_, _ = w.Write(png)
}

type pagJSON struct {
	URL string `json:"url"`
}

// Pag implements the /pag.sqrl endpoint
func (api *SqrlSspAPI) Pag(w http.ResponseWriter, r *http.Request) {
	nut := r.URL.Query().Get("nut")
	if nut == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Missing required nut parameter"))
		return
	}
	pagnut := r.URL.Query().Get("pag")
	if pagnut == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("Missing required pag parameter"))
		return
	}

	hoardCache, err := api.getAndDelete(Nut(pagnut))
	if err != nil {
		if err == ErrNotFound {
			// SECURITY: Sanitize pagnut value to prevent log injection
			SafeLogInfo("Pag requested with invalid pagnut: %s", sanitizeForLog(pagnut))
			w.WriteHeader(http.StatusNotFound)
			return
		}
		SafeLogError("pag_nut_lookup", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Failed nut lookup"))
		return
	}

	// Check if authentication is still pending
	if hoardCache.State == "pending" || hoardCache.Identity == nil {
		// SECURITY: Sanitize pagnut value to prevent log injection
		SafeLogInfo("Pag polling for pending authentication: %s", sanitizeForLog(pagnut))
		// Re-save the entry since we called getAndDelete
		err = api.hoard.Save(Nut(pagnut), hoardCache, api.NutExpiration)
		if err != nil {
			SafeLogError("pag_resave", err)
		}
		// Return 200 OK with empty response to indicate authentication still pending
		if r.Header.Get("Accept") == "application/json" {
			w.Header().Add("Content-Type", "application/json")
			respObj := &pagJSON{URL: ""}
			enc, err := json.Marshal(respObj)
			if err != nil {
				SafeLogError("json_encode_pag_pending", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			_, _ = w.Write(enc)
			return
		}
		// For non-JSON requests, return empty body
		w.WriteHeader(http.StatusOK)
		return
	}

	if hoardCache.OriginalNut != Nut(nut) {
		log.Print("Got query for pagnut but original nut doesn't match")
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	if hoardCache.Identity == nil {
		log.Print("Nil identity on pag hoardCache")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("Missing identity"))
		return
	}

	if r.Header.Get("Accept") == "application/json" {
		w.Header().Add("Content-Type", "application/json")
		respObj := &pagJSON{
			URL: api.Authenticator.AuthenticateIdentity(hoardCache.Identity),
		}
		enc, err := json.Marshal(respObj)
		if err != nil {
			SafeLogError("json_encode_pag", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(enc)
		return
	}

	_, _ = w.Write([]byte(api.Authenticator.AuthenticateIdentity(hoardCache.Identity)))
}

// nopCloser wraps an io.Writer to add a no-op Close() method, making it an io.WriteCloser
type nopCloser struct {
	io.Writer
}

func (nopCloser) Close() error { return nil }
