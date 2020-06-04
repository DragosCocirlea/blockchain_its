package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	mrand "math/rand"
	"strings"

	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-core/crypto"
	host "github.com/libp2p/go-libp2p-core/host"
	ma "github.com/multiformats/go-multiaddr"
)

// makeRegionHost creates a LibP2P host with a random peer ID listening on the given multiaddress
func makeRegionHost(listenPort int, randseed int64) (host.Host, error) {

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

	// Build host multiaddress
	// e.g.: hostAddr = /p2p/QmaCggGLJD1jXVvTUJhR18g2fUyS5UU6hGfTgsvqEuGdGr
	multiAddr := fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty())
	hostAddr, _ := ma.NewMultiaddr(multiAddr)

	// select the ipv4 address
	// e.g.: addr = /ip4/127.0.0.1/tcp/10000
	addrs := basicHost.Addrs()
	var addr ma.Multiaddr
	for _, i := range addrs {
		if strings.HasPrefix(i.String(), "/ip4") {
			addr = i
			break
		}
	}

	// Now we can build a full multiaddress to reach this host by encapsulating both addresses
	// e.g.: fullAddr = /ip4/127.0.0.1/tcp/10000/p2p/QmcbsMC9x97PXzZGhJ723dsuKKX6HH2fMpocDoMKAwmGXk
	fullAddr := addr.Encapsulate(hostAddr)

	fmt.Printf("I am region node %s\n", basicHost.ID().Pretty())
	fmt.Printf("\nNow run this on a different terminal in the user directory after the blockchain parsing has finished:\ngo run *.go -port %d -peer %s\n\n", listenPort+1, fullAddr)

	return basicHost, nil
}
