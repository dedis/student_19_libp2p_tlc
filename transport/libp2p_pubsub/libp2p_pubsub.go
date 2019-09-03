package libp2p_pubsub

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"github.com/dedis/student_19_libp2p_tlc/transport/libp2p_pubsub/protobuf"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	math_rand "math/rand"
	"strings"
	"time"

	"github.com/dedis/student_19_libp2p_tlc/model"

	"github.com/btcsuite/btcd/btcec"
	"github.com/libp2p/go-libp2p"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/crypto"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	quic "github.com/libp2p/go-libp2p-quic-transport"
	ws "github.com/libp2p/go-ws-transport"
)

const delayBias = 100
const delayRange = 1000

var Delayed = true

const BufferLen = 500

type libp2pPubSub struct {
	pubsub       *pubsub.PubSub       // PubSub of each individual node
	subscription *pubsub.Subscription // Subscription of individual node
	topic        string               // PubSub topic
	victim       bool                 // Flag for declaring a node as a victim
	buffer       chan model.Message   // A buffer for keeping received message in case the node is kept in the delayed set by adversary
	group        []int
}

// Broadcast Uses PubSub publish to broadcast messages to other peers
func (c *libp2pPubSub) Broadcast(msg model.Message) {
	// Broadcasting to a topic in PubSub
	msgBytes, err := proto.Marshal(ConvertModelMessage(msg))
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}

	// Send the message with a delay in order to prevent message loss in libp2p
	go func(msgBytes []byte, topic string, pubsub *pubsub.PubSub) {
		if Delayed {
			time.Sleep(time.Duration(delayBias+math_rand.Intn(delayRange)) * time.Millisecond)
		}

		err := pubsub.Publish(topic, msgBytes)
		if err != nil {
			fmt.Printf("Error(((( : %v\n", err)
			return
		}
	}(msgBytes, c.topic, c.pubsub)
}

// Send uses Broadcast for sending messages
func (c *libp2pPubSub) Send(msg model.Message, id int) {
	// In libp2p implementation, we also broadcast instead of sending directly. So Acks will be broadcast in this case.
	c.Broadcast(msg)
}

// Receive gets message from PubSub in a blocking way
func (c *libp2pPubSub) Receive() *model.Message {
	// Check buffer for existing messages
	if !c.victim {
		select {
		case msgFromBuffer := <-c.buffer:
			return &msgFromBuffer
		default:

		}
	}

	// Blocking function for consuming newly received messages
	// We can access message here
	msg, err := c.subscription.Next(context.Background())
	// handling canceled subscriptions
	if err != nil {
		return nil
	}

	msgBytes := msg.Data
	var pbMessage protobuf.PbMessage
	err = proto.Unmarshal(msgBytes, &pbMessage)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return nil
	}

	modelMsg := ConvertPbMessage(&pbMessage)
	if c.victim {
		fmt.Println("VICTIM !!!!")
		var connected bool
		for _, n := range c.group {
			if n == modelMsg.Source {
				connected = true
				break
			}
		}
		if !connected {
			c.buffer <- modelMsg
			return nil
		}
	}

	return &modelMsg
}

func (c *libp2pPubSub) Disconnect() {
	c.subscription.Cancel()
	fmt.Println("DISCONNECT")
}

func (c *libp2pPubSub) Reconnect(topic string) {
	var err error
	if topic != "" {
		c.topic = topic
	}

	c.subscription, err = c.pubsub.Subscribe(c.topic)
	if err != nil {
		panic(err)
	}
	fmt.Println("RECONNECT to topic ", c.topic)
}

// Cancel unsubscribes a node from pubsub
func (c *libp2pPubSub) Cancel(cancelTime int, reconnectTime int) {
	go func() {
		time.Sleep(time.Duration(cancelTime) * time.Millisecond)
		fmt.Println("	CANCELING	")
		c.subscription.Cancel()
		time.Sleep(time.Duration(reconnectTime) * time.Millisecond)
		fmt.Println("	RESUBBING	")
		c.subscription, _ = c.pubsub.Subscribe(c.topic)
	}()
}

// createPeer creates a peer on localhost and configures it to use libp2p.
func (c *libp2pPubSub) CreatePeer(nodeId int, port int) *core.Host {
	// Creating a node
	h, err := createHost(port)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Node %v is %s\n", nodeId, getLocalhostAddress(h))

	return &h
}

// initializePubSub creates a PubSub for the peer and also subscribes to a topic
func (c *libp2pPubSub) InitializePubSub(h core.Host) {
	var err error
	// Creating pubsub
	// every peer has its own PubSub
	c.pubsub, err = applyPubSub(h)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}

	// Creating a subscription and subscribing to the topic
	c.subscription, err = c.pubsub.Subscribe(c.topic)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}

}

// InitializeVictim initializes buffer for keeping messages when a node is attacked by adversary.
func (c *libp2pPubSub) InitializeVictim(makeBuffer bool) {
	// victim is always false in initialization
	c.victim = false
	if makeBuffer {
		c.buffer = make(chan model.Message, BufferLen)
	} else {
		c.buffer = make(chan model.Message, 0)
	}
}

// AttackVictim adds a node to the set of indefinite-delayed nodes.
func (c *libp2pPubSub) AttackVictim() {
	c.victim = true
	c.makeVictimNotGossip()
}

// ReleaseVictim removes the node from set of delayed nodes.
func (c *libp2pPubSub) ReleaseVictim() {
	c.Disconnect()
	c.victim = false
	c.Reconnect("")
	err := c.pubsub.UnregisterTopicValidator(c.topic)
	if err != nil {
		panic(err)
	}

	fmt.Println("REJOINING FROM DELAYED SET")
}

// JoinGroup adds nodes within the same groups to the node's group variable.
func (c *libp2pPubSub) JoinGroup(group []int) {
	c.group = group
}

// makeVictimNotGossip prevents victim from participating in gossip protocol
func (c *libp2pPubSub) makeVictimNotGossip() {
	// Registering a message validator function. This function will process every received message by pubsub and based
	// on return value will forward it to other nodes. Returning false will prevent the peer from forwarding the message
	err := c.pubsub.RegisterTopicValidator(c.topic, func(ctx context.Context, pid peer.ID, msg *pubsub.Message) bool {
		// Process message in a go routine to prevent blocking this function
		go func(data []byte) {
			msgBytes := data
			var pbMessage protobuf.PbMessage
			err := proto.Unmarshal(msgBytes, &pbMessage)
			if err != nil {
				panic(err)
			}

			modelMsg := ConvertPbMessage(&pbMessage)
			c.buffer <- modelMsg
		}(msg.Data)

		//if the return value is true, the message will hit other peers
		//if the return value is false, this peer prevented message broadcasting
		//note that this topic validator will be called also for messages sent by self
		return false
	})

	if err != nil {
		panic(err)
	}
}

// createHost creates a host with some default options and a signing identity
func createHost(port int) (core.Host, error) {
	// Producing private key
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	sk := (*crypto.Secp256k1PrivateKey)(prvKey)

	// Starting a peer with default configs
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port)),
		libp2p.Identity(sk),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
	}

	h, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	return h, nil
}

// createHostQUIC creates a host with QUIC as transport layer implementation
func createHostQUIC(port int) (core.Host, error) {
	// Producing private key
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, rand.Reader)

	quicTransport, err := quic.NewTransport(priv)
	if err != nil {
		return nil, err
	}

	// Starting a peer with QUIC transport
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", port)),
		libp2p.Transport(quicTransport),
		libp2p.Identity(priv),
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
	}

	h, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	return h, nil
}

// createHostWebSocket creates a host with WebSocket as transport layer implementation
func createHostWebSocket(port int) (core.Host, error) {

	// Starting a peer with QUIC transport
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/udp/%d/quic", port)),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d/ws", port)),
		libp2p.Transport(ws.New),
		libp2p.DefaultMuxers,
		libp2p.DefaultSecurity,
	}

	h, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		return nil, err
	}

	return h, nil
}

// getLocalhostAddress is used for getting address of hosts
func getLocalhostAddress(h core.Host) string {
	for _, addr := range h.Addrs() {
		if strings.Contains(addr.String(), "127.0.0.1") {
			return addr.String() + "/p2p/" + h.ID().Pretty()
		}
	}

	return ""
}

// applyPubSub creates a new GossipSub with message signing
func applyPubSub(h core.Host) (*pubsub.PubSub, error) {
	optsPS := []pubsub.Option{
		pubsub.WithMessageSigning(true),
	}

	return pubsub.NewGossipSub(context.Background(), h, optsPS...)
}

// connectHostToPeer is used for connecting a host to another peer
func connectHostToPeer(h core.Host, connectToAddress string) {
	// Creating multi address
	multiAddr, err := multiaddr.NewMultiaddr(connectToAddress)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}

	pInfo, err := peer.AddrInfoFromP2pAddr(multiAddr)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}

	err = h.Connect(context.Background(), *pInfo)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
	}
}

// ConvertModelMessage is for converting message defined in model to message used by protobuf
func ConvertModelMessage(msg model.Message) (message *protobuf.PbMessage) {
	source := int64(msg.Source)
	step := int64(msg.Step)

	msgType := protobuf.MsgType(int(msg.MsgType))

	history := make([]*protobuf.PbMessage, 0)

	for _, hist := range msg.History {
		history = append(history, ConvertModelMessage(hist))
	}

	message = &protobuf.PbMessage{
		Source:  &source,
		Step:    &step,
		MsgType: &msgType,
		History: history,
	}
	return
}

// ConvertPbMessage is for converting protobuf message to message used in model
func ConvertPbMessage(msg *protobuf.PbMessage) (message model.Message) {
	history := make([]model.Message, 0)

	for _, hist := range msg.History {
		history = append(history, ConvertPbMessage(hist))
	}

	message = model.Message{
		Source:  int(msg.GetSource()),
		Step:    int(msg.GetStep()),
		MsgType: model.MsgType(int(msg.GetMsgType())),
		History: history,
	}
	return
}
