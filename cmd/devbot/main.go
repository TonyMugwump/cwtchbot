package main

import (
	"cwtch.im/cwtch/event"
	"cwtch.im/cwtch/model"
	"cwtch.im/cwtch/protocol/connections"
	"encoding/json"
	"fmt"
	"git.openprivacy.ca/openprivacy/log"
	"git.openprivacy.ca/sarah/cwtchbot"
	"github.com/araddon/dateparse"
	"math/rand"
	"os/user"
	"path"
	"strings"
	"time"
)

var cwtchbot *bot.CwtchBot

type OverlayEnvelope struct {
	onion string
	Overlay int `json:"o"`
	Data string `json:"d"`
}

func Unwrap(onion, msg string) *OverlayEnvelope {
	var envelope OverlayEnvelope
	err := json.Unmarshal([]byte(msg), &envelope)
	if err != nil {
		log.Errorf("json error: %v", err)
		return nil
	}
	envelope.onion = onion
	return &envelope
}

func (this *OverlayEnvelope) reply(msg string) {
	retenv := OverlayEnvelope{Overlay:1, Data:msg}
	raw, _ := json.Marshal(retenv)
	log.Debugf("sending %v to %v", string(raw), this.onion)
	cwtchbot.Peer.SendMessageToPeer(this.onion, string(raw))
}

func (this *OverlayEnvelope) spam() {
	for {
		this.reply(fmt.Sprintf("%d", rand.Int()))
	}
}

func helpMessage() string {
	return "help\nevery\nin\nat\nspam\nstop"
}

func main() {
	user, _ := user.Current()
	log.SetLevel(log.LevelInfo)
	cwtchbot = bot.NewCwtchBot(path.Join(user.HomeDir, "/.echobot/"), "echobot")

	cwtchbot.Launch()

	for {
		log.Infof("Process.....\n")
		message := cwtchbot.Queue.Next()
		switch message.EventType {
		case event.NewMessageFromGroup:
			if message.Data[event.RemotePeer] != cwtchbot.Peer.GetOnion() {
				log.Infof("New Message: %v\v", message.Data[event.Data])
				cwtchbot.Peer.SendMessageToGroupTracked(message.Data[event.GroupID], message.Data[event.Data])
			}
		case event.NewMessageFromPeer:
			log.Infof("New Event: %v", message)
			cwtchbot.Queue.Publish(event.NewEvent(event.PeerAcknowledgement, map[event.Field]string{event.EventID: message.EventID, event.RemotePeer: message.Data[event.RemotePeer]}))

			envelope := Unwrap(message.Data[event.RemotePeer], message.Data[event.Data])
			mainTimer := time.NewTimer(time.Nanosecond)

			if envelope.Overlay == 1 {
				cmd := strings.Split(envelope.Data, " ")
				switch cmd[0] {
				case "help":
					envelope.reply(helpMessage())
				case "every":
					interval, err := time.ParseDuration(cmd[1])
					if err != nil {
						envelope.reply(fmt.Sprintf("parse error: %s", err))
						continue
					}
					envelope.reply("you got it!")
					mainTimer.Stop()
					mainTimer = time.AfterFunc(interval, func() {
						envelope.reply(cmd[2])
						mainTimer.Reset(interval)
					})
				case "in":
					interval, err := time.ParseDuration(cmd[1])
					if err != nil {
						envelope.reply(fmt.Sprintf("parse error: %s", err))
						continue
					}
					envelope.reply("will do!")
					mainTimer.Stop()
					mainTimer = time.AfterFunc(interval, func() {
						envelope.reply(cmd[2])
					})
				case "at":
					at, err := dateparse.ParseAny(cmd[1])
					if err != nil {
						envelope.reply(fmt.Sprintf("parse error: %s", err))
						continue
					}
					envelope.reply(fmt.Sprintf("ok, sending at %v", at))

					mainTimer.Stop()
					interval := time.Until(at)
					time.AfterFunc(interval, func() {
						envelope.reply(cmd[2])
					})
				case "spam":
					envelope.reply("lol ok you asked for it!")
					mainTimer.Stop()
					mainTimer = time.AfterFunc(time.Nanosecond, func() {
						envelope.reply(fmt.Sprintf("%d", rand.Int()))
						mainTimer.Reset(time.Nanosecond)
					})
				default:
					envelope.reply("unrecognized command")
				}
			} else {
				log.Warnf("unknown overlay type %d", envelope.Overlay)
			}
		case event.PeerStateChange:
			state := message.Data[event.ConnectionState]
			if state == connections.ConnectionStateName[connections.AUTHENTICATED] {
				log.Infof("Auto approving stranger %v", message.Data[event.RemotePeer])
				cwtchbot.Peer.AddContact("stranger", message.Data[event.RemotePeer], model.AuthApproved)
			}
		default:
			log.Infof("New Event: %v", message)
		}
	}
}
