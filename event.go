package main

import (
	"encoding/json"
	"time"
)

const (
	EventChatMessage = "chatMessage"
	EventChangeRoom  = "changeRoom"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(Event, *Client) error

type ChatMessageEvent struct {
	From   string    `json:"from"`
	Text   string    `json:"text"`
	SentAt time.Time `json:"sentAt"`
}

type ChangeRoomEvent struct {
	Name string `json:"name"`
}

func SendChatMessage(e Event, c *Client) error {
	var messageIn ChatMessageEvent
	if err := json.Unmarshal(e.Payload, &messageIn); err != nil {
		return err
	}

	var messageOut ChatMessageEvent
	messageOut.From = messageIn.From
	messageOut.Text = messageIn.Text
	messageOut.SentAt = time.Now()

	data, err := json.Marshal(messageOut)
	if err != nil {
		return err
	}

	event := Event{
		Type:    EventChatMessage,
		Payload: data,
	}

	for client := range c.manager.clients {
		if client.room == c.room {
			client.event <- event
		}
	}

	return nil
}

func ChangeRoom(e Event, c *Client) error {
	var changeRoomEvent ChangeRoomEvent
	err := json.Unmarshal(e.Payload, &changeRoomEvent)
	if err != nil {
		return err
	}
	c.room = changeRoomEvent.Name
	return nil
}
