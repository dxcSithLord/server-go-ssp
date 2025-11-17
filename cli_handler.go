package ssp

import (
	"fmt"
	"log"
	"net/http"
)

var supportedCommands = map[string]bool{
	"query":   true,
	"ident":   true,
	"enable":  true,
	"disable": true,
	"remove":  true,
}

// Cli implements the /cli.sqrl endpoint
func (api *SqrlSspAPI) Cli(w http.ResponseWriter, r *http.Request) {
	// SECURITY: Sanitize URL before logging to prevent log injection
	SafeLogInfo("Req: %v", sanitizeForLog(r.URL.String()))
	nut := Nut(r.URL.Query().Get("nut"))
	if nut == "" {
		_, _ = w.Write(NewCliResponse("", "").WithClientFailure().Encode())
		return
	}

	// response mutates from here depending on available values
	response := NewCliResponse(Nut(nut), api.qry(nut))
	req, err := ParseCliRequest(r)
	if err != nil {
		// SECURITY: Sanitize error to prevent log injection from user input
		SafeLogError("parse_request", err)
		_, _ = w.Write(response.WithClientFailure().WithCommandFailed().Encode())
		return
	}
	// Signature is OK from here on!

	// SECURITY: Clear sensitive data from request after it's stored in HoardCache
	// This defer runs AFTER writeResponse due to LIFO order, ensuring data is saved first
	defer req.Clear()

	// defer writing the response and saving the new nut
	defer api.writeResponse(req, response, w)

	// SECURITY: Use safe logging instead of dumping full request
	SafeLogRequest(req)

	hoardCache, err := api.getAndDelete(Nut(nut))
	if err != nil {
		if err == ErrNotFound {
			// SECURITY: Sanitize nut value to prevent log injection
			SafeLogInfo("Nut %s not found", sanitizeForLog(string(nut)))
			response.WithClientFailure().WithCommandFailed()
			return
		}
		SafeLogError("nut_lookup", err)
		response.WithTransientError().WithCommandFailed()
		return
	}
	response.HoardCache = hoardCache

	// validation checks
	err = api.requestValidations(hoardCache, req, r, response)
	if err != nil {
		return
	}

	if req.Client.Cmd == "query" {
		tmpIdent := req.Identity()
		tmpIdent.Btn = -1
		response.Ask = api.Authenticator.AskResponse(tmpIdent)
	}

	// generate new nut
	nut, err = api.tree.Nut()
	if err != nil {
		SafeLogError("nut_generation", err)
		response.WithCommandFailed()
		return
	}

	// new nut to the response from here on out
	response.Nut = nut
	response.Qry = api.qry(nut)

	// check if the same user has already been authenticated previously

	identity, err := api.authStore.FindIdentity(req.Client.Idk)
	if err != nil && err != ErrNotFound {
		SafeLogError("identity_lookup", err)
		response.WithCommandFailed()
		return
	}

	// Check is we know about a previous identity
	previousIdentity, err := api.checkPreviousIdentity(req, response)
	if err != nil {
		return
	}

	if identity != nil {
		err := api.knownIdentity(req, response, identity)
		if err != nil {
			return
		}
	} else if req.Client.Cmd == "ident" {
		// create new identity from the request
		identity = req.Identity()
		// handle previous identity swap if the current identity is new
		err := api.checkPreviousSwap(previousIdentity, identity, response)
		if err != nil {
			return
		}

		// Do we id match on first auth? grc says nope; PaulF and I think yes
		response.WithIDMatch()
	}
	api.setSuk(req, response, identity)

	// Finish authentication and saving
	api.finishCliResponse(req, response, identity, hoardCache)
}

func (api *SqrlSspAPI) writeResponse(req *CliRequest, response *CliResponse, w http.ResponseWriter) {
	respBytes := response.Encode()
	// SECURITY: Do not log full response content as it may contain sensitive data
	SafeLogResponse(response)

	// always save back the new nut
	if response.HoardCache != nil {
		err := api.hoard.Save(response.Nut, &HoardCache{
			State:        "associated",
			RemoteIP:     response.HoardCache.RemoteIP,
			OriginalNut:  response.HoardCache.OriginalNut,
			PagNut:       response.HoardCache.PagNut,
			LastRequest:  req,
			LastResponse: respBytes,
		}, api.NutExpiration)
		if err != nil {
			SafeLogError("hoard_save", err)
			response.WithCommandFailed()
			respBytes = response.Encode()
		} else {
			// SECURITY: Sanitize nut before logging
			SafeLogInfo("Saved nut %s in hoard", sanitizeForLog(string(response.Nut)))
		}
	}
	_, _ = w.Write(respBytes)
	log.Println()
}

func (api *SqrlSspAPI) setSuk(req *CliRequest, response *CliResponse, identity *SqrlIdentity) {
	if req.Client.Opt["suk"] {
		if identity != nil {
			response.Suk = identity.Suk
		} else if req.Client.Cmd == "ident" {
			response.Suk = req.Client.Suk
		}
	}
}

func (api *SqrlSspAPI) finishCliResponse(req *CliRequest, response *CliResponse, identity *SqrlIdentity, hoardCache *HoardCache) {
	accountDisabled := false
	if identity != nil {
		accountDisabled = identity.Disabled
	}
	if req.IsAuthCommand() && !accountDisabled {
		// SECURITY: Use safe logging for identity information
		SafeLogAuth("authenticate", identity.Idk, true)
		authURL, err := api.authenticateIdentity(identity, req.Client.Btn)
		if err != nil {
			SafeLogError("save_identity", err)
			response.WithCommandFailed()
			return
		}
		if req.Client.Opt["cps"] {
			// SECURITY: Sanitize auth URL before logging to prevent log injection
			SafeLogAuth("cps_auth_set", sanitizeForLog(authURL), true)
			response.URL = authURL
		}
	}

	// fail the ident on account disable
	if req.Client.Cmd == "ident" && accountDisabled {
		response.WithCommandFailed()
	}

	if req.IsAuthCommand() && !accountDisabled {
		// for non-CPS we save the state back to the PagNut for redirect on polling
		if !req.Client.Opt["cps"] {
			err := api.hoard.Save(hoardCache.PagNut, &HoardCache{
				State:       "authenticated",
				RemoteIP:    hoardCache.RemoteIP,
				OriginalNut: hoardCache.OriginalNut,
				PagNut:      hoardCache.PagNut,
				LastRequest: req,
				Identity:    identity,
			}, api.NutExpiration)
			if err != nil {
				SafeLogError("hoard_save_pagnut", err)
				response.WithCommandFailed()
			}
			// SECURITY: Sanitize pagnut before logging
			SafeLogInfo("Saved pagnut %s in hoard", sanitizeForLog(string(hoardCache.PagNut)))
		}
	}
}

func (api *SqrlSspAPI) checkPreviousSwap(previousIdentity, identity *SqrlIdentity, response *CliResponse) error {
	if previousIdentity != nil {
		err := api.swapIdentities(previousIdentity, identity)
		if err != nil {
			SafeLogError("identity_swap", err)
			response.WithCommandFailed()
			return fmt.Errorf("identity swap error")
		}
		// SECURITY: Use safe logging without exposing full identity details
		SafeLogAuth("identity_swap", identity.Idk, true)
		// TODO should we clear the PreviousIDMatch here?
		response.ClearPreviousIDMatch()
	}
	return nil
}

func (api *SqrlSspAPI) checkPreviousIdentity(req *CliRequest, response *CliResponse) (*SqrlIdentity, error) {
	var previousIdentity *SqrlIdentity
	var err error
	if req.Client.Pidk != "" {
		previousIdentity, err = api.authStore.FindIdentity(req.Client.Pidk)
		if err != nil && err != ErrNotFound {
			SafeLogError("lookup_previous_identity", err)
			response.WithCommandFailed()
			return nil, err
		}
	}
	if previousIdentity != nil {
		response.WithPreviousIDMatch()
		// as per spec, proactively return the suk on pidk match
		req.Client.Opt["suk"] = true
	}
	return previousIdentity, nil
}

func (api *SqrlSspAPI) requestValidations(hoardCache *HoardCache, req *CliRequest, r *http.Request, response *CliResponse) error {
	req.IPAddress = api.RemoteIP(r)
	// validate last response against this request
	if hoardCache.LastResponse != nil && !req.ValidateLastResponse(hoardCache.LastResponse) {
		response.WithCommandFailed()
		// SECURITY: Do not log response content as it contains sensitive data
		log.Printf("Last response validation failed")
		return fmt.Errorf("validation error")
	}

	// validate the IP if required
	if hoardCache.RemoteIP != req.IPAddress {
		if !req.Client.Opt["noiptest"] {
			// SECURITY: Mask IP addresses to prevent log injection and maintain privacy
			SafeLogInfo("Rejecting on IP mis-match orig: %s current: %s", maskIP(hoardCache.RemoteIP), maskIP(api.RemoteIP(r)))
			response.WithCommandFailed()
			return fmt.Errorf("validation error")
		}
	} else {
		log.Print("Matched IP addresses")
		response = response.WithIPMatch()
	}

	// validating the current request and associated Idk's match
	if hoardCache.LastRequest != nil && hoardCache.LastRequest.Client.Idk != req.Client.Idk {
		// SECURITY: Truncate identity keys to prevent log injection
		SafeLogInfo("Identity mismatch orig: %s... current: %s...", truncateKey(hoardCache.LastRequest.Client.Idk, 8), truncateKey(req.Client.Idk, 8))
		response.WithCommandFailed().WithClientFailure().WithBadIDAssociation()
		return fmt.Errorf("validation error")
	}

	if !supportedCommands[req.Client.Cmd] {
		response.WithFunctionNotSupported()
		return fmt.Errorf("Uknown command: %v", req.Client.Cmd)
	}

	return nil
}

func (api *SqrlSspAPI) knownIdentity(req *CliRequest, response *CliResponse, identity *SqrlIdentity) error {
	if identity.Rekeyed != "" {
		response.WithIdentitySuperseded()
		// SECURITY: Use truncated key for logging
		SafeLogAuth("rekeyed_attempt", identity.Idk, false)
		if req.Client.Cmd != "query" {
			response.WithCommandFailed()
		}
		return fmt.Errorf("attempted use of rekeyed identity")
	} else {
		response.WithIDMatch()
	}
	// copy the current Btn value from the request
	identity.Btn = req.Client.Btn
	changed := false
	if req.IsAuthCommand() {
		changed = req.UpdateIdentity(identity)
	}
	if req.Client.Cmd == "enable" || req.Client.Cmd == "remove" {
		err := req.VerifyUrs(identity.Vuk)
		if err != nil {
			SafeLogError("urs_validation", err)
			// TODO: remove since sig check failed here?
			if identity.Disabled {
				response.WithSQRLDisabled()
			}
			response.WithClientFailure().WithCommandFailed()
			return fmt.Errorf("identity error")
		}
		if req.Client.Cmd == "enable" {
			SafeLogAuth("enable_account", identity.Idk, true)
			identity.Disabled = false
			changed = true
		} else if req.Client.Cmd == "remove" {
			err := api.removeIdentity(identity)
			if err != nil {
				SafeLogError("remove_identity", err)
				response.WithClientFailure().WithCommandFailed()
				return fmt.Errorf("identity error")
			}
			response.ClearIDMatch()
			SafeLogAuth("remove_identity", identity.Idk, true)
		}
	}
	if req.Client.Cmd == "disable" {
		identity.Disabled = true
		changed = true
	}

	if identity.Disabled {
		req.Client.Opt["suk"] = true
		response.WithSQRLDisabled()
	}
	if changed {
		err := api.authStore.SaveIdentity(identity)
		if err != nil {
			SafeLogError("save_identity", err)
			response.WithClientFailure().WithCommandFailed()
			return fmt.Errorf("identity error")
		}
	}
	return nil
}
