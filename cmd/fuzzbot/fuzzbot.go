package main

import (
	"crypto/rand"
	"crypto/sha256"
	"cwtch.im/cwtch/event"
	"cwtch.im/cwtch/functionality/filesharing"
	"cwtch.im/cwtch/model"
	"cwtch.im/cwtch/model/attr"
	"cwtch.im/cwtch/model/constants"
	"cwtch.im/cwtch/protocol/connections"
	"cwtch.im/cwtch/protocol/files"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"git.openprivacy.ca/openprivacy/log"
	"git.openprivacy.ca/sarah/cwtchbot"
	_ "github.com/mutecomm/go-sqlcipher/v4"
	"io"
	"io/ioutil"
	"math/big"
	"os"
	"os/user"
	"path"
	"strings"
	"time"
)

type BLNS struct {
	inputs []string
}

func main() {
	user, _ := user.Current()
	log.SetLevel(log.LevelInfo)
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
	cwtchbot.Peer.SetScopedZonedAttribute(attr.LocalScope, attr.ProfileZone, constants.Name, "fuzzbot")

	// Will currently only work on Sarah's custom fork (testing custom profile images)
	fh := new(filesharing.Functionality)
	fileKey, _, err := fh.ShareFile("./fuzzbot.png", cwtchbot.Peer)
	log.Errorf("sharing file: %v %v", fileKey, err)
	const CustomProfileImageKey = "custom-profile-image"
	cwtchbot.Peer.SetScopedZonedAttribute(attr.PublicScope, attr.ProfileZone, CustomProfileImageKey, fileKey)

	// Create a group for this session:
	// group, invite := cwtchbot.Peer.StartGroup("ur33edbwvbevcls5ue6jpkoubdptgkgl5bedzfyau2ibf5276lyp4uid")

	// fmt.Printf("invite: %v", invite)

	for {
		log.Infof("Process.....\n")
		message := cwtchbot.Queue.Next()
		switch message.EventType {
		case event.NewMessageFromPeer:
			log.Infof("New Event: %v", message)
			cwtchbot.Queue.Publish(event.NewEvent(event.PeerAcknowledgement, map[event.Field]string{event.EventID: message.EventID, event.RemotePeer: message.Data[event.RemotePeer]}))
			msg := cwtchbot.UnpackMessage(message.Data[event.Data])
			log.Infof("Message: %v", msg)
			command := strings.Split(msg.Data, " ")
			cid, _ := cwtchbot.Peer.FetchConversationInfo(message.Data[event.RemotePeer])
			switch command[0] {
			case "blns":
				{
					reply := string(cwtchbot.PackMessage(msg.Overlay, "Starting the Fuzzing Process..."))
					cwtchbot.Peer.SendMessage(cid.ID, reply)
					for _, input := range blns.inputs {
						reply := string(cwtchbot.PackMessage(msg.Overlay, input))
						cwtchbot.Peer.SendMessage(cid.ID, reply)
					}
				}
			case "blns-mutate":
				{
					reply := string(cwtchbot.PackMessage(msg.Overlay, "Starting the BLNS Mutation Process..."))
					cwtchbot.Peer.SendMessage(cid.ID, reply)
					for _, input := range blns.inputs {
						for i := 0; i < 5; i++ {
							reply := string(cwtchbot.PackMessage(msg.Overlay, mutate(input)))
							cwtchbot.Peer.SendMessage(cid.ID, reply)
							time.Sleep(time.Millisecond * 50)
						}
					}
				}
			case "random-overlay":
				{
					reply := string(cwtchbot.PackMessage(msg.Overlay, "Starting the Fuzzing Process..."))
					cwtchbot.Peer.SendMessage(cid.ID, reply)
					for i := 0; i < 100; i++ {
						input := make([]byte, 64)
						_, err := rand.Read(input)
						if err != nil {
							panic(err)
						}
						reply := string(cwtchbot.PackMessage(int(input[0]), string(input)))
						cwtchbot.Peer.SendMessage(cid.ID, reply)
					}
				}
			case "random":
				{
					reply := string(cwtchbot.PackMessage(msg.Overlay, "Starting the Fuzzing Process..."))
					cwtchbot.Peer.SendMessage(cid.ID, reply)
					for i := 0; i < 100; i++ {
						input := make([]byte, 64)
						_, err := rand.Read(input)
						if err != nil {
							panic(err)
						}
						reply := string(input)
						cwtchbot.Peer.SendMessage(cid.ID, reply)
					}
				}
			case "quoteme":
				hashSum := sha256.Sum256([]byte(message.Data[event.RemotePeer] + message.Data[event.Data]))
				contentHash := base64.StdEncoding.EncodeToString(hashSum[:])
				reply := string(cwtchbot.PackMessage(10, `{"quotedHash":"`+contentHash+`","body":"quoted for you"}`))
				cwtchbot.Peer.SendMessage(cid.ID, reply)
			case "quoteme-evil":
				hashSum := sha256.Sum256([]byte(message.Data[event.RemotePeer] + message.Data[event.Data]))
				contentHash := base64.StdEncoding.EncodeToString(hashSum[:])
				reply := string(cwtchbot.PackMessage(10, `{"quotedHash":"`+contentHash+`","body":"quoted for you"}`))
				cwtchbot.Peer.SendMessage(cid.ID, mutate(reply))
			case "help":
				reply := string(cwtchbot.PackMessage(msg.Overlay, "Fuzzing commands: [blns, invite-me]"))
				cwtchbot.Peer.SendMessage(cid.ID, reply)
				reply = string(cwtchbot.PackMessage(msg.Overlay, "Cwtch Testing Group Invite: [testgroup-invite]"))
				cwtchbot.Peer.SendMessage(cid.ID, reply)
			case "slow":
				for i := 0; i < 10; i++ {
					reply := string(cwtchbot.PackMessage(msg.Overlay, "Fuzzing commands: [blns, invite-me]"))
					cwtchbot.Peer.SendMessage(cid.ID, mutate(reply))
					time.Sleep(time.Second * 2)
				}
			case "sharefile":
				for i := 0; i < 100; i++ {
					manifest, _ := files.CreateManifest("./README.md")

					var nonce [24]byte
					if _, err := io.ReadFull(rand.Reader, nonce[:]); err != nil {
						log.Errorf("Cannot read from random: %v\n", err)
					}

					message := filesharing.OverlayMessage{
						Name:  path.Base(manifest.FileName),
						Hash:  hex.EncodeToString(manifest.RootHash),
						Nonce: hex.EncodeToString(nonce[:]),
						Size:  manifest.FileSizeInBytes,
					}

					data, _ := json.Marshal(message)

					wrapper := model.MessageWrapper{
						Overlay: model.OverlayFileSharing,
						Data:    string(data),
					}
					wrapperJSON, _ := json.Marshal(wrapper)
					cwtchbot.Peer.SendMessage(cid.ID, mutate(string(wrapperJSON)))
				}

			case "fuzz-peer-details":
				break
			case "testgroup-invite":
				reply := string(cwtchbot.PackMessage(101, "tofubundle:server:eyJLZXlzIjp7ImJ1bGxldGluX2JvYXJkX29uaW9uIjoiaXNicjJ0NmJmbHVsMnp5aTZoanRudWV6YjJ4dmZyNDJzdnpqZzJxM2d5cWZnZzN3bW5yYmtrcWQiLCJwcml2YWN5X3Bhc3NfcHVibGljX2tleSI6Ik1JWC93L2VKeHQ4TTZMRW5TNnU1MStFQTVUNFVZY3VIZ3d6TElrYkhkeVk9IiwidG9rZW5fc2VydmljZV9vbmlvbiI6ImxpNTNxNmp1YWZ1NGF2cjdydHlsdG1zcTJ1anl5N3NjcnIzZnRua3JsaWNzeGV3Njd4cTY0c3lkIn0sIlNpZ25hdHVyZSI6IjIvTWw3T09HK2FYSFh2NTFkU2xJRHQxZjUxK1VUUmRTWnNFSHVxYlRqc3N4alZ5Qm1RUm1QU0xWSnZKUXBwS2cvZ1N0MzZrWVJKNXl1WWxEUDhzQ0NBPT0ifQ==||torv3eyJHcm91cElEIjoiOTQwYTc5OGI4MjY4YzI1Yjg0ZmMzYThlNWFhM2RiMzkiLCJHcm91cE5hbWUiOiJDd3RjaCBSZWxlYXNlIENhbmRpZGF0ZSBUZXN0ZXJzISIsIlNpZ25lZEdyb3VwSUQiOm51bGwsIlRpbWVzdGFtcCI6MCwiU2hhcmVkS2V5IjoiS3lmT2F6YzJuNUZyS1AzYzV5allheTZpVEN5TXhKQUhrT29YVWpSV3k4QT0iLCJTZXJ2ZXJIb3N0IjoiaXNicjJ0NmJmbHVsMnp5aTZoanRudWV6YjJ4dmZyNDJzdnpqZzJxM2d5cWZnZzN3bW5yYmtrcWQifQ=="))
				cwtchbot.Peer.SendMessage(cid.ID, reply)
			case "invite-me":

				//num := 1
				//if len(command) >= 2 {
				//    num, _ = strconv.Atoi(command[1])
				//}
				//
				//for i := 0; i < num; i++ {
				//    randIndex, _ := rand.Int(rand.Reader, big.NewInt(int64(len(blns.inputs))))
				//    cwtchbot.Peer.SetGroupAttribute(group, "local.name", mutate(blns.inputs[randIndex.Uint64()]))
				//    group := cwtchbot.Peer.GetGroup(group)
				//    randIndex, _ = rand.Int(rand.Reader, big.NewInt(int64(len(blns.inputs))))
				//    group.GroupID = mutate(blns.inputs[randIndex.Uint64()])
				//    invite, _ := group.Invite()
				//    inviteMessage := cwtchbot.PackMessage(101, fmt.Sprintf("tofubundle:server:%s||%s", "eyJLZXlzIjp7ImJ1bGxldGluX2JvYXJkX29uaW9uIjoidXIzM2VkYnd2YmV2Y2xzNXVlNmpwa291YmRwdGdrZ2w1YmVkemZ5YXUyaWJmNTI3Nmx5cDR1aWQiLCJwcml2YWN5X3Bhc3NfcHVibGljX2tleSI6Iml2UnNSOUNpMGdqWHhjTk5LSVVqOTdwQU1rdndhV1Vta25WMnlOU3lWQ2c9IiwidG9rZW5fc2VydmljZV9vbmlvbiI6ImN4ang1c3Izb3AyaTZoanJqc2Z6amJ1ZWZoaXlxM3RlbDV1bHhuYmoyNnZ0dm9ycGhsZW1zbGlkIn0sIlNpZ25hdHVyZSI6IktDckxGZ3QxZU1KYnptOS9wUWZxY1F5a3lBVU5hV1FKQnlTRTdIdXc5N2NZTHlXYmR0SGxSVWx4VG1hK3JMMVcybTNQOTRrVEszclFnZi9XUjhiTkRRPT0ifQ==", invite))
				//    //cwtchbot.Peer.SendMessageToPeer(message.Data[event.RemotePeer], string(cwtchbot.PackMessage(msg.Overlay, fmt.Sprintf("tofubundle:server:%s||torv3%s",  "eyJLZXlzIjp7ImJ1bGxldGluX2JvYXJkX29uaW9uIjoidXIzM2VkYnd2YmV2Y2xzNXVlNmpwa291YmRwdGdrZ2w1YmVkemZ5YXUyaWJmNTI3Nmx5cDR1aWQiLCJwcml2YWN5X3Bhc3NfcHVibGljX2tleSI6Iml2UnNSOUNpMGdqWHhjTk5LSVVqOTdwQU1rdndhV1Vta25WMnlOU3lWQ2c9IiwidG9rZW5fc2VydmljZV9vbmlvbiI6ImN4ang1c3Izb3AyaTZoanJqc2Z6amJ1ZWZoaXlxM3RlbDV1bHhuYmoyNnZ0dm9ycGhsZW1zbGlkIn0sIlNpZ25hdHVyZSI6IktDckxGZ3QxZU1KYnptOS9wUWZxY1F5a3lBVU5hV1FKQnlTRTdIdXc5N2NZTHlXYmR0SGxSVWx4VG1hK3JMMVcybTNQOTRrVEszclFnZi9XUjhiTkRRPT0ifQ==", base64.StdEncoding.EncodeToString(invite)))))
				//    cwtchbot.Peer.SendMessage(cid, string(inviteMessage))
				//}
			}
		case event.PeerStateChange:
			state := message.Data[event.ConnectionState]
			if state == connections.ConnectionStateName[connections.AUTHENTICATED] {
				log.Infof("Auto approving stranger %v", message.Data[event.RemotePeer])
				cwtchbot.Peer.NewContactConversation(message.Data[event.RemotePeer], model.DefaultP2PAccessControl(), true)
			}

		default:
			log.Infof("New Event: %v", message)
		}
	}
}

// mutate is a very basic string mutator that simply garbles a random byte. We've got no success conditions
// to feed back to the mutator so we need to rely on a larger corpus, custom injection and simple mutations.
func mutate(input string) string {
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

func randomString() string {
	input := make([]byte, 64)
	_, err := rand.Read(input)
	if err != nil {
		panic(err)
	}
	return string(input)
}
