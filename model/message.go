package model

type MsgType int

const (
	Raw = iota
	Ack
	Wit
)

type Message struct {
	Source  int       // NodeID of message's source
	Step    int       // Time step of message
	MsgType MsgType   // Type of message
	History []Message // History of messages. Sent to ensure causality, But in an inefficient way.
}
