package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const sessionCookieName = "session"

var (
	sessionSecret       []byte
	sessionCookieSecure bool
	sessionDuration     = 24 * time.Hour
)

type sessionClaims struct {
	Subject   string `json:"sub"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

type jwtHeaderFields struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

func initAuth() error {
	secret := strings.TrimSpace(os.Getenv("SESSION_SECRET"))
	if secret == "" {
		return errors.New("SESSION_SECRET environment variable must be set")
	}
	sessionSecret = []byte(secret)

	switch strings.ToLower(strings.TrimSpace(os.Getenv("SESSION_COOKIE_SECURE"))) {
	case "1", "true", "t", "yes", "y", "on":
		sessionCookieSecure = true
	default:
		sessionCookieSecure = false
	}

	return nil
}

// getUser retrieves the logged-in user from the session cookie
func getUser(r *http.Request) *User {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return nil
	}
	claims, err := parseSessionToken(cookie.Value)
	if err != nil {
		log.Printf("invalid session token: %v", err)
		return nil
	}
	user, err := getUserByUsername(claims.Subject)
	if err != nil {
		return nil
	}
	return user
}

// setSession logs in a user by setting a JWT session cookie
func setSession(w http.ResponseWriter, userID string) error {
	token, expiresAt, err := createSessionToken(userID)
	if err != nil {
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   sessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// clearSession logs out a user
func clearSession(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   sessionCookieSecure,
		SameSite: http.SameSiteLaxMode,
	})
}

func createSessionToken(userID string) (string, time.Time, error) {
	if len(sessionSecret) == 0 {
		return "", time.Time{}, errors.New("session secret is not initialized")
	}
	now := time.Now().UTC()
	expiration := now.Add(sessionDuration)
	claims := sessionClaims{
		Subject:   userID,
		IssuedAt:  now.Unix(),
		ExpiresAt: expiration.Unix(),
	}
	payload, err := json.Marshal(claims)
	if err != nil {
		return "", time.Time{}, err
	}
	headerJSON, err := json.Marshal(jwtHeaderFields{
		Alg: "HS256",
		Typ: "JWT",
	})
	if err != nil {
		return "", time.Time{}, err
	}
	header := base64.RawURLEncoding.EncodeToString(headerJSON)
	body := base64.RawURLEncoding.EncodeToString(payload)
	unsigned := header + "." + body
	signature := sign(unsigned)
	token := unsigned + "." + signature
	return token, expiration, nil
}

func parseSessionToken(token string) (*sessionClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 || len(parts) > 3 {
		return nil, errors.New("token format invalid")
	}
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	var header jwtHeaderFields
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, err
	}
	alg := strings.ToUpper(strings.TrimSpace(header.Alg))
	if alg == "" {
		alg = "HS256"
	}
	signature := ""
	if len(parts) == 3 {
		signature = parts[2]
	}
	unsigned := parts[0] + "." + parts[1]
	if alg == "HS256" {
		if len(sessionSecret) == 0 {
			return nil, errors.New("session secret is not initialized")
		}
		expectedSig := sign(unsigned)
		if !hmac.Equal([]byte(signature), []byte(expectedSig)) {
			return nil, errors.New("signature mismatch")
		}
	} else if alg != "NONE" {
		return nil, errors.New("unsupported alg: " + header.Alg)
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var claims sessionClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, err
	}
	if claims.Subject == "" {
		return nil, errors.New("subject missing")
	}
	now := time.Now().UTC().Unix()
	if claims.ExpiresAt == 0 || claims.ExpiresAt < now {
		return nil, errors.New("token expired")
	}
	return &claims, nil
}

func sign(unsigned string) string {
	mac := hmac.New(sha256.New, sessionSecret)
	mac.Write([]byte(unsigned))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
