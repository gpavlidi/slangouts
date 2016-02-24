# Slangouts
Slack up front, Hangouts in the rear.

## Why?
I use Slack daily at work, and lots of friends/family were using Hangouts. Didnt want to keep both open so Slangouts came to the rescue. It acts as a bridge between the 2 keeping them in sync while it runs. This means people talk to you in Hangouts, and you can reply on Slack. 

## Configuration
First time your run Slangouts, it 'll ask you for access to both your Slack and Hangout accounts. Follow the on-screen instructions to complete the configuration.

## Usage
```
# build slangouts
go build -o slangouts

# see all available switches
./slangouts help

# example
./slangouts --config ~/.slangouts/config.json --poll 10 
```

## Cross Compiling for the Pi
Below is mostly for me to keep handy for compiling for my Pi. Might be useful to other people too.

```
docker pull golang:1.4.2-cross

docker run --rm -it -v "$GOPATH":/go -w /go/src/github.com/gpavlidi/slangouts golang:1.4.2-cross bash

# no cgo stuff
GOOS=linux GOARCH=arm GOARM=7 CGO_ENABLED=0 go build -v -o ./bin/arm7/slangouts
scp ./bin/arm7/slangouts pi@gataki:~/
scp ~/.slangouts/config.json pi@gataki:~/.slangouts/config.json

```