package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	WS "github.com/Akumzy/websocket-client"
)

var token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJfaWQiOiI1ZGI4ZDE1YThiNjg3NjI3M2QzMDNjYWUiLCJlbWFpbCI6ImFrdW1haXNhYWNha3VtYUBnbWFpbC5jb20iLCJkZXZpY2VfaWQiOiI1ZGI4ZDM1MzA1MGRiOTVlM2Y5MGIwNjUiLCJ0b2tlbklkIjoiNWU0NTdiYWZiYmY5MTI0MzkxY2ZlYzQ5IiwiaWF0IjoxNTgxNjExOTUxLCJleHAiOjE1ODQyMDM5NTEsImlzcyI6ImNvbS5mZXhzcGFjZS5jbG91ZCJ9.xcqr8ZeHZiOKzakfhhOXbl8RcnLaNhTopN95Vd8r-ro"

type User struct {
	ID   string `json:"_id,omitempty"`
	Name string `json:"name,omitempty"`
}

var client *WS.Client

func main() {
	header := http.Header{"Content-Type": []string{"application/json"}, "Authorization": []string{token}}
	client = WS.NewClient("ws://localhost:4444/ws", WS.Options{Headers: header, AutoReconnect: true, ReconnetAfter: time.Second * 20})
	client.On("update", func(user User) {
		log.Println(user)
	})
	client.On("ready", func(ready bool) {
		log.Println(ready)
		sendMessageWithAcknowledgment()
	})

	err := client.Connect()
	log.Println(err)
}
func sendMessageWithAcknowledgment() {
	log.Printf("Is connection ready %t", client.Ready)
	var wg sync.WaitGroup
	wg.Add(1)
	var mode bool
	reply, err := client.Send("bingo", true, true)
	if err != nil {
		panic(err)
	}
	timer := time.AfterFunc(time.Second*20, func() {
		wg.Done()
	})
	log.Println("Acknowledgment channel", reply)
	client.On(reply, func(y bool) {
		timer.Stop()
		mode = y
		wg.Done()
		client.RemoveHandler(reply)
	})
	wg.Wait()
	log.Println("Mode: ", mode)
}
