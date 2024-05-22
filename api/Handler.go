package api

import (
	"bytes"
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// DiscordInteraction is the request we get from Discord when a user
// triggers a slash Command i.e. /zoom
type DiscordInteraction struct {
	Type   float64                  `json:"type"`
	Data   DiscordInteractionData   `json:"data"`
	Member DiscordInteractionMember `json:"member"`
}

// DiscordInteractionData is present for the slash command itself
// i.e. /zoom
type DiscordInteractionData struct {
	Name    string                          `json:"name"`
	ID      string                          `json:"id"`
	Type    float64                         `json:"type"`
	Options []DiscordInteractionDataOptions `json:"options"`
}

// DiscordInteractionDataOptions contains the option passed in
// within the slash command i.e. the parameters
type DiscordInteractionDataOptions struct {
	Name  string      `json:"name"`
	Type  float64     `json:"type"`
	Value interface{} `json:"value"`
}

type DiscordInteractionMember struct {
	User DiscordInteractionMemberUser `json:"user"`
}

// DiscordInteractionMemberUser gives a way to uniquely
// identify a user by adding # between the Username and
// the Discriminator
type DiscordInteractionMemberUser struct {
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
}

// DiscordResponse is the response we send back to Discord
// See also: https://discord.com/developers/docs/interactions/receiving-and-responding
type DiscordResponse struct {
	Type float64             `json:"type"`
	Data DiscordResponseData `json:"data"`
}

type DiscordResponseData struct {
	Content string                 `json:"content"`
	Embeds  []DiscordResponseEmbed `json:"embeds,omitempty"`
}

type DiscordResponseEmbed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Type        string `json:"type"`
}

const ChannelMessageWithSource = float64(4)

func writeResponse(w http.ResponseWriter, commandRes DiscordResponse) {
	data, _ := json.Marshal(commandRes)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}

func Handle(w http.ResponseWriter, r *http.Request) {
	var body []byte
	if r.Body != nil {
		defer r.Body.Close()
		body, _ = io.ReadAll(r.Body)
	}

	discordMsg := make(map[string]interface{})

	if err := json.Unmarshal(body, &discordMsg); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if v, ok := discordMsg["type"].(float64); ok && v == 1 {
		verify(w, r, body)
		return
	}

	msg := DiscordInteraction{}

	if err := json.Unmarshal(body, &msg); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	command := msg.Data.Name

	if command != "streak" {
		http.Error(w, "invalid command", http.StatusBadRequest)
		return
	}

	resp := checkDailyReward()
	if !resp.Result {
		commandRes := DiscordResponse{
			Type: ChannelMessageWithSource,
			Data: DiscordResponseData{
				Content: "Unable to fetch data from the API try again later :(",
			},
		}

		writeResponse(w, commandRes)
		return
	}

	commandRes := DiscordResponse{
		Type: ChannelMessageWithSource,
		Data: DiscordResponseData{
			Content: fmt.Sprintf("Your current streak is %d that's %f years, last reward was <t:%d:D>", resp.RewardStreak, float32(resp.RewardStreak)/365.0, resp.LastRewardTimestamp),
		},
	}

	writeResponse(w, commandRes)
}

type ResponseAPI struct {
	LastRewardTimestamp int64  `json:"lastRewardTimestamp"`
	CurrentTimestamp    int64  `json:"currentTimestamp"`
	LastReward          string `json:"lastReward"`
	CurrentDate         string `json:"currentDate"`
	RewardStreak        int    `json:"rewardStreak"`
	Result              bool   `json:"result"`
}

func checkDailyReward() ResponseAPI {
	res, err := http.Get(os.Getenv("api-url"))
	if err != nil {
		return ResponseAPI{Result: false}
	}
	defer res.Body.Close()

	var apiResponse ResponseAPI
	if err = json.NewDecoder(res.Body).Decode(&apiResponse); err != nil {
		return ResponseAPI{Result: false}
	}

	return apiResponse
}

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
