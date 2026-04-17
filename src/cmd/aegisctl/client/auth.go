package client

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

const accessKeyTokenPath = "/api/v2/auth/access-key/token"

// AccessKeyTokenDebug contains the fully materialized signed request data for
// POST /api/v2/auth/access-key/token.
type AccessKeyTokenDebug struct {
	Method          string
	Path            string
	AccessKey       string
	Timestamp       string
	Nonce           string
	CanonicalString string
	Signature       string
}

func (d *AccessKeyTokenDebug) Headers() map[string]string {
	return map[string]string{
		"X-Access-Key": d.AccessKey,
		"X-Timestamp":  d.Timestamp,
		"X-Nonce":      d.Nonce,
		"X-Signature":  d.Signature,
	}
}

// accessKeyTokenResponseData matches dto.AccessKeyTokenResp.
type accessKeyTokenResponseData struct {
	Token     string    `json:"token"`
	TokenType string    `json:"token_type"`
	ExpiresAt time.Time `json:"expires_at"`
	AuthType  string    `json:"auth_type"`
	AccessKey string    `json:"access_key"`
}

// tokenRefreshRequest matches dto.TokenRefreshReq.
type tokenRefreshRequest struct {
	Token string `json:"token"`
}

// tokenRefreshResponseData matches dto.TokenRefreshResp.
type tokenRefreshResponseData struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
}

// LoginResult contains the result of a successful login.
type LoginResult struct {
	Token     string
	ExpiresAt time.Time
	AuthType  string
	AccessKey string
}

// LoginWithAccessKey exchanges an access key signature for a bearer token.
func LoginWithAccessKey(server, accessKey, secretKey string) (*LoginResult, error) {
	accessKey = strings.TrimSpace(accessKey)
	secretKey = strings.TrimSpace(secretKey)
	if accessKey == "" {
		return nil, fmt.Errorf("access key is required")
	}
	if secretKey == "" {
		return nil, fmt.Errorf("secret key is required")
	}

	c := NewClient(server, "", 30*time.Second)
	debugInfo, err := PrepareAccessKeyTokenDebug(accessKey, secretKey, time.Now().UTC(), "")
	if err != nil {
		return nil, fmt.Errorf("prepare signed headers: %w", err)
	}

	var resp APIResponse[accessKeyTokenResponseData]
	if err := c.PostWithHeaders(accessKeyTokenPath, debugInfo.Headers(), &resp); err != nil {
		return nil, fmt.Errorf("exchange access key token failed: %w", err)
	}

	return &LoginResult{
		Token:     resp.Data.Token,
		ExpiresAt: resp.Data.ExpiresAt,
		AuthType:  resp.Data.AuthType,
		AccessKey: resp.Data.AccessKey,
	}, nil
}

// RefreshToken refreshes an existing JWT token.
func RefreshToken(server, currentToken string) (string, time.Time, error) {
	c := NewClient(server, currentToken, 30*time.Second)

	var resp APIResponse[tokenRefreshResponseData]
	err := c.Post("/api/v2/auth/refresh", tokenRefreshRequest{
		Token: currentToken,
	}, &resp)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("token refresh failed: %w", err)
	}

	return resp.Data.Token, resp.Data.ExpiresAt, nil
}

// ProfileData represents the user profile returned by GET /api/v2/auth/profile.
type ProfileData struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar,omitempty"`
	Role     string `json:"role,omitempty"`
}

// GetProfile fetches the current user's profile.
func GetProfile(server, token string) (*ProfileData, error) {
	c := NewClient(server, token, 30*time.Second)

	var resp APIResponse[ProfileData]
	if err := c.Get("/api/v2/auth/profile", &resp); err != nil {
		return nil, fmt.Errorf("get profile failed: %w", err)
	}

	return &resp.Data, nil
}

// IsTokenExpired checks whether the stored token expiry has passed.
func IsTokenExpired(expiry time.Time) bool {
	if expiry.IsZero() {
		return false // unknown expiry — assume valid
	}
	return time.Now().After(expiry)
}

// PrepareAccessKeyTokenDebug builds the canonical string, signature, and
// headers for the access-key token exchange request.
func PrepareAccessKeyTokenDebug(accessKey, secretKey string, now time.Time, nonce string) (*AccessKeyTokenDebug, error) {
	accessKey = strings.TrimSpace(accessKey)
	secretKey = strings.TrimSpace(secretKey)
	nonce = strings.TrimSpace(nonce)
	if accessKey == "" {
		return nil, fmt.Errorf("access key is required")
	}
	if secretKey == "" {
		return nil, fmt.Errorf("secret key is required")
	}
	var err error
	if nonce == "" {
		nonce, err = newAccessKeyNonce()
		if err != nil {
			return nil, err
		}
	}

	timestamp := strconv.FormatInt(now.Unix(), 10)
	canonical := canonicalAccessKeyString("POST", accessKeyTokenPath, accessKey, timestamp, nonce)

	return &AccessKeyTokenDebug{
		Method:          "POST",
		Path:            accessKeyTokenPath,
		AccessKey:       accessKey,
		Timestamp:       timestamp,
		Nonce:           nonce,
		CanonicalString: canonical,
		Signature:       signAccessKeyRequest(secretKey, canonical),
	}, nil
}

func buildAccessKeyHeaders(accessKey, secretKey string, now time.Time, path string) (map[string]string, error) {
	debugInfo, err := prepareAccessKeyDebug(accessKey, secretKey, now, path, "")
	if err != nil {
		return nil, err
	}
	return debugInfo.Headers(), nil
}

func prepareAccessKeyDebug(accessKey, secretKey string, now time.Time, path, nonce string) (*AccessKeyTokenDebug, error) {
	accessKey = strings.TrimSpace(accessKey)
	secretKey = strings.TrimSpace(secretKey)
	nonce = strings.TrimSpace(nonce)
	if accessKey == "" {
		return nil, fmt.Errorf("access key is required")
	}
	if secretKey == "" {
		return nil, fmt.Errorf("secret key is required")
	}
	var err error
	if nonce == "" {
		nonce, err = newAccessKeyNonce()
		if err != nil {
			return nil, err
		}
	}

	timestamp := strconv.FormatInt(now.Unix(), 10)
	canonical := canonicalAccessKeyString("POST", path, accessKey, timestamp, nonce)

	return &AccessKeyTokenDebug{
		Method:          "POST",
		Path:            path,
		AccessKey:       accessKey,
		Timestamp:       timestamp,
		Nonce:           nonce,
		CanonicalString: canonical,
		Signature:       signAccessKeyRequest(secretKey, canonical),
	}, nil
}

func canonicalAccessKeyString(method, path, accessKey, timestamp, nonce string) string {
	return strings.Join([]string{
		strings.ToUpper(method),
		path,
		accessKey,
		timestamp,
		nonce,
	}, "\n")
}

func signAccessKeyRequest(secretKey, payload string) string {
	mac := hmac.New(sha256.New, []byte(secretKey))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

func newAccessKeyNonce() (string, error) {
	nonce := make([]byte, 16)
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	return hex.EncodeToString(nonce), nil
}
