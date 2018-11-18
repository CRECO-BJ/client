package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli"
)

var account = common.HexToAddress("0x8f2d2b848ede60d9480631fe6a365cbc8e304c14")

var (
	app   = cli.NewApp()
	flags = []cli.Flag{
		cli.StringFlag{
			Name:  "nodeaddr",
			Usage: "the address of the eth node/server",
			Value: "http://139.9.2.54:4567",
		},
	}
	commands = []cli.Command{
		{
			Name:    "new",
			Aliases: []string{"a"},
			Usage:   "create a new account and attach the account to this wallet",
			Action:  newAccount,
		},
		{
			Name:    "import",
			Aliases: []string{"i"},
			Usage:   "import an account from eth-json private key-store file",
			Action:  importKeystore,
		},
		{
			Name:    "export",
			Aliases: []string{"e"},
			Usage:   "export an account's private key-store file",
			Action:  exportKeystore,
		},
		{
			Name:    "balance",
			Aliases: []string{"b"},
			Usage:   "check the balance of an account",
			Action:  checkBalance,
		},
		{
			Name:    "transfer",
			Aliases: []string{"t"},
			Usage:   "transfer a few coin to the target address",
			Action:  transfer,
		},
	}
	wallet *WalletClient
)

// WalletClient wraps all the basics for a wallet
type WalletClient struct {
	client *ethclient.Client
	ctx    context.Context
	cancel context.CancelFunc
	exit   chan struct{} // channel to signal the default acction that exit is safe
}

// NewWalletClient returns a wallet client with ctx and exit channel
func NewWalletClient() *WalletClient {
	re := WalletClient{}
	re.ctx, re.cancel = context.WithTimeout(context.Background(), 30*time.Second)
	re.exit = make(chan struct{})
	return &re
}

// Exit signal that exit is allowed
func (w *WalletClient) Exit() {
	if w.client != nil {
		w.client.Close()
	}
	w.cancel()
	close(w.exit) // close channel to signal exit
}

// WaitExit wait for the signal of exit
func (w *WalletClient) WaitExit() {
	<-w.exit
}

func init() {
	app.Flags = flags
	app.Commands = commands
	app.Action = defaulWork

	app.Before = func(ctx *cli.Context) error {
		client, err := ethclient.Dial(ctx.String("nodeaddr"))
		if err != nil {
			return err
		}
		wallet.client = client
		return nil
	}

	app.After = func(ctx *cli.Context) error {
		wallet.WaitExit()
		return nil
	}
}

func defaulWork(c *cli.Context) error {
	client := wallet.client
	syncing, err := client.SyncProgress(wallet.ctx)
	if err != nil {
		return err
	}
	if syncing != nil {
		return fmt.Errorf("syncing %v %v %v %v %v", syncing.StartingBlock, syncing.CurrentBlock, syncing.HighestBlock,
			syncing.PulledStates, syncing.KnownStates)
	}
	balance, err := client.BalanceAt(wallet.ctx, account, nil)
	if err == nil {
		var fBalance = big.NewFloat(0).SetInt(balance)
		fmt.Println(fBalance.Quo(fBalance, big.NewFloat(100000000000000000)).String())
	}

	wallet.Exit()

	return err
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func newAccount(c *cli.Context) error {
	return nil
}

func importKeystore(c *cli.Context) error {
	return nil
}

func exportKeystore(c *cli.Context) error {
	return nil
}

func checkBalance(c *cli.Context) error {
	return nil
}

func transfer(c *cli.Context) error {
	return nil
}
