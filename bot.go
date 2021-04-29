package bot

import (
	"crypto/rand"
	"cwtch.im/cwtch/app"
	"cwtch.im/cwtch/app/plugins"
	"cwtch.im/cwtch/event"
	"cwtch.im/cwtch/peer"
	"encoding/base64"
	"encoding/json"
	"git.openprivacy.ca/openprivacy/connectivity"
	"git.openprivacy.ca/openprivacy/connectivity/tor"
	"git.openprivacy.ca/openprivacy/log"
	"os"
	"path"
	"path/filepath"
	"time"
	mrand "math/rand"
)

type CwtchBot struct {
	dir   string
	Peer  peer.CwtchPeer
	Queue event.Queue
	acn   connectivity.ACN
	peername string
}

func NewCwtchBot(userdir string, peername string) *CwtchBot {
	cb := new(CwtchBot)
	cb.dir = userdir
	cb.peername = peername
	return cb
}

type MessageWrapper struct {
	Overlay int `json:"o"`
	Data string `json:"d"`
}

func (cb * CwtchBot) PackMessage(overlay int, message string) []byte {
	mw := new(MessageWrapper)
	mw.Overlay = overlay
	mw.Data = message
	data,_ := json.Marshal(mw)
	return data
}

func (cb * CwtchBot) UnpackMessage(message string) MessageWrapper {
	mw := new(MessageWrapper)
	json.Unmarshal([]byte(message), mw)
	return *mw
}

func (cb *CwtchBot) Launch() {
	mrand.Seed(int64(time.Now().Nanosecond()))
	port := mrand.Intn(1000) + 9600
	controlPort := port + 1

	// generate a random password (actually random, stored in memory, for the control port)
	key := make([]byte, 64)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}

	log.Infof("making directory %v", cb.dir)
	os.MkdirAll(path.Join(cb.dir, "/.tor","tor"),0700)
	tor.NewTorrc().WithSocksPort(port).WithOnionTrafficOnly().WithControlPort(controlPort).WithHashedPassword(base64.StdEncoding.EncodeToString(key)).Build(filepath.Join(cb.dir, ".tor", "tor", "torrc"))
	cb.acn, err = tor.NewTorACNWithAuth(path.Join(cb.dir, "/.tor"), "", controlPort, tor.HashedPasswordAuthenticator{base64.StdEncoding.EncodeToString(key)})
	if err != nil {
		log.Errorf("\nError connecting to Tor: %v\n", err)
	}
	cb.acn.WaitTillBootstrapped()
	app := app.NewApp(cb.acn, cb.dir)


	app.LoadProfiles("")
	if len(app.ListPeers()) == 0 {
		app.CreatePeer(cb.peername, "")
	}

	peers := app.ListPeers()
	for onion, _ := range peers {
		app.AddPeerPlugin(onion, plugins.CONNECTIONRETRY)
		cb.Peer = app.GetPeer(onion)
		log.Infof("Running %v", onion)
		cb.Queue = event.NewQueue()
		eb := app.GetEventBus(onion)
		eb.Subscribe(event.NewMessageFromPeer, cb.Queue)
		eb.Subscribe(event.PeerAcknowledgement, cb.Queue)
		eb.Subscribe(event.NewMessageFromGroup, cb.Queue)
		eb.Subscribe(event.NewGroupInvite, cb.Queue)
		eb.Subscribe(event.SendMessageToGroupError, cb.Queue)
		eb.Subscribe(event.SendMessageToPeerError, cb.Queue)
		eb.Subscribe(event.ServerStateChange, cb.Queue)
		eb.Subscribe(event.PeerStateChange, cb.Queue)
		eb.Subscribe(event.NewGetValMessageFromPeer,cb.Queue)
		time.Sleep(time.Second * 4)
	}
	app.LaunchPeers()

}
