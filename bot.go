package bot

import (
	"cwtch.im/cwtch/app"
	"cwtch.im/cwtch/app/plugins"
	"cwtch.im/cwtch/event"
	"cwtch.im/cwtch/peer"
	"git.openprivacy.ca/openprivacy/libricochet-go/connectivity"
	"git.openprivacy.ca/openprivacy/libricochet-go/log"
	"path"
	"time"
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

func (cb *CwtchBot) Launch() {
	mn, err := connectivity.StartTor(path.Join(cb.dir, "/.tor"), "./tor")
	if err != nil {
		log.Errorf("\nError connecting to Tor: %v\n", err)
	}
	cb.acn = mn
	cb.acn.WaitTillBootstrapped()
	app := app.NewApp(mn, cb.dir)


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
		time.Sleep(time.Second * 4)
	}
	app.LaunchPeers()

}
