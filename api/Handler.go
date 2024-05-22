package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

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
