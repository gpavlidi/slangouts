package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/gpavlidi/go-hangups"
)

type HangoutsMessage struct {
	senderName       string
	message          string
	conversationId   string
	conversationName string
}

type HangoutsClient struct {
	PollFrequency int
	Session       *hangups.Session
	Client        *hangups.Client
	Messages      chan HangoutsMessage
	DonePolling   chan bool
	lastSync      uint64
	SelfId        string
}

func (c *HangoutsClient) Init(refreshToken string) error {
	c.Session = &hangups.Session{RefreshToken: refreshToken}
	err := c.Session.Init()
	if err != nil {
		return err
	}

	c.Client = &hangups.Client{Session: c.Session}
	c.Messages = make(chan HangoutsMessage)
	c.DonePolling = make(chan bool)

	return nil
}

func (c *HangoutsClient) StartPolling() error {
	// find whoami and seed the sync timestamp to current time
	getSelfInfo, _ := c.Client.GetSelfInfo()
	c.lastSync = *getSelfInfo.ResponseHeader.CurrentServerTime
	c.SelfId = *getSelfInfo.SelfEntity.Id.GaiaId

	go func() {
		ticker := time.NewTicker(time.Second * time.Duration(c.PollFrequency))
		for {
			select {
			case <-ticker.C:
				c.poll()
			case <-c.DonePolling:
				return
			}
		}
	}()

	return nil
}

func (c *HangoutsClient) StopPolling() {
	c.DonePolling <- true
	return
}

func (c *HangoutsClient) SendMessage(msg HangoutsMessage) error {
	_, err := c.Client.SendChatMessage(msg.conversationId, msg.message)
	return err
}

func (c *HangoutsClient) poll() {
	newEvents, err := c.Client.SyncAllNewEvents(c.lastSync, 1048576) //1 MB
	if err != nil {
		return
	}
	c.lastSync = *newEvents.ResponseHeader.CurrentServerTime

	for _, conversation := range newEvents.ConversationState {

		//find or generate conversation name
		conversationName := ""
		if conversation.Conversation.Name != nil {
			conversationName = fmt.Sprintf("hangouts-%s", *conversation.Conversation.Name)
		} else {
			participants := make([]string, 0)
			for _, participant := range conversation.Conversation.ParticipantData {
				// skip my name from participants list
				if *participant.Id.GaiaId == c.SelfId {
					continue
				}
				participants = append(participants, *participant.FallbackName)
			}
			conversationName = fmt.Sprintf("hangouts-%s", strings.Join(participants, ","))
		}

		for _, event := range conversation.Event {
			senderId := *event.SenderId.GaiaId

			// dont echo my msgs
			if senderId == c.SelfId {
				continue
			}

			// find sender name
			senderName := "Unknown"
			for _, participant := range conversation.Conversation.ParticipantData {
				if *participant.Id.GaiaId == senderId {
					senderName = *participant.FallbackName
					break
				}
			}

			// reconstruct msg text
			message := ""
			for _, segment := range event.ChatMessage.MessageContent.Segment {
				message = fmt.Sprint(message, *segment.Text)
			}

			// send message to channel
			c.Messages <- HangoutsMessage{message: message, senderName: senderName, conversationName: conversationName, conversationId: *conversation.ConversationId.Id}
		}

		// mark all events in this conversation as read
		_, _ = c.Client.UpdateWatermark(*conversation.ConversationId.Id, c.lastSync)
	}
	return
}
