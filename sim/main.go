package main

import (
	"fmt"
	"github.com/BurntSushi/toml"
	"go.dedis.ch/onet/v3"
	"go.dedis.ch/onet/v3/log"
	"go.dedis.ch/onet/v3/simul"
	"go.dedis.ch/protobuf"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
)

func init() {
	onet.SimulationRegister("TlcTest", NewSimulation)
}

type simulation struct {
	onet.SimulationBFTree
	Ratio float64
	Other string
}

// NewSimulation returns the new simulation, where all fields are
// initialised using the config-file
func NewSimulation(config string) (onet.Simulation, error) {
	es := &simulation{}
	_, err := toml.Decode(config, es)
	if err != nil {
		return nil, err
	}
	return es, nil
}

func (e *simulation) Setup(dir string, hosts []string) (
	*onet.SimulationConfig, error) {
	log.LLvl1("** Executing Setup")
	sc := &onet.SimulationConfig{}
	e.CreateRoster(sc, hosts, 2000)
	err := e.CreateTree(sc)
	if err != nil {
		return nil, err
	}

	log.LLvl1("** Executing Setup, dumping Roster")
	buf, err := protobuf.Encode(sc.Roster)
	if err != nil {
		log.Fatal(err)
	}
	err = ioutil.WriteFile(path.Join(dir, "config"), buf, 0644)
	if err != nil {
		log.Fatal(err)
	}

	log.LLvl1("** Executing Setup, cp into setup dir")
	cmd := exec.Command("cp", "protocol/protocol", dir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	log.LLvl1("** Executing Setup, cp ok")

	return sc, nil
}

func (e *simulation) Run(config *onet.SimulationConfig) error {
	return nil
}

func (e *simulation) Node(config *onet.SimulationConfig) error {
	log.LLvl1("** Executing Node")
	fmt.Println(config.Server.ServerIdentity.Address)
	a, _ := config.Roster.Search(config.Server.ServerIdentity.ID)
	fmt.Println(a)
	cmd := exec.Command("./protocol", strconv.Itoa(a))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}

func main() {
	simul.Start()
}
