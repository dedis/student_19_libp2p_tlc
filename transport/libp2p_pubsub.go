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
	hosts []host.Host
}

func (c *libp2pPubSub) Broadcast(msg *model.Message) {

}

func (c *libp2pPubSub) Send(msg *model.Message, n *model.Node) {

}

func (c *libp2pPubSub) Receive() *model.Message {

}

// createHosts creates some peers on localhost and configures them to use libp2p.
func (c *libp2pPubSub) createHosts(n int) {
	c.hosts = make([]host.Host, 0)
	startingPort := 10000
	//var err error

	// Creating nodes
	for i := 0; i < n; i++ {
		h, err := createHost(startingPort + i)
		if err != nil {
			fmt.Printf("Error : %v\n", err)
			return
		}

		c.hosts = append(c.hosts, h)
		fmt.Printf("Node %v is %s\n", i, getLocalhostAddress(h))
	}

	defer func() {
		for _, h := range c.hosts {
			_ = h.Close()
		}
	}()
}

// initializePubSubs creates a PubSub for every peer and also subscribes to a topic
func (c *libp2pPubSub) initializePubSubs() {
	var err error
	// Creating pubsubs
	pubsubs := make([]*pubsub.PubSub, len(c.hosts))
	for i, h := range c.hosts {
		// every peer has its own PubSub
		pubsubs[i], err = applyPubSub(h)
		if err != nil {
			fmt.Printf("Error : %v\n", err)
			return
		}
	}

	// Registering to the topic
	topic := "TLC"

	for i := 0; i < len(pubsubs); i++ {
		// Creating a subscription and subscribing to the topic
		var subscription *pubsub.Subscription
		subscription, err = pubsubs[i].Subscribe(topic)
		if err != nil {
			fmt.Printf("Error : %v\n", err)
			return
		}

		// TODO: this function must not be here!
		// Blocking function for consuming newly received messages
		go func() {
			for {
				// We can access message here
				msg, _ := subscription.Next(context.Background())
				fmt.Printf("I was waiting for the message : %s\n", msg.Data)
			}
		}()

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
