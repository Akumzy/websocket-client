package websocket_client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	eventsLock sync.RWMutex
	sendLock   sync.Mutex
	events     map[string]*caller
	ws         *websocket.Conn
	Ready      bool
	options    Options
	uri        string
}
type Options struct {
	Headers       http.Header
	AutoReconnect bool
	ReconnetAfter time.Duration
}
type Payload struct {
	// Event name used to identify event handlers
	Event string `json:"event,omitempty"`
	// Message payload
	Data interface{} `json:"data,omitempty"`
	// Ack is string(event name) that will be sent to server which
	// an acknowledgment will be published/sent to and the client will
	// need to get the event name from client.Send method after emitting an
	// event to server.
	Ack string `json:"ack,omitempty"`
}

func NewClient(uri string, options Options) *Client {
	return &Client{uri: uri, options: options, events: make(map[string]*caller)}
}
func (c *Client) Connect() error {
	err := c.connect()
	// Something seams off with the code below
	// until I figure it out it will remain like that.
	if err != websocket.ErrCloseSent {
		if c.options.AutoReconnect {
			time.Sleep(c.options.ReconnetAfter)
			c.Connect()
		}
	}
	return err
}

// On method is used to subscribe to server sent events
func (client *Client) On(message string, f interface{}) (err error) {
	c, err := newCaller(f)
	if err != nil {
		return
	}
	client.eventsLock.Lock()
	client.events[message] = c
	client.eventsLock.Unlock()
	return
}

// RemoveHandler
func (client *Client) RemoveHandler(message string) {
	client.eventsLock.Lock()
	delete(client.events, message)
	client.eventsLock.Unlock()
}

var id = &ID{id: 0}

// Send method sends message to your Websocket server
// It can take only take upto 3 arguments
// first being the event-name your server handler will subscribed to (required)
// second being the message payload of any type (optional)
// and third is being an indicator (recommended as bool) showing that this message requires
// an acknowledgement
func (client *Client) Send(event string, data ...interface{}) (string, error) {
	client.sendLock.Lock()
	defer client.sendLock.Unlock()
	if len(data) > 1 {
		reply := fmt.Sprintf("%v__%d", event, id.new())
		return reply, client.ws.WriteJSON(Payload{Event: event, Data: data[0], Ack: reply})
	}
	return "", client.ws.WriteJSON(Payload{Event: event, Data: data[0]})
}

type ID struct {
	sync.Mutex
	id int
}

func (i *ID) new() int {
	i.Lock()
	defer i.Unlock()
	i.id++
	return i.id
}

func (c *Client) connect() error {
	ws, _, err := websocket.DefaultDialer.Dial(c.uri, c.options.Headers)
	c.ws = ws
	log.Printf("Connecting to %v", c.uri)
	if err != nil {
		return err
	}
	defer func() {
		c.Ready = false
		ws.Close()
	}()
	ws.SetPingHandler(func(appData string) error {
		c.Ready = true
		return nil
	})
	for {
		var payload Payload
		err := ws.ReadJSON(&payload)
		if err != nil {
			return err
		}
		err = func() error {
			c.eventsLock.RLock()
			defer c.eventsLock.RUnlock()
			handler, ok := c.events[payload.Event]
			if ok {
				args := handler.GetArgs()
				if len(args) > 0 {
					t := args[0]
					buf, err := json.Marshal(payload.Data)
					if err != nil {
						return err
					}
					err = json.Unmarshal(buf, &t)
					if err != nil {
						return err
					}
					args[0] = t
				}
				go handler.Call(args)
			}
			return nil
		}()
		if err != nil {
			return err
		}

	}
}
