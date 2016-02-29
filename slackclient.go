package main

import (
	"fmt"
	"log"
	"regexp"
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
	groups      []slack.Group
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
	var err error
	// check local group cache first
	existingGroup := c.GetGroupByPurpose(msg.conversationId)

	// if not found update cache and check again
	if existingGroup == nil {
		c.UpdateGroups()
		existingGroup = c.GetGroupByPurpose(msg.conversationId)
	}

	// nowhere to be found, create a new group
	if existingGroup == nil {
		existingGroup, err = c.Client.CreateGroup(fmt.Sprint("hangouts-", msg.conversationId))
		if err != nil {
			return err
		}
		c.groups = append(c.groups, *existingGroup)

		topic, err := c.Client.SetGroupTopic(existingGroup.ID, msg.conversationName)
		if err != nil {
			return err
		}
		existingGroup.Topic.Value = topic

		purpose, err := c.Client.SetGroupPurpose(existingGroup.ID, msg.conversationId)
		if err != nil {
			return err
		}
		existingGroup.Purpose.Value = purpose
	} else {
		if existingGroup.IsArchived {
			err = c.Client.UnarchiveGroup(existingGroup.ID)
			if err != nil {
				return err
			}
			existingGroup.IsArchived = false
		}
	}

	_, _, err = c.Client.PostMessage(existingGroup.ID, msg.message, slack.PostMessageParameters{Username: msg.senderName})
	return err
}

func (c *SlackClient) GetGroupById(id string) *slack.Group {
	for _, group := range c.groups {
		if group.ID == id {
			return &group
		}
	}
	return nil
}

func (c *SlackClient) GetGroupByPurpose(purpose string) *slack.Group {
	for _, group := range c.groups {
		if group.Purpose.Value == purpose {
			return &group
		}
	}
	return nil
}

func (c *SlackClient) UpdateGroups() error {
	var err error
	c.groups, err = c.Client.GetGroups(false)
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
				case *slack.ConnectedEvent:
					// get available groups at slack so we can correlate them to a hangout id
					c.groups = ev.Info.Groups
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
						group := c.GetGroupById(ev.Channel)
						if group == nil {
							log.Println("Slack: Got msg from a group (", ev.Channel, ") not cached locally. Refreshing groups.")
							c.UpdateGroups()
							group = c.GetGroupById(ev.Channel)
							if group == nil {
								log.Println("Slack: Cant find group", ev.Channel, "even after updating Groups. Skipping message:", ev.Text)
							}
						}
						if group != nil && strings.Contains(group.Topic.Value, "hangouts-") {
							// links are enclosed in <> - remove them
							linkRegex := regexp.MustCompile("^<(.*)>$")
							pipeRegex := regexp.MustCompile("(.*)\\|(.*)")
							parsedText := ev.Text
							if linkRegex.MatchString(parsedText) {
								parsedText = pipeRegex.ReplaceAllString(linkRegex.ReplaceAllString(parsedText, "$1"), "$1")
							}

							c.Messages <- HangoutsMessage{message: parsedText, conversationId: group.Purpose.Value}
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
