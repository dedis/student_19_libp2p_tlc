package libp2p_pubsub

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/peer"
	"github.com/multiformats/go-multiaddr"
	"strings"

	"github.com/dedis/student_19_libp2p_tlc/model"

	"github.com/btcsuite/btcd/btcec"
	"github.com/libp2p/go-libp2p"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/crypto"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

type libp2pPubSub struct {
	pubsub       *pubsub.PubSub       // PubSub of each individual node
	subscription *pubsub.Subscription // Subscription of individual node
	topic        string               // PubSub topic
}

func (c *libp2pPubSub) Broadcast(msg *model.Message) {
	// Broadcasting to a topic in PubSub
	//err := c.pubsub.Publish(c.topic, []byte("a message from friend"))
	fmt.Printf("sending this message as struct: %v\n", msg)
	//ms := convertModelMessage(msg)
	//fmt.Printf("sending this message as protobuf: %v,%v\n", ms,ms.History)
	msgBytes, err := proto.Marshal(convertModelMessage(msg))
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}
	//fmt.Printf("sending this message as bytes: %s\n", msgBytes)
	err = c.pubsub.Publish(c.topic, msgBytes)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}
}

func (c *libp2pPubSub) Send(msg *model.Message, n *model.Node) {
	// TODO: Implement direct messages by using streams(?) or something
}

func (c *libp2pPubSub) Receive() *model.Message {
	// Blocking function for consuming newly received messages
	// We can access message here, but we need subscription. Then we can start processing the received message
	msg, _ := c.subscription.Next(context.Background())
	//fmt.Printf("I was waiting for the message : %s\n", msg.Data)
	msgBytes := msg.Data
	var pbMessage PbMessage
	err := proto.Unmarshal(msgBytes, &pbMessage)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return nil
	}
	modelMsg := convertPbMessage(&pbMessage)
	if len(modelMsg.History) > 8 {
		mmm := *(modelMsg.History[len(modelMsg.History)-1])
		fmt.Println(mmm)
	}
	return modelMsg
}

// createPeer creates a peer on localhost and configures it to use libp2p.
func (c *libp2pPubSub) createPeer(nodeId int, port int) *core.Host {
	//c.hosts = make([]host.Host, 0)
	//var err error

	// Creating nodes
	h, err := createHost(port)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Node %v is %s\n", nodeId, getLocalhostAddress(h))

	// Returning pointer to the created libp2p host
	return &h
	/*
		// TODO : it is wrong. you are closing it immediately!
		defer func() {
			for _, h := range c.hosts {
				_ = h.Close()
			}
		}()
	*/

	// TODO: I have to connect peers to each other.
}

// initializePubSub creates a PubSub for the peer and also subscribes to a topic
func (c *libp2pPubSub) initializePubSub(h core.Host) {
	var err error
	// Creating pubsub
	// every peer has its own PubSub
	c.pubsub, err = applyPubSub(h)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}

	// Registering to the topic
	c.topic = "TLC"
	// Creating a subscription and subscribing to the topic
	c.subscription, err = c.pubsub.Subscribe(c.topic)
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}

}

// createHost creates a host with some defaults options and a signing identity
func createHost(port int) (core.Host, error) {
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), rand.Reader)
	sk := (*crypto.Secp256k1PrivateKey)(prvKey)

	// starting a peer
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

func connectHostToPeer(h core.Host, connectToAddress string) {
	multiAddr, err := multiaddr.NewMultiaddr(connectToAddress)
	if err != nil {
		fmt.Printf("Error encountered: %v\n", err)
		return
	}

	pInfo, err := peer.AddrInfoFromP2pAddr(multiAddr)
	if err != nil {
		fmt.Printf("Error encountered: %v\n", err)
		return
	}

	err = h.Connect(context.Background(), *pInfo)
	if err != nil {
		fmt.Printf("Error encountered: %v\n", err)
	}
}

// convertModelMessage is for converting message defined in model to message used by protobuf
func convertModelMessage(msg *model.Message) (message *PbMessage) {
	source := int64(msg.Source)
	step := int64(msg.Step)

	msgType := MsgType(int(msg.MsgType))

	history := make([]*PbMessage, 0)

	for _, hist := range msg.History {
		history = append(history, convertModelMessage(hist))
	}

	message = &PbMessage{
		Source:  &source,
		Step:    &step,
		MsgType: &msgType,
		History: history,
	}
	return
}

// convertPbMessage is for converting protobuf message to message used in model
func convertPbMessage(msg *PbMessage) (message *model.Message) {
	history := make([]*model.Message, 0)

	for _, hist := range msg.History {
		history = append(history, convertPbMessage(hist))
	}

	message = &model.Message{
		Source:  int(msg.GetSource()),
		Step:    int(msg.GetStep()),
		MsgType: model.MsgType(int(msg.GetMsgType())),
		History: history,
	}
	return
}
