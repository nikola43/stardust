package cli

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/fatih/color"
	"github.com/nikola43/stardust/NodeManagerV83"
	"github.com/nikola43/stardust/config"
	"github.com/nikola43/stardust/router"
	"github.com/nikola43/stardust/server"
	"github.com/nikola43/stardust/sysinfo"
	"github.com/nikola43/stardust/wallet"
	"github.com/nikola43/web3golanghelper/web3helper"
	"gopkg.in/yaml.v2"
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
		go RunServer()
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

func RunServer() {

	mw := wallet.NewMasterWallet()

	// system config
	numCpu := runtime.NumCPU()
	usedCpu := numCpu
	runtime.GOMAXPROCS(usedCpu)

	PrintSystemInfo(numCpu, usedCpu)
	PrintNetworkStatus()
	PrintUserBalance(mw.PublicKey, 932)
	PrintUserBalance2(mw.PublicKey, 923)

	mw.ToString()

	//sysHash := GetSysInfo()

	//key := make([]byte, 32)
	//rand.Read(key)
	//fmt.Println(key)
	//crypto.EncryptSysData([]byte(sysHash), []byte(key))
	//crypto.DecryptFile("sysdata.txt.bin", []byte(key))
	//os.Exit(0)

	// create unix syscall
	sig := make(chan os.Signal, 1)
	notify := make(chan struct{}, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	UpdateEtcdConf()

	// get etcd config
	cfg := GetEtcdConfig()
	cfg.Watcher(ctx, notify)

	// init server
	r := router.New(ctx)
	s := server.Server{
		Config: cfg,
		Router: r,
		Notify: notify,
	}
	s.Run(ctx)
	<-sig
}

type Conf struct {
	Revision uint32 `yaml:"revision"`
	Etcd     Etcd   `yaml:"etcd"`
	Server   Server `yaml:"server"`
	Crypto   Crypto `yaml:"crypto"`
	Nodes    []Node `yaml:"nodes"`
}

type Etcd struct {
	Endpoints []string `yaml:"endpoints"`
	Timeout   uint32   `yaml:"timeout"`
}

type Server struct {
	Keepalive uint32 `yaml:"keepalive"`
	Insecure  bool   `yaml:"insecure"`
	Mtu       uint32 `yaml:"mtu"`
}

type Crypto struct {
	Type string `yaml:"type"`
	Key  string `yaml:"key"`
}

type Node struct {
	Node NodeInfo `yaml:"node"`
}

type NodeInfo struct {
	Name             string   `yaml:"name"`
	Address          string   `yaml:"address"`
	PrivateAddresses []string `yaml:"privateAddresses"`
	PrivateSubnets   []string `yaml:"privateSubnets"`
}

func GetConf() *Conf {
	var c *Conf

	yamlFile, err := ioutil.ReadFile("./stardust.yaml")
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}

func WriteConf(localIp net.IP) {
	c := GetConf()
	var a []string
	var b []string
	a = append(a, "10.110.0.4/24")
	b = append(b, "10.110.0.0/24")

	name := "node" + (strconv.Itoa(len(c.Nodes) + 1))
	config := Node{Node: NodeInfo{name, localIp.String(), a, b}}

	c.Nodes = append(c.Nodes, config)

	data, err := yaml.Marshal(&c)

	if err != nil {

		log.Fatal(err)
	}

	err2 := ioutil.WriteFile("stardust.yaml", data, 0)

	if err2 != nil {

		log.Fatal(err2)
	}

	fmt.Println("data written")
}

func CreateOwnConf(localIp net.IP) {
	var c Conf
	c.Revision = 1
	c.Etcd.Endpoints = append(c.Etcd.Endpoints, "localhost:2379")
	c.Etcd.Timeout = 5
	c.Server.Keepalive = 10
	c.Server.Insecure = false
	c.Server.Mtu = 1300
	c.Crypto.Type = "gcm"
	c.Crypto.Key = "6368616e676520746869732070617373776f726420746f206120736563726574"

	var a []string
	var b []string
	a = append(a, "10.110.0.4/24")
	b = append(b, "10.110.0.0/24")

	name := "node1"
	config := Node{Node: NodeInfo{name, localIp.String(), a, b}}
	c.Nodes = append(c.Nodes, config)

	data, err := yaml.Marshal(&c)

	if err != nil {

		log.Fatal(err)
	}

	err2 := ioutil.WriteFile("stardustNew.yaml", data, 0)

	if err2 != nil {

		log.Fatal(err2)
	}

	fmt.Println("new network config created")
}

func InitServer(octx context.Context, notify *chan struct{}) {
	r := router.New(octx)
	s := server.Server{
		Config: cfg,
		Router: r,
		Notify: *notify,
	}

	s.Run(octx)
}

func UpdateEtcdConf() {
	// check if we need update nodes config file
	if update != "" {
		err := config.UpdateConf(update, configFile)
		if err != nil {
			fmt.Println("UpdateConf")
			panic(err)
		}
		os.Exit(0)
	}
}

func GetEtcdConfig() *config.Config {
	var cfg *config.Config

	// get etcd config
	if etcd {
		cfg = config.New().FromEtcd(configFile)
	} else {
		cfg = config.New().FromFile(configFile)
	}
	err := cfg.Load()
	if err != nil {
		log.Fatal(err)
	}

	return cfg
}

func GetSysInfo() string {

	info := sysinfo.NewSysInfo()
	fmt.Printf("%+v\n", info)
	fmt.Printf("%+s\n", info.ToString())
	fmt.Printf("%+s\n", info.ToHash())
	return info.ToHash()
}

func InitWeb3() {
	pk := "b366406bc0b4883b9b4b3b41117d6c62839174b7d21ec32a5ad0cc76cb3496bd"
	rpcUrl := "https://speedy-nodes-nyc.moralis.io/84a2745d907034e6d388f8d6/avalanche/testnet"
	wsUrl := "wss://speedy-nodes-nyc.moralis.io/84a2745d907034e6d388f8d6/avalanche/testnet/ws"
	web3GolangHelper := web3helper.NewWeb3GolangHelper(rpcUrl, wsUrl, pk)

	chainID, err := web3GolangHelper.HttpClient().NetworkID(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Chain Id: " + chainID.String())

	proccessEvents(web3GolangHelper)
}

func proccessEvents(web3GolangHelper *web3helper.Web3GolangHelper) {
	nodeAddress := "0x2Fcd73952e53aAd026c378F378812E5bb069eF6E"
	nodeAbi, _ := abi.JSON(strings.NewReader(string(NodeManagerV83.NodeManagerV83ABI)))
	fmt.Println(color.YellowString("  ----------------- Blockchain Events -----------------"))
	fmt.Println(color.CyanString("\tListen node manager address: "), color.GreenString(nodeAddress))
	logs := make(chan types.Log)
	sub := BuildContractEventSubscription(web3GolangHelper, nodeAddress, logs)

	for {
		select {
		case err := <-sub.Err():
			fmt.Println(err)
			//out <- err.Error()

		case vLog := <-logs:
			fmt.Println("paco")
			fmt.Println("vLog.TxHash: " + vLog.TxHash.Hex())
			res, err := nodeAbi.Unpack("GiftCardPayed", vLog.Data)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(res)
			//services.SetGiftCardIntentPayment(res[2].(string))

		}
	}
}

func BuildContractEventSubscription(web3GolangHelper *web3helper.Web3GolangHelper, contractAddress string, logs chan types.Log) ethereum.Subscription {

	query := ethereum.FilterQuery{
		Addresses: []common.Address{common.HexToAddress(contractAddress)},
	}

	sub, err := web3GolangHelper.WebSocketClient().SubscribeFilterLogs(context.Background(), query, logs)
	if err != nil {
		fmt.Println(sub)
	}
	return sub
}

func PrintSystemInfo(numCpu, usedCpu int) {
	fmt.Println("")
	fmt.Println(color.YellowString("  ----------------- System Info -----------------"))
	fmt.Println(color.CyanString("\t    Number CPU cores available: "), color.GreenString(strconv.Itoa(numCpu)))
	fmt.Println(color.CyanString("\t    Used of CPU cores: "), color.YellowString(strconv.Itoa(usedCpu)))
	fmt.Println()
}

func PrintNetworkStatus() {
	fmt.Println(color.YellowString("  ----------------- Network Info -----------------"))
	fmt.Println(color.CyanString("\t    Number Nodes: "), color.YellowString(strconv.Itoa(3)))
	fmt.Println(color.CyanString("\t    Prague: "), color.YellowString(strconv.Itoa(1)))
	fmt.Println(color.CyanString("\t    Kiev: "), color.YellowString(strconv.Itoa(1)))
	fmt.Println(color.CyanString("\t    Singapour: "), color.YellowString(strconv.Itoa(1)))
	fmt.Println()
}

func PrintUserBalance(address string, balance int) {
	fmt.Println(color.YellowString("  ----------------- Node Owner -----------------"))
	fmt.Println(color.CyanString("  "), color.GreenString(address))
	fmt.Println(color.CyanString("\t    Balance: "), color.YellowString(strconv.Itoa(balance)), color.YellowString(" $ZOE"))
	fmt.Println()
}

func PrintUserBalance2(address string, balance int) {
	fmt.Println(color.YellowString("  ----------------- Network Info -----------------"))
	fmt.Println(color.CyanString("\t    Send: "), color.YellowString(strconv.Itoa(1732)), color.YellowString("MB"))
	fmt.Println(color.CyanString("\t    Received: "), color.YellowString(strconv.Itoa(1343)), color.YellowString("MB"))
	fmt.Println(color.CyanString("\t    Duration: "), color.YellowString("19:20:04"))
	fmt.Println(color.RedString("\t    Paid: "), color.YellowString(strconv.Itoa(1)), color.YellowString("$ZOE = 2.52$"))
	fmt.Println()
}
