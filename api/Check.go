package api

import (
	"encoding/json"
	"net/http"
	"os"
)

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
