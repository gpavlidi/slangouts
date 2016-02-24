package main

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/nlopes/slack"
)

type SlackClient struct {
	Client      *slack.Client
	Token       string
	SelfId      string
	Messages    chan HangoutsMessage
	DonePolling chan bool
}

func (c *SlackClient) Init(apiToken string) error {
	c.Token = apiToken
	c.Client = slack.New(c.Token)
	//c.Client.SetDebug(true)
	authResponse, err := c.Client.AuthTest()
	if err != nil {
		// ask the user for the slack token
		log.Println("Invalid Slack Token. Please navigate to the below address, create a Token, confirm it and paste it below\n")
		fmt.Println("https://api.slack.com/docs/oauth-test-tokens#test_token_generator\n")
		fmt.Print("Slack Token: ")
		fmt.Scanln(&c.Token)
		c.Client = slack.New(c.Token)
		authResponse, err = c.Client.AuthTest()
		return err
	}

	c.SelfId = authResponse.UserID
	c.Messages = make(chan HangoutsMessage)
	c.DonePolling = make(chan bool)

	return nil
}

func (c *SlackClient) SendMessage(msg HangoutsMessage) error {
	groups, err := c.Client.GetGroups(false)
	if err != nil {
		return err
	}

	var existingGroup *slack.Group = nil
	for _, group := range groups {

		if group.Purpose.Value == msg.conversationId {
			existingGroup = &group
			break
		}
	}
	if existingGroup == nil {
		existingGroup, err = c.Client.CreateGroup(fmt.Sprint("hangouts-", msg.conversationId))
		if err != nil {
			return err
		}
		_, err = c.Client.SetGroupTopic(existingGroup.ID, msg.conversationName)
		if err != nil {
			return err
		}
		_, err = c.Client.SetGroupPurpose(existingGroup.ID, msg.conversationId)
		if err != nil {
			return err
		}
	} else {
		if existingGroup.IsArchived {
			err = c.Client.UnarchiveGroup(existingGroup.ID)
		}
	}

	_, _, err = c.Client.PostMessage(existingGroup.ID, msg.message, slack.PostMessageParameters{Username: msg.senderName})
	return err
}

func (c *SlackClient) StartPolling() error {
	rtm := c.Client.NewRTM()
	go rtm.ManageConnection()

	// slack for some reason echos the last message on connect
	isFirstMsg := true
	go func() {
		for {
			select {
			case <-c.DonePolling:
				return
			case msg := <-rtm.IncomingEvents:
				switch ev := msg.Data.(type) {
				case *slack.MessageEvent:
					// for first message specifically, check if it's older
					// slack for some reason echos the last msg
					if isFirstMsg {
						isFirstMsg = false
						timestamp, _ := strconv.ParseFloat(ev.Timestamp, 64)
						if int64(timestamp) < time.Now().Unix() {
							log.Println("Slack: Skipping first message since it appears old:", ev.Text)
							continue
						}
					}

					if ev.User == c.SelfId {
						group := rtm.GetInfo().GetGroupByID(ev.Channel)
						if group != nil && strings.Contains(group.Topic.Value, "hangouts-") {
							c.Messages <- HangoutsMessage{message: ev.Text, conversationId: group.Purpose.Value}
						}
					}
				case *slack.RTMError:
					log.Println("Slack Error: ", ev.Error())
				case *slack.InvalidAuthEvent:
					log.Fatal("Slack Error: Invalid Slack credentials")
				}
			}
		}
	}()

	return nil
}
