package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"

	"go.dedis.ch/onet/v3"
	"go.dedis.ch/protobuf"
	//mrand "math/rand"
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

	var id int
	id, _ = strconv.Atoi(os.Args[1])
	log.Println("** This is the protocol binary running.", id)
	peers := getPeers(r, id)
	log.Println("** all peers:", strings.Join(peers, " "))
}

func getPeers(r *onet.Roster, id int) (out []string) {
	fmt.Println(r.List[id].Address)
	out = append(out, string(r.List[id].Address))
	//for _, x := range r.List {
	//	s := string(x.Address)
	//	s = s[6:]
	//	address := strings.Split(s,":")
	//	r := mrand.New(mrand.NewSource(int64(i)))
	//out = append(out, address[1])
	//}
	return
}
