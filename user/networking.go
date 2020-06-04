package main

import (
	"context"
	"fmt"

	host "github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
)

// newStream opens a new stream to the master node, and writes a p2p/protocol
func newSpeedStream(ha host.Host, target string) network.Stream {
	peerID, targetAddr := getPeerIDAndMultiaddr(target)
	ha.Peerstore().AddAddr(peerID, targetAddr, pstore.PermanentAddrTTL)

	// make a new stream from this user node to the region node
	s, err := ha.NewStream(context.Background(), peerID, UserSpeedProtocolID)
	checkErrorFatal(err)

	return s
}

func newAlertStream(ha host.Host, target string) network.Stream {
	peerID, targetAddr := getPeerIDAndMultiaddr(target)
	ha.Peerstore().AddAddr(peerID, targetAddr, pstore.PermanentAddrTTL)

	// make a new stream from this user node to the region node
	s, err := ha.NewStream(context.Background(), peerID, UserAlertProtocolID)
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
