package main

import (
	"context"
	"crypto/ed25519"
	"encoding/hex"
	"regexp"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan bool)
	res := make(chan KeyPair)
	for i := 0; i < 10; i++ {
		go findKey(ctx, res)
	}
	go func() {
		pair := <-res
		cancel()
		println(pair.Public)
		println(pair.Private)
		done <- true
	}()
	<-done
}

func findKey(ctx context.Context, res chan KeyPair) {
	var public = ""
	var private = ""

	exp := regexp.MustCompile(`83e(0[1-9]|1[0-2])23$`)

	select {
	case <-ctx.Done():
		return
	default:
		for !exp.MatchString(public) {
			pub, priv, err := ed25519.GenerateKey(nil)
			if err != nil {
				panic(err)
			}
			public = hex.EncodeToString(pub)
			private = hex.EncodeToString(priv)
		}
		res <- KeyPair{public, private}
	}
}

type KeyPair struct {
	Public  string
	Private string
}
