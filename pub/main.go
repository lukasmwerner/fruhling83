package main

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var server = flag.String("server", "fruhling.serveit.space", "Server address")

func main() {
	fileName := os.Args[1]
	file, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	d, err := io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	content := string(d)

	file, err = os.Open("secret.keys")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	d, err = io.ReadAll(file)
	if err != nil {
		panic(err)
	}
	var secret *KeyPair
	err = json.Unmarshal(d, &secret)
	if err != nil {
		panic(err)
	}

	privBytes, err := hex.DecodeString(secret.Private)
	if err != nil {
		panic(err)
	}
	priv := ed25519.PrivateKey(privBytes)
	now := time.Now()
	content = fmt.Sprintf("%s\n<time datetime=\"%s\">%s</time>", content, now.Format(time.RFC3339), now.Format(time.RFC3339))
	sig := ed25519.Sign(priv, []byte(content))
	sigHex := hex.EncodeToString(sig)
	pub := ed25519.PrivateKey(priv).Public().(ed25519.PublicKey)
	pubHex := hex.EncodeToString(pub)
	println("Signature: ", sigHex)
	println("Sending to server...", fmt.Sprintf("https://%s/%s", *server, pubHex))
	req, err := http.NewRequest("PUT", fmt.Sprintf("https://%s/%s", *server, pubHex), bytes.NewBufferString(content))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "text/html")
	req.Header.Set("Spring-Signature", sigHex)
	req.Header.Set("Spring-Version", "83")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	println("Response: ", resp.Status)
	io.Copy(os.Stdout, resp.Body)
}

type KeyPair struct {
	Private string `json:"private"`
	Public  string `json:"public"`
}
