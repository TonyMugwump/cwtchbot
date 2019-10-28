package echobot

import (
	"cwtch.im/cwtch/event"
	"git.openprivacy.ca/openprivacy/libricochet-go/log"
	"git.openprivacy.ca/sarah/cwtchbot"
	"os/user"
	"path"
)

func main() {
	user, _ := user.Current()
	log.SetLevel(log.LevelInfo)
	cwtchbot := bot.NewCwtchBot(path.Join(user.HomeDir, "/.echobot/"), "echobot")
	cwtchbot.Launch()

	for {
		log.Infof("Process.....\n")
		message := cwtchbot.Queue.Next()
		switch message.EventType {
		case event.NewMessageFromGroup:
			if message.Data[event.RemotePeer] != cwtchbot.Peer.GetProfile().Onion {
				log.Infof("New Message: %v\v", message.Data[event.Data])
				cwtchbot.Peer.SendMessageToGroup(message.Data[event.GroupID], message.Data[event.Data])
			}
		default:
			log.Infof("New Event: %v", message)
		}
	}
}
