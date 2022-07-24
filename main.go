package main

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/fatih/color"
	"github.com/nikola43/stardust/NodeManagerV83"
	"github.com/nikola43/stardust/config"
	"github.com/nikola43/stardust/router"
	"github.com/nikola43/stardust/server"
	sysinfo "github.com/nikola43/stardust/sysinfo"
	wallet "github.com/nikola43/stardust/wallet"
	"github.com/nikola43/web3golanghelper/web3helper"

	mrand "math/rand"

	"github.com/davecgh/go-spew/spew"
	golog "github.com/ipfs/go-log"
	log2 "github.com/ipfs/go-log/v2"
	libp2p "github.com/libp2p/go-libp2p"
	crypto "github.com/libp2p/go-libp2p-crypto"
	host "github.com/libp2p/go-libp2p-host"
	net "github.com/libp2p/go-libp2p-net"
	peer "github.com/libp2p/go-libp2p-peer"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
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

// Block represents each 'item' in the blockchain
type Block struct {
	Index     int
	Timestamp string
	BPM       int
	Hash      string
	PrevHash  string
}

// Blockchain is a series of validated Blocks
var Blockchain []Block

var mutex = &sync.Mutex{}

// makeBasicHost creates a LibP2P host with a random peer ID listening on the
// given multiaddress. It will use secio if secio is true.
func makeBasicHost(listenPort int, secio bool, randseed int64) (host.Host, error) {

	// If the seed is zero, use real cryptographic randomness. Otherwise, use a
	// deterministic randomness source to make generated keys stay the same
	// across multiple runs
	var r io.Reader
	if randseed == 0 {
		r = rand.Reader
	} else {
		r = mrand.New(mrand.NewSource(randseed))
	}

	// Generate a key pair for this host. We will use it
	// to obtain a valid host ID.
	priv, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		return nil, err
	}

	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", listenPort)),
		libp2p.Identity(priv),
	}

	basicHost, err := libp2p.New(opts...)
	if err != nil {
		return nil, err
	}

	// Build host multiaddress
	hostAddr, _ := ma.NewMultiaddr(fmt.Sprintf("/ipfs/%s", basicHost.ID().Pretty()))

	// Now we can build a full multiaddress to reach this host
	// by encapsulating both addresses:
	addrs := basicHost.Addrs()
	var addr ma.Multiaddr
	// select the address starting with "ip4"
	for _, i := range addrs {
		if strings.HasPrefix(i.String(), "/ip4") {
			addr = i
			break
		}
	}
	fullAddr := addr.Encapsulate(hostAddr)
	log.Printf("I am %s\n", fullAddr)
	if secio {
		log.Printf("Now run \"go run main.go -l %d -d %s -secio\" on a different terminal\n", listenPort+1, fullAddr)
	} else {
		log.Printf("Now run \"go run main.go -l %d -d %s\" on a different terminal\n", listenPort+1, fullAddr)
	}

	return basicHost, nil
}

func handleStream(s net.Stream) {

	log.Println("Got a new stream!")

	// Create a buffer stream for non blocking read and write.
	rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

	go readData(rw)
	go writeData(rw)

	// stream 's' will stay open until you close it (or the other side closes it).
}

func readData(rw *bufio.ReadWriter) {

	for {
		str, err := rw.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		if str == "" {
			return
		}
		if str != "\n" {

			chain := make([]Block, 0)
			if err := json.Unmarshal([]byte(str), &chain); err != nil {
				log.Fatal(err)
			}

			mutex.Lock()
			if len(chain) > len(Blockchain) {
				Blockchain = chain
				bytes, err := json.MarshalIndent(Blockchain, "", "  ")
				if err != nil {

					log.Fatal(err)
				}
				// Green console color: 	\x1b[32m
				// Reset console color: 	\x1b[0m
				fmt.Printf("\x1b[32m%s\x1b[0m> ", string(bytes))
			}
			mutex.Unlock()
		}
	}
}

func writeData(rw *bufio.ReadWriter) {

	go func() {
		for {
			time.Sleep(5 * time.Second)
			mutex.Lock()
			bytes, err := json.Marshal(Blockchain)
			if err != nil {
				log.Println(err)
			}
			mutex.Unlock()

			mutex.Lock()
			rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
			rw.Flush()
			mutex.Unlock()

		}
	}()

	stdReader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		sendData, err := stdReader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		sendData = strings.Replace(sendData, "\n", "", -1)
		bpm, err := strconv.Atoi(sendData)
		if err != nil {
			log.Fatal(err)
		}
		newBlock := generateBlock(Blockchain[len(Blockchain)-1], bpm)

		if isBlockValid(newBlock, Blockchain[len(Blockchain)-1]) {
			mutex.Lock()
			Blockchain = append(Blockchain, newBlock)
			mutex.Unlock()
		}

		bytes, err := json.Marshal(Blockchain)
		if err != nil {
			log.Println(err)
		}

		spew.Dump(Blockchain)

		mutex.Lock()
		rw.WriteString(fmt.Sprintf("%s\n", string(bytes)))
		rw.Flush()
		mutex.Unlock()
	}

}

func main() {

	// CHAIN

	t := time.Now()
	genesisBlock := Block{}
	genesisBlock = Block{0, t.String(), 0, calculateHash(genesisBlock), ""}

	Blockchain = append(Blockchain, genesisBlock)

	// LibP2P code uses golog to log messages. They log with different
	// string IDs (i.e. "swarm"). We can control the verbosity level for
	// all loggers with:
	golog.SetAllLoggers(log2.LevelDebug) // Change to DEBUG for extra info

	// Parse options from the command line
	listenF := flag.Int("l", 0, "wait for incoming connections")
	target := flag.String("d", "", "target peer to dial")
	secio := flag.Bool("secio", false, "enable secio")
	seed := flag.Int64("seed", 0, "set random seed for id generation")
	flag.Parse()

	if *listenF == 0 {
		log.Fatal("Please provide a port to bind on with -l")
	}

	// Make a host that listens on the given multiaddress
	ha, err := makeBasicHost(*listenF, *secio, *seed)
	if err != nil {
		log.Fatal(err)
	}

	if *target == "" {
		log.Println("listening for connections")
		// Set a stream handler on host A. /p2p/1.0.0 is
		// a user-defined protocol name.
		ha.SetStreamHandler("/p2p/1.0.0", handleStream)

		select {} // hang forever
		/**** This is where the listener code ends ****/
	} else {
		ha.SetStreamHandler("/p2p/1.0.0", handleStream)

		// The following code extracts target's peer ID from the
		// given multiaddress
		ipfsaddr, err := ma.NewMultiaddr(*target)
		if err != nil {
			log.Fatalln(err)
		}

		pid, err := ipfsaddr.ValueForProtocol(ma.P_IPFS)
		if err != nil {
			log.Fatalln(err)
		}

		peerid, err := peer.IDB58Decode(pid)
		if err != nil {
			log.Fatalln(err)
		}

		// Decapsulate the /ipfs/<peerID> part from the target
		// /ip4/<a.b.c.d>/ipfs/<peer> becomes /ip4/<a.b.c.d>
		targetPeerAddr, _ := ma.NewMultiaddr(
			fmt.Sprintf("/ipfs/%s", peer.IDB58Encode(peerid)))
		targetAddr := ipfsaddr.Decapsulate(targetPeerAddr)

		// We have a peer ID and a targetAddr so we add it to the peerstore
		// so LibP2P knows how to contact it
		//ha.Peerstore().AddAddr(peerid, targetAddr, pstore.PermanentAddrTTL)
		ha.Peerstore().AddAddr(peerid, targetAddr, time.Second * 5)

		log.Println("opening stream")
		// make a new stream from host B to host A
		// it should be handled on host A by the handler we set above because
		// we use the same /p2p/1.0.0 protocol
		s, err := ha.NewStream(context.Background(), peerid, "/p2p/1.0.0")
		if err != nil {
			log.Fatalln(err)
		}
		// Create a buffered stream so that read and writes are non blocking.
		rw := bufio.NewReadWriter(bufio.NewReader(s), bufio.NewWriter(s))

		// Create a thread to read and write data.
		go writeData(rw)
		go readData(rw)

		select {} // hang forever

	}
	// -------------

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

// make sure block is valid by checking index, and comparing the hash of the previous block
func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}

	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}

	if calculateHash(newBlock) != newBlock.Hash {
		return false
	}

	return true
}

// SHA256 hashing
func calculateHash(block Block) string {
	record := strconv.Itoa(block.Index) + block.Timestamp + strconv.Itoa(block.BPM) + block.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

// create a new block using previous block's hash
func generateBlock(oldBlock Block, BPM int) Block {

	var newBlock Block

	t := time.Now()

	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = t.String()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Hash = calculateHash(newBlock)

	return newBlock
}
