package transport

import (
	"context"
	"crypto/ecdsa"
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/dedis/student_19_libp2p_tlc/model"

	"github.com/btcsuite/btcd/btcec"
	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p-core/crypto"
	host "github.com/libp2p/go-libp2p-host"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
)

type libp2pPubSub struct {
	pubsub       *pubsub.PubSub       // PubSub of each individual node
	subscription *pubsub.Subscription // Subscription of individual node
	topic        string               // PubSub topic
}

func (c *libp2pPubSub) Broadcast(msg *model.Message) {
	// Broadcasting to a topic in PubSub
	// TODO: I have to use Protobuf here for sending messages
	err := c.pubsub.Publish(c.topic, []byte("a message from friend"))
	if err != nil {
		fmt.Printf("Error : %v\n", err)
		return
	}
}

func (c *libp2pPubSub) Send(msg *model.Message, n *model.Node) {

}

func (c *libp2pPubSub) Receive() *model.Message {
	// Blocking function for consuming newly received messages
	// We can access message here, but we need subscription. Then we can start processing the received message
	msg, _ := c.subscription.Next(context.Background())
	fmt.Printf("I was waiting for the message : %s\n", msg.Data)

}

// createHost creates a peer on localhost and configures it to use libp2p.
func (c *libp2pPubSub) createHost(n int, nodeId int, port int) *host.Host {
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
func (c *libp2pPubSub) initializePubSub(h host.Host) {
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

// createHost creates a peer with some defaults options and a signing identity
func createHost(port int) (host.Host, error) {
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

func getLocalhostAddress(h host.Host) string {
	for _, addr := range h.Addrs() {
		if strings.Contains(addr.String(), "127.0.0.1") {
			return addr.String() + "/p2p/" + h.ID().Pretty()
		}
	}

	return ""
}

// applyPubSub creates a new GossipSub with message signing
func applyPubSub(h host.Host) (*pubsub.PubSub, error) {
	optsPS := []pubsub.Option{
		pubsub.WithMessageSigning(true),
	}

	return pubsub.NewGossipSub(context.Background(), h, optsPS...)
}
