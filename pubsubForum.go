package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/libp2p/go-libp2p"
	core "github.com/libp2p/go-libp2p-core"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/multiformats/go-multiaddr"
	mrand "math/rand"
)

func main() {
	hosts := make([]core.Host, 0)
	startingPort := 10000
	var err error

	//Step 1. Create 5 nodes
	for i := 0; i < 5; i++ {
		h, err := createHost(startingPort + i)
		if err != nil {
			fmt.Printf("Error encountered: %v\n", err)
			return
		}

		hosts = append(hosts, h)
		fmt.Printf("Node %v is %s with ID %s\n", i, getLocalhostAddress(h), h.ID())
	}

	defer func() {
		for _, h := range hosts {
			_ = h.Close()
		}
	}()

	//Step 2. Create pubsubs
	// TODO: `mahdi` We are creating a slice of pubsubs!
	pubsubs := make([]*pubsub.PubSub, len(hosts))
	for i, h := range hosts {
		// TODO: `mahdi` Every host is starting his own pubsub?
		pubsubs[i], err = applyPubSub(h)
		if err != nil {
			fmt.Printf("Error encountered: %v\n", err)
			return
		}
	}

	//Step 3. Register to the topic and add topic validators
	//var subscrC *pubsub.Subscription
	topic := "test"
	for i := 0; i < len(pubsubs); i++ {
		// TODO: `mahdi` Just catching a subscription (?) I guess it means that using this you can get notifications
		// Bt
		var subscr *pubsub.Subscription
		subscr, err = pubsubs[i].Subscribe(topic)
		if err != nil {
			fmt.Printf("Error encountered: %v\n", err)
			return
		}

		//if i == 4 {
		//	subscrC = subscr
		//}

		//just a dummy func to consume messages received by the newly created topic
		go func() {
			for {
				//here you will actually have the message received after passing all validators
				//not required since we put validators on each topic and the message has already been processed there
				msg, err := subscr.Next(context.Background())
				if err != nil {
					return
				}
				fmt.Printf("I was waiting for the message %s\n", msg.Data)
			}
		}()
		//if i == 2 || i == 3 {
		//	subscr.Cancel()
		//}
		// TODO: `mahdi` current host?
		crtHost := hosts[i]
		// TODO: `mahdi` what is this function supposed to do?
		err = pubsubs[i].RegisterTopicValidator(topic, func(ctx context.Context, pid peer.ID, msg *pubsub.Message) bool {
			//do the message validation
			//example: deserialize msg.Data, do checks on the message, etc.

			//processing part should be done on a go routine as the validator func should return immediately
			go func(data []byte, p peer.ID, h core.Host) {
				fmt.Printf("%s: Message: '%s' was received from %s\n", crtHost.ID(), msg.Data, pid.Pretty())
			}(msg.Data, pid, crtHost)

			//if the return value is true, the message will hit other peers
			//if the return value is false, this peer prevented message broadcasting
			//note that this topic validator will be called also for messages sent by self
			return true
		})

	}

	//Step 4. Connect the nodes as following:
	//
	// node0 --------- node1
	//   |               |
	//   +------------ node2
	//   |               |
	//   |             node3
	//   |               |
	//   +------------ node4
	// Manual peer discovery
	fmt.Println("GG", getLocalhostAddress(hosts[0]))
	ma, _ := multiaddr.NewMultiaddr("/ip4/127.0.0.1/tcp/10001")
	fmt.Println("GG", ma)

	connectHostToPeer(hosts[1], getLocalhostAddress(hosts[0]))
	connectHostToPeer(hosts[2], "/ip4/127.0.0.1/tcp/10001")
	connectHostToPeer(hosts[3], "/ip4/127.0.0.1/tcp/10002")
	connectHostToPeer(hosts[4], "/ip4/127.0.0.1/tcp/10003")

	//Step 5. Wait so that subscriptions on topic will be done and all peers will "know" of all other peers
	time.Sleep(time.Second * 2)

	fmt.Println("Broadcasting a message on node 0...")

	err = pubsubs[0].Publish(topic, []byte("a message from friend"))
	if err != nil {
		fmt.Printf("Error encountered: %v\n", err)
		return
	}

	time.Sleep(2 * time.Second)
	//subscrC.Cancel()
	err = pubsubs[0].Publish(topic, []byte("Cancel"))

	time.Sleep(10 * time.Second)

}

func createHost(port int) (core.Host, error) {
	r := mrand.New(mrand.NewSource(int64(port)))
	prvKey, _ := ecdsa.GenerateKey(btcec.S256(), r)
	sk := (*crypto.Secp256k1PrivateKey)(prvKey)

	id, _ := peer.IDFromPrivateKey(sk)
	fmt.Println("ZZ", id.Pretty())

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

func applyPubSub(h core.Host) (*pubsub.PubSub, error) {
	optsPS := []pubsub.Option{
		pubsub.WithMessageSigning(true),
	}

	return pubsub.NewGossipSub(context.Background(), h, optsPS...)
}

func getLocalhostAddress(h core.Host) string {
	for _, addr := range h.Addrs() {
		if strings.Contains(addr.String(), "127.0.0.1") {
			return addr.String() + "/p2p/" + h.ID().Pretty()
		}
	}

	return ""
}

func connectHostToPeer(h core.Host, connectToAddress string) {
	multiAddr, err := multiaddr.NewMultiaddr(connectToAddress)
	if err != nil {
		fmt.Printf("Error encountered1: %v\n", err)
		return
	}

	pInfo, err := peer.AddrInfoFromP2pAddr(multiAddr)
	if err != nil {
		fmt.Printf("Error encountered2: %v\n", err)
		return
	}

	err = h.Connect(context.Background(), *pInfo)
	if err != nil {
		fmt.Printf("Error encountered3: %v\n", err)
	}
}
