# remindlater-bot

[![Go Report Card](https://goreportcard.com/badge/github.com/vladislavtomenko/remindlater-bot)](https://goreportcard.com/report/github.com/vladislavtomenko/remindlater-bot)
[![Build Status](https://travis-ci.org/vladislavtomenko/remindlater-bot.svg?branch=master)](https://travis-ci.org/vladislavtomenko/remindlater-bot)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](https://opensource.org/licenses/MIT)

This telegram bot strives to implement missing reminder functionality in the messenger. You can run your own bot or use [@remindlater_bot](https://telegram.me/remindlater_bot).

<img src="img/bot.jpg" width="30%">

# Features

* Receives forwarded or regular messages and reminds about it in the predefined time
* Reminders can be either snoozed or marked as completed

# Install

Generate an authorisation token for a bot. The process is described [here](https://core.telegram.org/bots#6-botfather).

Clone the repo and and compile the project
```bash
git clone https://github.com/vladislavtomenko/remindlater-bot.git
cd reminder-bot
go get -d -v ./...
go install -v ./...
```

Set the environment variables

Name | Description | Default
--------- | ----------- | -------
`DEBUG` | Log debug message to stdout | `false`
`DURATION` | Default reminder delay in seconds | `20`
`TOKEN` | Telegram bot API token | 

Run the bot
```bash
DEBUG=true DURATION=1200 TOKEN=xxxxxx $GOBIN/remindlater-bot
```
