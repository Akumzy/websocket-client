package main

import (
	"log"
	"net/http"
	"sync"

	WS "github.com/Akumzy/websocket-client"
)

var token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJfaWQiOiI1ZGI4ZDE1YThiNjg3NjI3M2QzMDNjYWUiLCJlbWFpbCI6ImFrdW1haXNhYWNha3VtYUBnbWFpbC5jb20iLCJkZXZpY2VfaWQiOiI1ZGI4ZDM1MzA1MGRiOTVlM2Y5MGIwNjUiLCJ0b2tlbklkIjoiNWU0NTdiYWZiYmY5MTI0MzkxY2ZlYzQ5IiwiaWF0IjoxNTgxNjExOTUxLCJleHAiOjE1ODQyMDM5NTEsImlzcyI6ImNvbS5mZXhzcGFjZS5jbG91ZCJ9.xcqr8ZeHZiOKzakfhhOXbl8RcnLaNhTopN95Vd8r-ro"

type User struct {
	ID              string `json:"_id,omitempty"`
	CompanyID       string `json:"company_id,omitempty"`
	AccountType     int    `json:"account_type,omitempty"`
	FilesystemEntry string `json:"filesystem_entry,omitempty"`
	DirectoryID     string `json:"directory_id,omitempty"`
	DepartmentID    string `json:"department_id,omitempty"`
	DeviceID        string `json:"device_id,omitempty"`
}

var client *WS.Client

func main() {
	header := http.Header{"Content-Type": []string{"application/json"}, "Authorization": []string{token}}
	client = WS.NewClient("ws://localhost:4444/ws", WS.Options{Headers: header})
	client.On("update", func(user User) {
		log.Println(user)
	})
	client.On("ready", func(ready bool) {

		log.Println(ready)
		go send()
	})

	err := client.Connect()
	log.Println(err)
}
func send() {
	var wg sync.WaitGroup
	wg.Add(1)
	var mode bool
	reply, err := client.Send("bingo", true, true)
	if err != nil {
		panic(err)
	}
	client.On(reply, func(y bool) {
		mode = y
		wg.Done()
	})
	wg.Wait()
	log.Println("Mode: ", mode)
}
