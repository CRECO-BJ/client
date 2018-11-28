package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"os"
	"strconv"
	"syscall"
	"unicode"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh/terminal"
)

const etherWei = 100000000000000000

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
			Aliases: []string{"n"},
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

func init() {
	wallet = NewWalletClient()
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
		//		wallet.Wait()
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

	//	go wallet.Exit()

	return err
}

func main() {
	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func newAccount(c *cli.Context) error {
	fmt.Printf("Please input the password for the new account:")
	bytePassword, err := terminal.ReadPassword(int(syscall.Stdin))
	if err == nil {
		return err
	}
	fmt.Println("\nPassword typed: " + string(bytePassword))
	password := string(bytePassword)

	a, err := wallet.keys.NewAccount(password)
	if err != nil {
		return err
	}

	fmt.Println("Account ", a, " has been created and added to the wallet")

	return nil
}

// verifyPassword check the password string:
// > 8 letters, has number, has uppercase, and has specialChar
func verifyPassword(s string) bool {
	letters := 0
	var hasNum, hasUp, hasSpecial bool
	for _, s := range s {
		switch {
		case unicode.IsNumber(s):
			hasNum = true
		case unicode.IsUpper(s):
			hasUp = true
			letters++
		case unicode.IsPunct(s) || unicode.IsSymbol(s):
			hasSpecial = true
		case unicode.IsLetter(s) || s == ' ':
			letters++
		default: // letter not allowed
			return false
		}
	}
	return hasNum && hasUp && hasSpecial && (letters >= 8)
}

func importKeystore(c *cli.Context) error {
	passwd := c.String("passwd")
	keyfile := c.String("keyfile")

	keyJSON, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return err
	}

	_, err = wallet.keys.Import(keyJSON, passwd, passwd)
	if err != nil {
		return err
	}

	return nil
}

func exportKeystore(c *cli.Context) error {
	address := c.String("account")
	account := accounts.Account{Address: common.HexToAddress(address)}
	account, err := wallet.keys.Find(account)
	if err != nil {
		return err
	}

	passwd := c.String("passwd")
	keyJSON, err := wallet.keys.Export(account, passwd, passwd)
	if err != nil {
		return err
	}

	keyfile := c.String("keyfile")
	err = ioutil.WriteFile(keyfile, keyJSON, 0755)
	if err != nil {
		return err
	}

	return nil
}

func checkBalance(c *cli.Context) error {
	address := common.HexToAddress(c.String("address"))

	balance, err := wallet.client.BalanceAt(wallet.ctx, address, nil)
	if err != nil {
		return err
	}

	var fBalance = big.NewFloat(0).SetInt(balance)
	fmt.Println(fBalance.Quo(fBalance, big.NewFloat(etherWei)).String())

	return nil
}

func transfer(c *cli.Context) error {
	from := common.HexToAddress(c.String("from"))
	account := accounts.Account{Address: from}
	account, err := wallet.keys.Find(account)
	if err != nil {
		return err
	}
	passwd := c.String("passwd")

	to := common.HexToAddress(c.String("to"))
	ta, err := strconv.ParseFloat(c.String("amount"), 64)
	if err != nil {
		return err
	}
	amount := big.NewInt(int64(ta * etherWei))

	nounce, err := wallet.nonceAt(account)
	gasLimit, err := wallet.gasLimitRecommended()
	gasPrice, err := wallet.gasPriceRecommended()
	tx := types.NewTransaction(nounce, to, amount, gasLimit.Uint64(), gasPrice, nil)

	id, err := wallet.chainID()
	tx, err = wallet.keys.SignTxWithPassphrase(account, passwd, tx, id)
	if err != nil {
		return err
	}

	// send tx and check

	return nil
}

func tmpKeyStore() *keystore.KeyStore {
	d, err := ioutil.TempDir("", ".wallet-creco")
	if err != nil {
		panic(err)
	}

	return keystore.NewKeyStore(d, 2, 1)
}

func transactionHistory(c *cli.Context) error {
	return nil
}
