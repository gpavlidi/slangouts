package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"
)

type SlangoutsApp struct {
	configFile     string
	config         SlangoutsConfig
	hangoutsClient *HangoutsClient
	slackClient    *SlackClient
}

type SlangoutsConfig struct {
	HangoutsRefreshToken string `json:"hangoutsRefreshToken"`
	SlackApiToken        string `json:"slackApiToken"`
}

func runSlangouts(hangoutsPollFreq int, configPath string) {
	if configPath == "" {
		configPath = getConfigPath()
	}
	app := &SlangoutsApp{configFile: configPath, hangoutsClient: &HangoutsClient{PollFrequency: hangoutsPollFreq}, slackClient: &SlackClient{}}
	app.Run()
}

func (app *SlangoutsApp) Run() {
	err := app.loadConfig()
	if err != nil {
		log.Println(err, ". Generating blank config...")
		app.saveConfig()
	}

	// check Slack token
	err = app.slackClient.Init(app.config.SlackApiToken)
	if err != nil {
		log.Fatal(err)
	}

	// update Slack token if changed
	if app.slackClient.Token != app.config.SlackApiToken {
		app.config.SlackApiToken = app.slackClient.Token
		app.saveConfig()
		log.Println("Slack token changed. Saving it.")
	}

	// check Hangouts token
	err = app.hangoutsClient.Init(app.config.HangoutsRefreshToken)
	if err != nil {
		//re-try without a refresh token
		app.config.HangoutsRefreshToken = ""
		err = app.hangoutsClient.Init(app.config.HangoutsRefreshToken)
		if err != nil {
			log.Fatal(err)
		}
	}

	// update Hangouts token if changed
	if app.hangoutsClient.Session.RefreshToken != app.config.HangoutsRefreshToken {
		app.config.HangoutsRefreshToken = app.hangoutsClient.Session.RefreshToken
		app.saveConfig()
		log.Println("Hangouts refresh token changed. Saving it.")
	}

	// start Slack Polling
	app.slackClient.StartPolling()
	//defer app.slackClient.StopPolling()

	// start Hangouts Polling
	app.hangoutsClient.StartPolling()
	defer app.hangoutsClient.StopPolling()

	// catch Ctrl+C
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill, syscall.SIGABRT)

	// event loop
	for {
		select {
		case <-signals:
			return
		case msg := <-app.hangoutsClient.Messages:
			log.Println("Hangouts: ", msg)
			err = app.slackClient.SendMessage(msg)
			if err != nil {
				log.Println(err)
			}
		case msg := <-app.slackClient.Messages:
			log.Println("Slack: ", msg)
			err = app.hangoutsClient.SendMessage(msg)
			if err != nil {
				log.Println(err)
			}
		}
	}
}

func (app *SlangoutsApp) loadConfig() error {
	b, err := ioutil.ReadFile(app.configFile)
	if err == nil {
		return json.Unmarshal(b, &app.config)
	}
	return err
}

func (app *SlangoutsApp) saveConfig() error {
	j, err := json.MarshalIndent(&app.config, "", "\t")
	if err == nil {
		os.MkdirAll(path.Dir(app.configFile), os.ModePerm)
		return ioutil.WriteFile(app.configFile, j, 0600)
	}
	return err
}

func getConfigPath() string {
	workingDir := "."
	for _, name := range []string{"HOME", "USERPROFILE"} { // *nix, windows
		if dir := os.Getenv(name); dir != "" {
			workingDir = dir
		}
	}

	return path.Join(workingDir, ".slangouts", "config.json")
}

func OrDie(e error) {
	if e != nil {
		log.Fatal(e)
	}
}
