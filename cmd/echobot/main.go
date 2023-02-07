package main

import (
	"cwtch.im/cwtch/event"
	"cwtch.im/cwtch/model"
	"cwtch.im/cwtch/protocol/connections"
	"git.openprivacy.ca/openprivacy/log"
	"git.openprivacy.ca/sarah/cwtchbot"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"os/user"
	"path"
)

func main() {
	user, _ := user.Current()
	log.SetLevel(log.LevelInfo)
	cwtchbot := bot.NewCwtchBot(path.Join(user.HomeDir, "/.echobot/"), "echobot")

	cwtchbot.Launch()

	log.Infof("echbot address: %v", cwtchbot.Peer.GetOnion())

	for {
		log.Infof("Process.....\n")
		message := cwtchbot.Queue.Next()
		cid, _ := cwtchbot.Peer.FetchConversationInfo(message.Data[event.RemotePeer])
		switch message.EventType {
		case event.NewMessageFromPeer:
			log.Infof("New Event: %v", message)
			cwtchbot.Queue.Publish(event.NewEvent(event.PeerAcknowledgement, map[event.Field]string{event.EventID: message.EventID, event.RemotePeer: message.Data[event.RemotePeer]}))
			msg := cwtchbot.UnpackMessage(message.Data[event.Data])
			log.Infof("Message: %v", msg)
			reply := string(cwtchbot.PackMessage(msg.Overlay, msg.Data))
			cwtchbot.Peer.SendMessage(cid.ID, reply)
		case event.PeerStateChange:
			state := message.Data[event.ConnectionState]
			if state == connections.ConnectionStateName[connections.AUTHENTICATED] {
				log.Infof("Auto approving stranger %v", message.Data[event.RemotePeer])
				// accept the stranger as a new contact
				cwtchbot.Peer.NewContactConversation("stranger", model.DefaultP2PAccessControl(), true)
			}
		default:
			log.Infof("New Event: %v", message)
		}
	}
}
