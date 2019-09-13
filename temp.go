package main

import (
	"fmt"
	"os"
	"sync"
)

var wg sync.WaitGroup

type ss struct {
	value  int
	shadow int
}

type pp struct {
	ss
	mod    int
	shadow int
	sig    []byte
}

func (ss *ss) play(k int, c chan int) {
	go func() {
		c <- ss.value
		c <- k
	}()
}

func (ppp *pp) play(k int, c chan int) {
	go func() {
		c <- ppp.value
		c <- k
	}()
}

func main() {
	fmt.Println(os.Args[1])
	ppp := pp{}
	ppp.value = 23

	zz := pp{}
	zz.sig = []byte("siktir")
	zz.sig = nil
	if nil == zz.sig {
		fmt.Println(zz.sig)
	}

}
