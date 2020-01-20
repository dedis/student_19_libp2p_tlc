# TLC Implementations
This repository contains different transport layer implementations for [TLC](https://arxiv.org/abs/1907.07010) protocol.
 Currently, there are 2 main implementations:
* Libp2p implementation
* Mail-based implementation

## Libp2p Implementation
In this impementation, libp2p is used as the communication layer of the TLC protocol. Nodes broadcast and receive messages 
using libp2p implementation of PubSub which is named GossipSub. All nodes subscribe to a topic and then start publishing 
messages to that topic. Subscribed nodes will receive published messages on the topic.

3 different transport layer implementations have been used as the tranport module of libp2p:
* TCP
* QUIC
* Websocket

Simulations are mainly based on the TCP implementation but all functions are provided for simulating on top of the other 
protocols.

## Mail-based Implementation
In this implementation, messages will be sent using mail protocols. SMTP is used for ongoing mails and IMAP is used for 
receiving incoming mails. Mail server implementation is not present in this repository and anyone interested in using 
mail-based implementation needs to setup a working mail server. You may use iRedMail to start a mail server.

## Repository structure
    .
    ├── Report                       A short report, illustrating results obtained in simulations
    ├── model                        TLC protocol implementation with abstract transport layer
    ├── sim                          Onet simulation driver                        
    ├── transport                    Contains differnent transport layer implementations
        ├── channel                  Go channel based transport
        ├── libp2p_pubsub            Libp2p-based implementataion using pubsub
        └── mail                     Mail-based implementation
        

Contact me ([Mahdi Bakhshi](https://github.com/MBakhshi96)) if you have any problems.
