package main

import (
	"NEWDailyRewardBot/api"
	"net/http"
)

//https://discord.com/oauth2/authorize?client_id=1242937339886436525&scope=applications.commands%20bot&permissions=2048

func main() {
	http.HandleFunc("/api/init", api.Init)
	http.HandleFunc("/api/handle", api.Handle)
	http.ListenAndServe(":8000", nil)
}
