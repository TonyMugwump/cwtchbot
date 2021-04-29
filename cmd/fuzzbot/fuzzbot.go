package main

import (
	"crypto/rand"
	"cwtch.im/cwtch/event"
	"cwtch.im/cwtch/model"
	"cwtch.im/cwtch/protocol/connections"
	"encoding/json"
	"git.openprivacy.ca/openprivacy/log"
	"git.openprivacy.ca/sarah/cwtchbot"
	"io/ioutil"
	"os"
	"os/user"
	"path"
)

type BLNS struct {
	inputs []string
}

func main() {
	user, _ := user.Current()
	log.SetLevel(log.LevelDebug)
	cwtchbot := bot.NewCwtchBot(path.Join(user.HomeDir, "/.fuzzbot/"), "fuzzbot")

	cwtchbot.Launch()

	blns := new(BLNS)
	blns_file, err := ioutil.ReadFile("./cmd/fuzzbot/blns.json")
	if err != nil {
		log.Errorf("could not read BLNS file %v", err)
		os.Exit(1)
	}
	var inputs []string
	err = json.Unmarshal(blns_file, &inputs)
	if err != nil {
		log.Errorf("could not decode BLNS file %v", err)
	}
	blns.inputs = inputs

	input := make([]byte, 64)
	_, err = rand.Read(input)
	if err != nil {
		panic(err)
	}
	cwtchbot.Peer.SetName(string(input))

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
			switch msg.Data {
			case "blns":
				{
					reply := string(cwtchbot.PackMessage(msg.Overlay, "Starting the Fuzzing Process..."))
					cwtchbot.Peer.SendMessageToPeer(message.Data[event.RemotePeer], reply)
					for _, input := range blns.inputs {
						reply := string(cwtchbot.PackMessage(msg.Overlay, input))
						cwtchbot.Peer.SendMessageToPeer(message.Data[event.RemotePeer], reply)
					}
				}
			case "random-overlay":
				{
					reply := string(cwtchbot.PackMessage(msg.Overlay, "Starting the Fuzzing Process..."))
					cwtchbot.Peer.SendMessageToPeer(message.Data[event.RemotePeer], reply)
					for i := 0; i < 100; i++ {
						input := make([]byte, 64)
						_, err := rand.Read(input)
						if err != nil {
							panic(err)
						}
						reply := string(cwtchbot.PackMessage(msg.Overlay, string(input)))
						cwtchbot.Peer.SendMessageToPeer(message.Data[event.RemotePeer], reply)
					}
				}
			case "random":
				{
					reply := string(cwtchbot.PackMessage(msg.Overlay, "Starting the Fuzzing Process..."))
					cwtchbot.Peer.SendMessageToPeer(message.Data[event.RemotePeer], reply)
					for i := 0; i < 100; i++ {
						input := make([]byte, 64)
						_, err := rand.Read(input)
						if err != nil {
							panic(err)
						}
						reply := string(input)
						cwtchbot.Peer.SendMessageToPeer(message.Data[event.RemotePeer], reply)
					}
				}
			case "fuzz-peer-details":
			}
		case event.PeerStateChange:
			state := message.Data[event.ConnectionState]
			if state == connections.ConnectionStateName[connections.AUTHENTICATED] {
				log.Infof("Auto approving stranger %v", message.Data[event.RemotePeer])
				cwtchbot.Peer.AddContact("stranger", message.Data[event.RemotePeer], model.AuthApproved)
			}
		case event.NewGetValMessageFromPeer:

		default:
			log.Infof("New Event: %v", message)
		}
	}
}
