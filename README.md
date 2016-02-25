# Slangouts
Slack up front, Hangouts in the rear.

## Why?
I use Slack daily at work, and lots of friends/family were using Hangouts. Didnt want to keep both open so Slangouts came to the rescue. It acts as a bridge between the 2 keeping them in sync while it runs. This means people talk to you in Hangouts, and you can reply on Slack and vice versa. 

## Configuration
First time your run Slangouts, it 'll ask you for access to both your Slack and Hangout accounts. Follow the on-screen instructions to complete the configuration.

## Usage
You can download slangouts or build it yourself. Pre-built binaries are provided:
- [Windows (64 bits)](https://raw.githubusercontent.com/gpavlidi/slangouts/master/builds/windows_x64/slangouts.exe)
- [Windows (32 bits)](https://raw.githubusercontent.com/gpavlidi/slangouts/master/builds/windows_x86/slangouts.exe)
- [Linux (64 bits)](https://raw.githubusercontent.com/gpavlidi/slangouts/master/builds/linux_x64/slangouts)
- [Linux (32 bits)](https://raw.githubusercontent.com/gpavlidi/slangouts/master/builds/linux_x86/slangouts)
- [Linux ARM7 (e.g. Pi)](https://raw.githubusercontent.com/gpavlidi/slangouts/master/builds/linux_arm7/slangouts)
- [Mac (32 bits)](https://raw.githubusercontent.com/gpavlidi/slangouts/master/builds/mac_x86/slangouts)
- [Mac (64 bits)](https://raw.githubusercontent.com/gpavlidi/slangouts/master/builds/mac_x64/slangouts)
```
# see all available switches
./slangouts help

# example
./slangouts --config ~/.slangouts/config.json --poll 10 
```

## Building/Cross Compiling
Below is mostly for me to keep handy for compiling for my Pi and other platforms. Might be useful to other people too.

```
# pull cross-compile toolchain
docker pull golang:1.4.2-cross

# build all versions
docker run --rm -v "$GOPATH":/go -w /go/src/github.com/gpavlidi/slangouts -e GOOS=darwin -e GOARCH=amd64 -e CGO_ENABLED=0 golang:1.4.2-cross go build -v -o ./builds/mac_x64/slangouts
docker run --rm -v "$GOPATH":/go -w /go/src/github.com/gpavlidi/slangouts -e GOOS=darwin -e GOARCH=386 -e CGO_ENABLED=0 golang:1.4.2-cross go build -v -o ./builds/mac_x86/slangouts
docker run --rm -v "$GOPATH":/go -w /go/src/github.com/gpavlidi/slangouts -e GOOS=windows -e GOARCH=386 -e CGO_ENABLED=0 golang:1.4.2-cross go build -v -o ./builds/windows_x86/slangouts.exe
docker run --rm -v "$GOPATH":/go -w /go/src/github.com/gpavlidi/slangouts -e GOOS=windows -e GOARCH=amd64 -e CGO_ENABLED=0 golang:1.4.2-cross go build -v -o ./builds/windows_x64/slangouts.exe
docker run --rm -v "$GOPATH":/go -w /go/src/github.com/gpavlidi/slangouts -e GOOS=linux -e GOARCH=386 -e CGO_ENABLED=0 golang:1.4.2-cross go build -v -o ./builds/linux_x86/slangouts
docker run --rm -v "$GOPATH":/go -w /go/src/github.com/gpavlidi/slangouts -e GOOS=linux -e GOARCH=amd64 -e CGO_ENABLED=0 golang:1.4.2-cross go build -v -o ./builds/linux_x64/slangouts
docker run --rm -v "$GOPATH":/go -w /go/src/github.com/gpavlidi/slangouts -e GOOS=linux -e GOARCH=arm -e GOARM=7 -e CGO_ENABLED=0 golang:1.4.2-cross go build -v -o ./builds/linux_arm7/slangouts

# need to clean these up every time I rebuild darwin_amd64
go clean -i github.com/nlopes/slack
go clean -i golang.org/x/net/websocket

# to debug cross compiling
docker run --rm -it -v "$GOPATH":/go -w /go/src/github.com/gpavlidi/slangouts golang:1.4.2-cross bash
GOOS=windows GOARCH=386 CGO_ENABLED=0 go build -v -o ./builds/windows_x86/slangouts.exe

# copy over to Pi
scp ./builds/linux_arm7/slangouts pi@gataki:~/
scp ~/.slangouts/config.json pi@gataki:~/.slangouts/config.json

```