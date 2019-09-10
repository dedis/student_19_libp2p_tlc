package main

import (
	"io/ioutil"
	"log"
	"strings"

	"go.dedis.ch/onet/v3"
	"go.dedis.ch/protobuf"
)

func main() {
	buf, err := ioutil.ReadFile("config")
	if err != nil {
		log.Fatal("could not read config", err)
	}

	var r = new(onet.Roster)
	err = protobuf.Decode(buf, r)
	if err != nil {
		log.Fatal("could not read decode", err)
	}

	log.Println("** This is the protocol binary running.")
	peers := getPeers(r)
	log.Println("** all peers:", strings.Join(peers, " "))
}

func getPeers(r *onet.Roster) (out []string) {
	for _, x := range r.List {
		s := string(x.Address)
		s = s[6:]
		out = append(out, s)
	}
	return
}
