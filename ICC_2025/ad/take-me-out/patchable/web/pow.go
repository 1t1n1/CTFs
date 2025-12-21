package main

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	powPurposeSubmission = "submission"
	powPurposeTest       = "test"
	powPurposeAdmin      = "admin_debug"
)

var (
	errPowInvalid       = errors.New("invalid proof-of-work payload")
	errPowExpired       = errors.New("proof-of-work token expired")
	errPowMismatch      = errors.New("proof-of-work signature mismatch")
	errPowDifficulty    = errors.New("insufficient proof-of-work difficulty")
	errPowReuse         = errors.New("proof-of-work value already used")
	errPowConfiguration = errors.New("proof-of-work manager not initialized")
)

type powChallenge struct {
	Target     string `json:"target"`
	Difficulty int    `json:"difficulty"`
	ExpiresAt  int64  `json:"expires_at"`
	Purpose    string `json:"purpose"`
	Signature  string `json:"signature"`
}

type powProof struct {
	Target     string `json:"target"`
	Difficulty int    `json:"difficulty"`
	ExpiresAt  int64  `json:"expires_at"`
	Purpose    string `json:"purpose"`
	Signature  string `json:"signature"`
	Nonce      string `json:"nonce"`
}

type powManager struct {
	mu         sync.Mutex
	secret     []byte
	ttl        time.Duration
	difficulty int
	used       map[string]time.Time
}

var powMgr *powManager

func initPowManager() {
	mgr, err := newPowManager()
	if err != nil {
		log.Fatalf("pow manager init failed: %v", err)
	}
	powMgr = mgr
}

func newPowManager() (*powManager, error) {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		return nil, fmt.Errorf("pow secret: %w", err)
	}

	difficulty := envIntWithClamp("POW_DIFFICULTY_BITS", 18, 1, 32)
	ttlSeconds := envIntWithClamp("POW_TTL_SECONDS", 60, 1, 3600)
	ttl := time.Duration(ttlSeconds) * time.Second

	return &powManager{
		secret:     secret,
		ttl:        ttl,
		difficulty: difficulty,
		used:       make(map[string]time.Time),
	}, nil
}

func envIntWithClamp(key string, fallback, min, max int) int {
	val := strings.TrimSpace(os.Getenv(key))
	if val == "" {
		return fallback
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	if n < min {
		return min
	}
	if n > max {
		return max
	}
	return n
}

func (pm *powManager) Difficulty() int {
	if pm == nil {
		return 0
	}
	return pm.difficulty
}

func (pm *powManager) Issue(userID int, challengeName, purpose string) (powChallenge, error) {
	if pm == nil {
		return powChallenge{}, errPowConfiguration
	}
	targetBytes := make([]byte, 32)
	if _, err := rand.Read(targetBytes); err != nil {
		return powChallenge{}, fmt.Errorf("pow target: %w", err)
	}
	target := base64.RawURLEncoding.EncodeToString(targetBytes)
	expiresAt := time.Now().Add(pm.ttl).Unix()

	sig := pm.sign(userID, challengeName, purpose, target, expiresAt, pm.difficulty)

	return powChallenge{
		Target:     target,
		Difficulty: pm.difficulty,
		ExpiresAt:  expiresAt,
		Purpose:    purpose,
		Signature:  sig,
	}, nil
}

func (pm *powManager) Verify(userID int, challengeName string, proof powProof, purpose string) error {
	if pm == nil {
		return errPowConfiguration
	}
	if proof.Nonce == "" || proof.Target == "" || proof.Signature == "" {
		return errPowInvalid
	}
	if proof.Purpose != purpose {
		return errPowInvalid
	}
	if proof.Difficulty != pm.difficulty {
		return errPowDifficulty
	}
	now := time.Now().Unix()
	if proof.ExpiresAt <= now {
		return errPowExpired
	}

	expectedSig := pm.sign(userID, challengeName, purpose, proof.Target, proof.ExpiresAt, proof.Difficulty)
	if !hmacEqual(expectedSig, proof.Signature) {
		return errPowMismatch
	}

	hashed := sha256.Sum256([]byte(proof.Target + ":" + proof.Nonce))
	if leadingZeroBits(hashed[:]) < proof.Difficulty {
		return errPowDifficulty
	}

	key := base64.RawURLEncoding.EncodeToString(hashed[:])

	pm.mu.Lock()
	defer pm.mu.Unlock()
	pm.cleanupLocked(time.Now())
	if _, ok := pm.used[key]; ok {
		return errPowReuse
	}
	pm.used[key] = time.Unix(proof.ExpiresAt, 0)
	return nil
}

func (pm *powManager) cleanupLocked(now time.Time) {
	for k, expiry := range pm.used {
		if expiry.Before(now) {
			delete(pm.used, k)
		}
	}
}

func (pm *powManager) sign(userID int, challengeName, purpose, target string, expiresAt int64, difficulty int) string {
	mac := hmac.New(sha256.New, pm.secret)
	data := fmt.Sprintf("%d|%s|%s|%s|%d|%d", userID, challengeName, purpose, target, expiresAt, difficulty)
	mac.Write([]byte(data))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func leadingZeroBits(b []byte) int {
	count := 0
	for _, v := range b {
		if v == 0 {
			count += 8
			continue
		}
		for i := 7; i >= 0; i-- {
			if v&(1<<uint(i)) == 0 {
				count++
			} else {
				return count
			}
		}
	}
	return count
}

func hmacEqual(expected, actual string) bool {
	expBytes, err1 := base64.RawURLEncoding.DecodeString(expected)
	actBytes, err2 := base64.RawURLEncoding.DecodeString(actual)
	if err1 != nil || err2 != nil {
		return false
	}
	return hmac.Equal(expBytes, actBytes)
}
