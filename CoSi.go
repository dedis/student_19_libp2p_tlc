package main

import (
	"crypto/sha256"
	"fmt"
	"go.dedis.ch/kyber/v3"
	"go.dedis.ch/kyber/v3/pairing"
	"go.dedis.ch/kyber/v3/sign"
	"go.dedis.ch/kyber/v3/sign/bdn"
)

func main() {
	suite := pairing.NewSuiteBn256()
	signatures := make([][]byte, 0)
	publicKeys := make([]kyber.Point, 0)
	var err error

	// Hashing
	h := sha256.New()
	msg := "HELLO"
	h.Write([]byte(msg))
	msgHash := h.Sum(nil)

	for i := 0; i < 5; i++ {

		// private key
		var a kyber.Scalar
		a = suite.Scalar().Pick(suite.RandomStream())

		// list of public keys
		publicKeys = append(publicKeys, suite.Point().Mul(a, nil))

		//signatures[i], err = bdn.Sign(suite, a, msgHash)
		if i == 3 {
			continue
		}
		// everyone signs the message
		sig, _ := bdn.Sign(suite, a, msgHash)

		// keeping list of all signatures
		signatures = append(signatures, sig)
		if err != nil {
			panic(err)
		}
	}
	//fmt.Println(len(signatures), signatures)
	//fmt.Println(len(publicKeys), publicKeys)

	/*
		temp := signatures[0]
		signatures[0] = signatures[1]
		signatures[1] = temp
	*/

	// get a mask of pubkeys
	myMask, _ := sign.NewMask(suite, publicKeys, nil)

	fmt.Println(myMask.Publics())
	fmt.Println(myMask.Participants())

	// We have to add signatures as participants
	for i := 0; i < 5; i++ {
		if i == 3 {
			continue
		}
		// adding people who might be needed for signature verifications
		myMask.SetBit(i, true)
		// getting number of participants
		fmt.Println(myMask.CountEnabled())
	}

	fmt.Println(len(myMask.Participants()))

	aggregatePubKey, _ := bdn.AggregatePublicKeys(suite, myMask)
	aggregateSig, err := bdn.AggregateSignatures(suite, signatures, myMask)
	if err != nil {
		panic(err)
	}

	fmt.Println("IN ", myMask.IndexOfNthEnabled(5))
	binAggregateSig, _ := aggregateSig.MarshalBinary()
	err = bdn.Verify(suite, aggregatePubKey, msgHash, binAggregateSig)
	if err != nil {
		panic(err)
	}

}
