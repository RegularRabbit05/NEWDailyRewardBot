package api

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os"
)

func verify(w http.ResponseWriter, r *http.Request, body []byte) {
	publicKey := os.Getenv("discord-public-key")

	signature := r.Header.Get("X-Signature-Ed25519")
	timestamp := r.Header.Get("X-Signature-Timestamp")

	signatureHexDecoded, err := hex.DecodeString(signature)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if len(signatureHexDecoded) != ed25519.SignatureSize {
		http.Error(w, "invalid signature length", http.StatusUnauthorized)
		return
	}

	publicKeyHexDecoded, err := hex.DecodeString(publicKey)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	pubKey := [32]byte{}

	copy(pubKey[:], publicKeyHexDecoded)

	var msg bytes.Buffer
	msg.WriteString(timestamp)
	msg.Write(body)

	verified := ed25519.Verify(publicKeyHexDecoded, msg.Bytes(), signatureHexDecoded)

	if !verified {
		http.Error(w, "invalid request signature", http.StatusUnauthorized)
		return
	}

	p := map[string]float64{
		"type": float64(1),
	}

	w.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(w).Encode(p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
