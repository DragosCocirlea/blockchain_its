package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	mrand "math/rand"

	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	host "github.com/libp2p/go-libp2p-core/host"
)

// makeUserHost creates a LibP2P host with a random peer ID listening on the given multiaddress
func makeUserHost(listenPort int, target string, randseed int64) (host.Host, error) {

	// seed == 0, real cryptographic randomness
	// else, deterministic randomness source to make generated keys stay the same across multiple runs
	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	// Generate a key pair for this host. We will use it to obtain a valid host ID.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	// Generate the libp2p host
	basicHost, err := libp2p.New(
		context.Background(),
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
	)
	if err != nil {
		return nil, err
	}

	fmt.Printf("I am user node %s\n", basicHost.ID().Pretty())
	fmt.Printf("\nNow run this on a different terminal in the user directory in order to connect to the same region node:\ngo run *.go -port %d -peer %s\n\n", listenPort+1, target)

	return basicHost, nil
}
