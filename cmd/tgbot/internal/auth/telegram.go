// Package auth validates Telegram Mini App initData.
package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

// Validator checks Telegram initData signatures.
type Validator struct {
	botToken    string
	internalKey string // server-to-server key (ZeroClaw)
	internalUID int64  // user ID to use for internal requests
}

// New creates a Validator for the given bot token.
func New(botToken, internalKey string, internalUID int64) *Validator {
	return &Validator{botToken: botToken, internalKey: internalKey, internalUID: internalUID}
}

// UserFromRequest extracts and validates identity from the request.
// Supports two auth methods:
//  1. X-Internal-Key header — for server-to-server calls (ZeroClaw)
//  2. X-Init-Data header / init_data query param — Telegram Mini App initData
func (v *Validator) UserFromRequest(r *http.Request) (int64, bool) {
	// Internal key auth (ZeroClaw / server calls)
	if key := r.Header.Get("X-Internal-Key"); key != "" {
		if v.internalKey != "" && key == v.internalKey {
			return v.internalUID, true
		}
		return 0, false
	}
	// Telegram initData auth
	initData := r.Header.Get("X-Init-Data")
	if initData == "" {
		initData = r.URL.Query().Get("init_data")
	}
	if initData == "" {
		return 0, false
	}
	return v.userID(initData)
}

func (v *Validator) userID(initData string) (int64, bool) {
	vals, err := url.ParseQuery(initData)
	if err != nil {
		return 0, false
	}
	hash := vals.Get("hash")
	if hash == "" {
		return 0, false
	}

	var pairs []string
	for k, vs := range vals {
		if k == "hash" {
			continue
		}
		pairs = append(pairs, k+"="+vs[0])
	}
	sort.Strings(pairs)
	checkStr := strings.Join(pairs, "\n")

	secretKey := hmacSHA256([]byte(v.botToken), []byte("WebAppData"))
	expected := hex.EncodeToString(hmacSHA256([]byte(checkStr), secretKey))
	if !hmac.Equal([]byte(expected), []byte(hash)) {
		return 0, false
	}

	userJSON := vals.Get("user")
	if userJSON == "" {
		return 0, false
	}
	var u struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal([]byte(userJSON), &u); err != nil {
		return 0, false
	}
	return u.ID, true
}

func hmacSHA256(data, key []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}
