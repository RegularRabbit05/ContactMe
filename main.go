package main

import (
	"ContactMe/api"
	"net/http"
)

func main() {
	http.HandleFunc("/api/sendmessage", api.SendMessage)
	http.ListenAndServe(":3000", nil)
}
