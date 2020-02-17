package websocket_client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	eventsLock sync.RWMutex
	sendLock   sync.Mutex
	events     map[string]*caller
	ws         *websocket.Conn
	ready      bool
	options    Options
	uri        string
}
type Options struct {
	Headers   http.Header
	Reconnect bool
}
type Payload struct {
	Event string      `json:"event,omitempty"`
	Data  interface{} `json:"data,omitempty"`
	Ack   string      `json:"ack,omitempty"`
}

func NewClient(uri string, options Options) *Client {
	return &Client{uri: uri, options: options, events: make(map[string]*caller)}
}
func (c *Client) Connect() error {
	return c.connect()
}
func (c *Client) connect() error {
	ws, _, err := websocket.DefaultDialer.Dial(c.uri, c.options.Headers)
	c.ws = ws
	log.Printf("Connecting to %v", c.uri)
	if err != nil {
		return err
	}
	defer func() {
		c.ready = true
	}()
	for {
		var payload Payload
		err := ws.ReadJSON(&payload)
		if err != nil {
			return err
		}
		c.eventsLock.RLock()
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
			handler.Call(args)
		}
		c.eventsLock.RUnlock()

	}
}
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

var id = &ID{id: 1}

func (client *Client) Send(event string, data ...interface{}) (string, error) {
	client.sendLock.Lock()
	defer client.sendLock.Unlock()
	if len(data) > 1 {
		replay := fmt.Sprintf("%v__%d", event, id.new())
		return replay, client.ws.WriteJSON(Payload{Event: event, Data: data[0], Ack: replay})
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
