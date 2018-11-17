package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli"
)

var account = common.HexToAddress("0x8f2d2b848ede60d9480631fe6a365cbc8e304c14")

var flags = []cli.Flag{
	cli.StringFlag{
		Name:  "nodeaddr",
		Usage: "the address of the eth node/server",
		Value: "http://139.9.2.54:4567",
	},
}

func work(c *cli.Context) error {
	client, err := ethclient.Dial(c.String("nodeaddr"))
	if err != nil {
		return err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	fmt.Println(client.NetworkID(ctx))

	syncing, err := client.SyncProgress(ctx)
	if err != nil {
		return err
	}
	if syncing != nil {
		return fmt.Errorf("syncing %v %v %v %v %v", syncing.StartingBlock, syncing.CurrentBlock, syncing.HighestBlock,
			syncing.PulledStates, syncing.KnownStates)
	}
	fmt.Println(client.BalanceAt(ctx, account, nil))

	return nil
}

func main() {
	app := cli.NewApp()
	app.Flags = flags
	app.Action = work

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
