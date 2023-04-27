package bot

import (
	"crypto/rand"
	"cwtch.im/cwtch/app"
	"cwtch.im/cwtch/event"
	"cwtch.im/cwtch/peer"
	"cwtch.im/cwtch/protocol/connections"
	"cwtch.im/cwtch/settings"
	"encoding/base64"
	"encoding/json"
	"git.openprivacy.ca/openprivacy/connectivity"
	"git.openprivacy.ca/openprivacy/connectivity/tor"
	"git.openprivacy.ca/openprivacy/log"
	mrand "math/rand"
	"os"
	"path"
	"path/filepath"
	"time"
)

type CwtchBot struct {
	dir         string
	Peer        peer.CwtchPeer
	Queue       event.Queue
	acn         connectivity.ACN
	peername    string
	engineHooks connections.EngineHooks
	experiments []string
}

func NewCwtchBot(userdir string, peername string) *CwtchBot {
	cb := new(CwtchBot)
	cb.dir = userdir
	cb.peername = peername
	cb.engineHooks = connections.DefaultEngineHooks{}
	cb.experiments = nil
	return cb
}

func NewCwtchBotWithExperiments(userdir string, peername string, experiments []string) *CwtchBot {
	cb := new(CwtchBot)
	cb.dir = userdir
	cb.peername = peername
	cb.engineHooks = connections.DefaultEngineHooks{}
	cb.experiments = experiments
	return cb
}

func (cb *CwtchBot) HookEngine(hooks connections.EngineHooks) {
	cb.engineHooks = hooks
}

type MessageWrapper struct {
	Overlay int    `json:"o"`
	Data    string `json:"d"`
}

func (cb *CwtchBot) PackMessage(overlay int, message string) []byte {
	mw := new(MessageWrapper)
	mw.Overlay = overlay
	mw.Data = message
	data, _ := json.Marshal(mw)
	return data
}

func (cb *CwtchBot) UnpackMessage(message string) MessageWrapper {
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
	os.MkdirAll(path.Join(cb.dir, "/.tor", "tor"), 0700)
	tor.NewTorrc().WithSocksPort(port).WithOnionTrafficOnly().WithControlPort(controlPort).WithHashedPassword(base64.StdEncoding.EncodeToString(key)).Build(filepath.Join(cb.dir, ".tor", "tor", "torrc"))
	cb.acn, err = tor.NewTorACNWithAuth(path.Join(cb.dir, "/.tor"), "", path.Join(cb.dir, "/.tor", "data"), controlPort, tor.HashedPasswordAuthenticator{base64.StdEncoding.EncodeToString(key)})
	if err != nil {
		log.Errorf("\nError connecting to Tor: %v\n", err)
	}
	cb.acn.WaitTillBootstrapped()
	settingsFile, _ := settings.InitGlobalSettingsFile(cb.dir, "")
	gSettings := settingsFile.ReadGlobalSettings()

	if cb.experiments != nil {
		gSettings.ExperimentsEnabled = true
	} else {
		gSettings.ExperimentsEnabled = false
	}
	gSettings.DownloadPath = "./"

	// Reset all Experiments...
	for experiment := range gSettings.Experiments {
		gSettings.Experiments[experiment] = false
	}

	// Explicitly Enable only the experiments we've specified...
	for _, experiment := range cb.experiments {
		gSettings.Experiments[experiment] = true
	}

	settingsFile.WriteGlobalSettings(gSettings)
	app := app.NewApp(cb.acn, cb.dir, settingsFile)

	app.InstallEngineHooks(cb.engineHooks)

	app.LoadProfiles("")
	if len(app.ListProfiles()) == 0 {
		app.CreateProfile(cb.peername, "", true)
	}

	peers := app.ListProfiles()

	for _, onion := range peers {
		cb.Peer = app.GetPeer(onion)
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
		eb.Subscribe(event.NewGetValMessageFromPeer, cb.Queue)
		eb.Subscribe(event.ContactCreated, cb.Queue)
	}
	app.ActivateEngines(true, true, true)

}
