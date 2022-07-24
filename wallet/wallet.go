package wallet

import (
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	skein "github.com/nikola43/stardust/crypto"
	"golang.org/x/crypto/sha3"
)

type Wallet struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
}

// SysInfo saves the basic system information
type MasterWallet struct {
	PublicKey  string `json:"public_key"`
	PrivateKey string `json:"private_key"`
	BtcWallet  Wallet `json:"btc_wallet"`
	EthWallet  Wallet `json:"eth_wallet"`
}

func NewMasterWallet() *MasterWallet {
	masterWallet := new(MasterWallet)
	masterWallet.BtcWallet = GenerateBTCWallet()
	masterWallet.EthWallet = GenerateETHWallet()

	btcPkSkeinHash := HashSkein1024([]byte(masterWallet.BtcWallet.PrivateKey))
	ethPkSkeinHash := HashSkein1024([]byte(masterWallet.EthWallet.PrivateKey))

	fmt.Println("len", len(btcPkSkeinHash))
	fmt.Println("len", len(ethPkSkeinHash))

	fmt.Println(len(ethPkSkeinHash))
	l := make([]byte, 1024)
	fmt.Println(len(l))
	for i := 0; i < 512; i++ {
		l[i] = btcPkSkeinHash[i]
		if i > 256 {
			l[i] = ethPkSkeinHash[i]
		}
	}
	//fmt.Println("len l", len(l))
	//fmt.Println(hex.EncodeToString(btcPkSkeinHash))
	//fmt.Println(hex.EncodeToString(ethPkSkeinHash))

	masterPrivate := HashSkein1024(btcPkSkeinHash[:64])
	masterPublicKey := HashSkein1024([]byte(masterPrivate[:64]))

	masterWallet.PrivateKey = hex.EncodeToString(masterPrivate)[:256]
	masterWallet.PublicKey = hex.EncodeToString(masterPublicKey)[:43]

	return masterWallet
}

func (mw MasterWallet) ToString() {
	fmt.Println("BTC Public Key", mw.BtcWallet.PublicKey)
	fmt.Println("BTC Private Key", mw.BtcWallet.PrivateKey)
	fmt.Println("ETH Public Key", mw.EthWallet.PublicKey)
	fmt.Println("ETH Private Key", mw.EthWallet.PrivateKey)
	fmt.Println("KAS Public Key", mw.PublicKey)
	fmt.Println("KAS Private Key", mw.PrivateKey)
}

func (mw MasterWallet) MasterAddressFromBtcEthPrivateKey(btcPk, ethPk string) string {
	btcPkSkeinHash := HashSkein1024([]byte(btcPk))
	//ethPkSkeinHash := HashSkein1024([]byte(ethPk))

	return string(btcPkSkeinHash)
}

func (mw MasterWallet) MasterAddressFromPrivateKey(masterPrivate string) {

}

func (mw MasterWallet) MasterAddress() string {
	return mw.PublicKey
}

func (mw MasterWallet) EthAddress() string {
	return mw.EthWallet.PublicKey
}

func (mw MasterWallet) BtcAddress() string {
	return mw.BtcWallet.PublicKey
}

func HashSkein1024(data []byte) []byte {
	sk := new(skein.Skein1024)
	sk.Init(1024)
	sk.Update(data)
	outputBuffer := make([]byte, 512)
	sk.Final(outputBuffer)
	//return hex.EncodeToString(outputBuffer)
	return outputBuffer
}

func GenerateBTCPrivateKey() string {
	wif, err := networks["btc"].CreatePrivateKey()
	if err != nil {
		log.Fatal(err)
	}
	pk := wif.String()
	return pk
}

func GenerateBTCWallet() Wallet {
	wif, err := networks["btc"].CreatePrivateKey()
	if err != nil {
		log.Fatal(err)
	}
	pk := wif.String()

	address, err := networks["btc"].GetAddress(wif)
	if err != nil {
		log.Fatal(err)
	}
	wallet := Wallet{
		PublicKey:  address.EncodeAddress(),
		PrivateKey: pk,
	}
	return wallet
}

func GenerateETHPrivateKey() string {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	publicKey := privateKey.Public()
	_, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}
	return hexutil.Encode(privateKeyBytes)[2:]
}

func GenerateETHWallet() Wallet {
	privateKey, err := crypto.GenerateKey()
	if err != nil {
		log.Fatal(err)
	}
	privateKeyBytes := crypto.FromECDSA(privateKey)
	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("cannot assert type: publicKey is not of type *ecdsa.PublicKey")
	}

	publicKeyBytes := crypto.FromECDSAPub(publicKeyECDSA)
	address := crypto.PubkeyToAddress(*publicKeyECDSA).Hex()
	hash := sha3.NewLegacyKeccak256()
	hash.Write(publicKeyBytes[1:])

	wallet := Wallet{
		PublicKey:  address,
		PrivateKey: hexutil.Encode(privateKeyBytes)[2:],
	}
	return wallet
}

func GenerateAddressFromPlainPrivateKey(pk string) (common.Address, error) {
	var address common.Address
	privateKey, err := crypto.HexToECDSA(pk)
	if err != nil {
		return address, err
	}

	publicKeyECDSA, ok := privateKey.Public().(*ecdsa.PublicKey)
	if !ok {
		return address, errors.New("error casting public key to ECDSA")
	}

	return crypto.PubkeyToAddress(*publicKeyECDSA), nil
}
