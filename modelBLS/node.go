package modelBLS

import "go.dedis.ch/kyber/v3"

// Node is the struct used for keeping everything related to a node in TLC.
type Node struct {
	Id           int                    // Id of the node
	TimeStep     int                    // Node's local time step
	ThresholdAck int                    // Threshold on number of messages
	ThresholdWit int                    // Threshold on number of witnessed messages
	Acks         int                    // Number of acknowledges
	Wits         int                    // Number of witnesses
	Comm         CommunicationInterface // interface for communicating with other nodes
	CurrentMsg   MessageWithSig         // Message which the node is waiting for acks
	History      []MessageWithSig       // History of received messages by a node
	PublicKeys   []kyber.Point          // Public keys of all nodes
	privateKey   kyber.Scalar           // Private key of the node
}

// CommunicationInterface is a interface used for communicating with transport layer.
type CommunicationInterface interface {
	Send(MessageWithSig, int)     // Send a message to a specific node
	Broadcast(sig MessageWithSig) // Broadcast messages to other nodes
	Receive() *MessageWithSig     // Blocking receive
	Disconnect()                  // Disconnect node
	Reconnect(string)             // Reconnect node
}
