package cli

import (
	"context"
	"flag"
	"fmt"

	"github.com/nikola43/stardust/config"
)

var (
	configFile string
	update     string
	etcd       bool
	cfg        *config.Config
)

func init() {
	flag.StringVar(&configFile, "config", "", "configuration file")
	flag.StringVar(&update, "update", "", "update etc / file")
	flag.BoolVar(&etcd, "etcd", false, "enable etcd")
	flag.Parse()
}

type CreateCommand struct {
	args *Args
}

/*

master-wallet
$ stardust create btc-wallet
$ stardust create eth-wallet
$ stardust create network
$ stardust create network 3
$ stardust create node 2cfd7f563b
*/

const (
	createCommand     = "create"
	createDescription = "create wallets, networks, nodes"

	createMasterWallet = "master-wallet"
	createBTCWallet    = "btc-wallet"
	createEthWallet    = "eth-wallet"
	createNetwork      = "network"
	createNode         = "node"
)

func newCreateCommand() Command {
	return Command{
		Name:        createCommand,
		Description: createDescription,
		Exec:        &CreateCommand{},
	}
}

func (c *CreateCommand) createMasterWallet() error {
	fmt.Println("create master wallet")
	return nil
}

func (c *CreateCommand) createBTCWallet() error {
	fmt.Println("create BTC wallet")
	return nil
}

func (c *CreateCommand) createEthWallet() error {
	fmt.Println("create ETH wallet")
	return nil
}

func (c *CreateCommand) createNetwork() error {
	networkTypeNumbder := c.args.pop()
	if networkTypeNumbder == "" {
		fmt.Println("create network")

		return nil
	}

	fmt.Printf("create network type with number: %s", networkTypeNumbder)
	return nil
}

func (c *CreateCommand) createNode() error {
	nodeHash := c.args.pop()
	if nodeHash == "" {
		return ErrorFromString(fmt.Sprintf("%s-node: invalid node hash", createCommand))
	}

	fmt.Println("create node")
	return nil
}

func (c *CreateCommand) ExecCommand(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return ErrorFromString(fmt.Sprintf("%s: no subcommand passed", createCommand))
	}
	c.args = &Args{args}

	subcommand := c.args.pop()
	switch subcommand {
	case createMasterWallet:
		return c.createMasterWallet()
	case createBTCWallet:
		return c.createBTCWallet()
	case createEthWallet:
		return c.createEthWallet()
	case createNode:
		return c.createNode()
	case createNetwork:
		return c.createNetwork()
	}

	return ErrorFromString(fmt.Sprintf("%s: invalid subcommand passed", createCommand))
}
