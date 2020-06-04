package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	host "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

// newStream opens a new stream to the master node, and writes a p2p/protocol
func newStream(ha host.Host, target string) network.Stream {
	peerID, targetAddr := getPeerIDAndMultiaddr(target)
	ha.Peerstore().AddAddr(peerID, targetAddr, pstore.PermanentAddrTTL)

	// make a new stream from this region node to the master node
	s, err := ha.NewStream(context.Background(), peerID, NodesProtocolID)
	checkErrorFatal(err)
	return s
}

func getPeerIDAndMultiaddr(target string) (peer.ID, ma.Multiaddr) {
	ipfsAddr, err := ma.NewMultiaddr(target)
	checkErrorFatal(err)

	pid, err := ipfsAddr.ValueForProtocol(ma.P_IPFS)
	checkErrorFatal(err)

	peerID, err := peer.IDB58Decode(pid)
	checkErrorFatal(err)

	targetPeerAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerID)))
	return peerID, ipfsAddr.Decapsulate(targetPeerAddr)
}

func handleUserSpeedStream(userStream net.Stream) {
	log.Println("Got a new user stream!")

	// Create a buffer stream for non blocking read and write.
	userRW := bufio.NewReadWriter(bufio.NewReader(userStream), bufio.NewWriter(userStream))

	// read user id
	userID := readUserID(userRW)

	// create a lock for acessing the users location
	userLocationMutexes[userID] = &sync.Mutex{}

	// Create goroutines that continuously pass speed and location data between region node(this) and user
	go readUserSpeedMessage(userRW, userID)
}

func handleUserAlertStream(userStream net.Stream) {
	// Create a buffer stream for non blocking read and write.
	userRW := bufio.NewReadWriter(bufio.NewReader(userStream), bufio.NewWriter(userStream))

	userID := readUserID(userRW)

	// send all alerts in the system
	bytes, err := json.Marshal(allActiveAlerts)
	checkErrorFatal(err)
	userRW.WriteString(fmt.Sprintf("%s\n", string(bytes)))
	userRW.Flush()

	// Create goroutines that continuously pass alert data between region node(this) and user
	go readUserAlertMessage(userRW, userID)
	go sendUserAlertMessage(userRW, userID)
}
