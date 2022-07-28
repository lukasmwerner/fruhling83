package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

func IsValidKey(key string) bool {
	exp := regexp.MustCompile(`83e(0[1-9]|1[0-2])(\d\d)$`)
	return exp.MatchString(key)
}

func ValidateKeyAndSignature(k string, signature string, content string) bool {
	key, err := hex.DecodeString(k)
	if err != nil {
		return false
	}
	publicKey := ed25519.PublicKey(key)
	sig, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	return ed25519.Verify(publicKey, []byte(content), sig)
}

var denyedKeys = map[string]bool{
	"d17eef211f510479ee6696495a2589f7e9fb055c2576749747d93444883e0123": true,
}

func OnDenyList(key string) bool {
	return denyedKeys[key]
}

type springHeaders struct {
	Version   string `json:"spring-version"`
	Signature string `json:"spring-signature"`
}

func SpringHeaders(r *http.Request) (springHeaders, error) {
	sh := springHeaders{
		Version:   r.Header.Get("Spring-Version"),
		Signature: r.Header.Get("Spring-Signature"),
	}
	if sh.Version != "83" {
		return springHeaders{}, errors.New("invalid spring version")
	}
	return sh, nil
}

func ValidTimestamp(stamp string) (bool, time.Time, error) {
	t, err := time.Parse(time.RFC3339, stamp)
	if err != nil {
		return false, time.Time{}, err
	}
	if t.Before(time.Now().Add(-time.Hour * 24)) {
		return false, t, errors.New("timestamp is too old")
	}
	if t.After(time.Now().Add(time.Minute)) {
		return false, t, errors.New("timestamp is too new")
	}
	return true, t, nil
}

func GenerateFakePage() (string, string, string) {
	privateHex := `3371f8b011f51632fea33ed0a3688c26a45498205c6097c352bd4d079d224419`
	privBytes, _ := hex.DecodeString(privateHex)
	priv := ed25519.NewKeyFromSeed(privBytes)
	faker := gofakeit.New(0)
	content := fmt.Sprintf("%s: %s", faker.Name(), faker.HackerPhrase())
	now := time.Now().Format(time.RFC3339)
	content = fmt.Sprintf("%s\n<time datetime=\"%s\">%s</time>", content, now, now)
	sig := ed25519.Sign(priv, []byte(content))
	return content, hex.EncodeToString(sig), now
}
