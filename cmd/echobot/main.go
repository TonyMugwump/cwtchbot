package main

import (
	"cwtch.im/cwtch/event"
	"cwtch.im/cwtch/model"
	"cwtch.im/cwtch/protocol/connections"
	"git.openprivacy.ca/openprivacy/log"
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
			if message.Data[event.RemotePeer] != cwtchbot.Peer.GetOnion() {
				log.Infof("New Message: %v\v", message.Data[event.Data])
				cwtchbot.Peer.SendMessageToGroupTracked(message.Data[event.GroupID], message.Data[event.Data])
			}
		case event.NewMessageFromPeer:
			log.Infof("New Event: %v", message)
			cwtchbot.Queue.Publish(event.NewEvent(event.PeerAcknowledgement, map[event.Field]string{event.EventID: message.EventID, event.RemotePeer: message.Data[event.RemotePeer]}))
			msg := cwtchbot.UnpackMessage(message.Data[event.Data])
			log.Infof("Message: %v", msg)
			reply := string(cwtchbot.PackMessage(msg.Overlay, msg.Data))
			cwtchbot.Peer.SendMessageToPeer(message.Data[event.RemotePeer], reply)
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
