package main

import (
	"context"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/ethclient"
)

// WalletClient wraps all the basics for a wallet
type WalletClient struct {
	client *ethclient.Client  // RPC client to the blockchain nodes
	keys   *keystore.KeyStore // The keys managed by the wallet client
	ctx    context.Context    // Context object serving the network calling
	cancel context.CancelFunc // The cancel function for the context object, for clear cleaning
	exit   chan struct{}      // channel to signal the wallet exit is safe, for cases of parreling processing
}

// NewWalletClient returns a wallet client with ctx and exit channel
func NewWalletClient() *WalletClient {
	re := WalletClient{}
	re.ctx, re.cancel = context.WithTimeout(context.Background(), 30*time.Second)
	re.exit = make(chan struct{})
	keys := tmpKeyStore()
	re.keys = keys

	return &re
}

// Exit signalling that exit is allowed
func (w *WalletClient) Exit() {
	if w.client != nil {
		w.client.Close()
	}
	w.cancel()
	w.exit <- struct{}{} // close channel to signal exit
}

// Wait wait for the signal of exit
func (w *WalletClient) Wait() {
	if w != nil {
		<-w.exit
	}
	panic("error exit chan has not been initialized!")
}

// dev... the followings are for convenience, methods of blockchain client, exposed by wallet client
// NonceAt call the nonceat of blockchain client
func (w *WalletClient) nonceAt(acc accounts.Account) (uint64, error) {
	return w.client.PendingNonceAt(w.ctx, acc.Address)
}

func (w *WalletClient) gasLimitRecommended() (*big.Int, error) {
	return big.NewInt(300000), nil
}

func (w *WalletClient) gasPriceRecommended() (*big.Int, error) {
	return w.client.SuggestGasPrice(w.ctx)
}

func (w *WalletClient) chainID() (*big.Int, error) {
	return w.client.NetworkID(w.ctx)
}
