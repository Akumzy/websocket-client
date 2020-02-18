## It's worth pointing out that this is a complete rewrite from the original forked [go-socket.io-client](https://github.com/zhouhui8915/go-socket.io-client) which is tailored around my current requirement. But I believe it will definitely work with any base websocket server.

```go

package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	WS "github.com/Akumzy/websocket-client"
)

var token = "jwt-token"

type User struct {
	ID   string `json:"_id,omitempty"`
	Name string `json:"name,omitempty"`
}

var client *WS.Client

func main() {
    // Your custom headers
	header := http.Header{"Authorization": []string{token}}
    client = WS.NewClient("ws://localhost:4444/ws", WS.Options{Headers: header})
    // Subscribe to server sent events
	client.On("update", func(user User) {
		log.Println(user)
	})
	client.On("ready", func(ready bool) {
		log.Println(ready)
		 sendMessageWithAcknowledgment()
    })
    // Send message to your websocket server
    client.Send("feed-me-more", time.Now())

	err := client.Connect()
	log.Println(err)
}
func sendMessageWithAcknowledgment() {
	var wg sync.WaitGroup
    wg.Add(1)

	var mode bool
	reply, err := client.Send("bingo", true, true)
	if err != nil {
		panic(err)
    }
    // Make sure to set a timeout
	timer := time.AfterFunc(time.Second*20, func() {
		wg.Done()
	})
	client.On(reply, func(y bool) {
		timer.Stop()
		mode = y
		wg.Done()
	})
	wg.Wait()
	log.Println("Mode: ", mode)
}

```

## Messages sent or received from server most have this object structure

```go
type Payload struct {
	// Event name used to identify event handlers
	Event string `json:"event"`
	// Message payload
	Data interface{} `json:"data"`
	// Ack is string(event name) that will be sent to server which
	// an acknowledgment will be published/sent to and the client will
	// need to get the event name from client.Send method after emitting an
	// event to server. (optional)
	Ack string `json:"ack"`
}
```

This package depends on awesome [gorilla/websocket](https://github.com/gorilla/websocket)
