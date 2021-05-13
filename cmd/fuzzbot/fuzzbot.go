package main

import (
	"crypto/rand"
	"cwtch.im/cwtch/event"
	"cwtch.im/cwtch/model"
	"cwtch.im/cwtch/protocol/connections"
	"encoding/json"
	"fmt"
	"git.openprivacy.ca/openprivacy/log"
	"git.openprivacy.ca/sarah/cwtchbot"
	"io/ioutil"
	"math/big"
	"os"
	"os/user"
	"path"
	"strconv"
	"strings"
	"time"
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
	cwtchbot.Peer.SetAttribute("public.name", string(input))

	// Will currently only work on Sarah's custom fork (testing custom profile images)
	cwtchbot.Peer.SetAttribute("public.picture", "profiles/fuzzbot.png")

	// Create a group for this session:
	group,_,_ := cwtchbot.Peer.StartGroup("ur33edbwvbevcls5ue6jpkoubdptgkgl5bedzfyau2ibf5276lyp4uid")

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
			command := strings.Split(msg.Data, " ")
			switch command[0] {
			case "blns":
				{
					reply := string(cwtchbot.PackMessage(msg.Overlay, "Starting the Fuzzing Process..."))
					cwtchbot.Peer.SendMessageToPeer(message.Data[event.RemotePeer], reply)
					for _, input := range blns.inputs {
						reply := string(cwtchbot.PackMessage(msg.Overlay, input))
						cwtchbot.Peer.SendMessageToPeer(message.Data[event.RemotePeer], reply)
					}
				}
			case "blns-mutate":
				{
					reply := string(cwtchbot.PackMessage(msg.Overlay, "Starting the BLNS Mutation Process..."))
					cwtchbot.Peer.SendMessageToPeer(message.Data[event.RemotePeer], reply)
					for _, input := range blns.inputs {
						for i :=0; i< 5; i++ {
								reply := string(cwtchbot.PackMessage(msg.Overlay, mutate(input)))
								cwtchbot.Peer.SendMessageToPeer(message.Data[event.RemotePeer], reply)
								time.Sleep(time.Millisecond * 50)
						}
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
						reply := string(cwtchbot.PackMessage(int(input[0]), string(input)))
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
				break
			case "invite-me":

				num := 1
				if len(command) >= 2 {
					num,_ = strconv.Atoi(command[1])
				}

				for i:=0; i < num; i++ {
					randIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(blns.inputs))))
					cwtchbot.Peer.SetGroupAttribute(group, "local.name", mutate(blns.inputs[randIndex.Uint64()]))
					group := cwtchbot.Peer.GetGroup(group)
					randIndex, _ = rand.Int(rand.Reader, big.NewInt(int64(len(blns.inputs))))
					group.GroupID =  mutate(blns.inputs[randIndex.Uint64()])
					invite, _ := group.Invite()
					inviteMessage := cwtchbot.PackMessage(101, fmt.Sprintf("tofubundle:server:%s||%s", "eyJLZXlzIjp7ImJ1bGxldGluX2JvYXJkX29uaW9uIjoidXIzM2VkYnd2YmV2Y2xzNXVlNmpwa291YmRwdGdrZ2w1YmVkemZ5YXUyaWJmNTI3Nmx5cDR1aWQiLCJwcml2YWN5X3Bhc3NfcHVibGljX2tleSI6Iml2UnNSOUNpMGdqWHhjTk5LSVVqOTdwQU1rdndhV1Vta25WMnlOU3lWQ2c9IiwidG9rZW5fc2VydmljZV9vbmlvbiI6ImN4ang1c3Izb3AyaTZoanJqc2Z6amJ1ZWZoaXlxM3RlbDV1bHhuYmoyNnZ0dm9ycGhsZW1zbGlkIn0sIlNpZ25hdHVyZSI6IktDckxGZ3QxZU1KYnptOS9wUWZxY1F5a3lBVU5hV1FKQnlTRTdIdXc5N2NZTHlXYmR0SGxSVWx4VG1hK3JMMVcybTNQOTRrVEszclFnZi9XUjhiTkRRPT0ifQ==", invite))
					//cwtchbot.Peer.SendMessageToPeer(message.Data[event.RemotePeer], string(cwtchbot.PackMessage(msg.Overlay, fmt.Sprintf("tofubundle:server:%s||torv3%s",  "eyJLZXlzIjp7ImJ1bGxldGluX2JvYXJkX29uaW9uIjoidXIzM2VkYnd2YmV2Y2xzNXVlNmpwa291YmRwdGdrZ2w1YmVkemZ5YXUyaWJmNTI3Nmx5cDR1aWQiLCJwcml2YWN5X3Bhc3NfcHVibGljX2tleSI6Iml2UnNSOUNpMGdqWHhjTk5LSVVqOTdwQU1rdndhV1Vta25WMnlOU3lWQ2c9IiwidG9rZW5fc2VydmljZV9vbmlvbiI6ImN4ang1c3Izb3AyaTZoanJqc2Z6amJ1ZWZoaXlxM3RlbDV1bHhuYmoyNnZ0dm9ycGhsZW1zbGlkIn0sIlNpZ25hdHVyZSI6IktDckxGZ3QxZU1KYnptOS9wUWZxY1F5a3lBVU5hV1FKQnlTRTdIdXc5N2NZTHlXYmR0SGxSVWx4VG1hK3JMMVcybTNQOTRrVEszclFnZi9XUjhiTkRRPT0ifQ==", base64.StdEncoding.EncodeToString(invite)))))
					cwtchbot.Peer.SendMessageToPeer(message.Data[event.RemotePeer], string(inviteMessage))
				}
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


// mutate is a very basic string mutator that simply garbles a random byte. We've got no success conditions
// to feed back to the mutator so we need to rely on a larger corpus, custom injection and simple mutations.
func mutate (input string) string {
	if len(input) > 0 {
		randByte, _ := rand.Int(rand.Reader, big.NewInt(int64(len(input)+1)))
		randMask, _ := rand.Int(rand.Reader, big.NewInt(255))
		// zero indexed...
		index := randByte.Uint64()
		mutatedInput := input
		if index < uint64(len(input)) {
			mutatedInput = input[:index]
			mutatedInput = string(append([]byte(mutatedInput), input[index]^uint8(randMask.Uint64())))
			if index+1 <= uint64(len(input)) {
				mutatedInput = string(append([]byte(mutatedInput), input[index+1:]...))
			}
			return mutatedInput
		}
	}
	return input
}