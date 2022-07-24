package wallet

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	"github.com/btcsuite/btcd/btcutil"
	secp256k1 "github.com/decred/dcrd/dcrec/secp256k1/v4"
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

	pk, err := GenerateEcdsaPrivateKey()
	if err != nil {
		log.Fatal(err)
	}

	//s := GenerateETHWalletFromPrivateKEy(pk)
	//fmt.Println(s)

	masterWallet := new(MasterWallet)
	masterWallet.BtcWallet = GenerateBTCWallet()
	masterWallet.EthWallet = GenerateETHWallet()

	btcPkSkeinHash := HashSkein1024([]byte(masterWallet.BtcWallet.PrivateKey))
	ethPkSkeinHash := HashSkein1024([]byte(masterWallet.EthWallet.PrivateKey))
	l := make([]byte, len(ethPkSkeinHash))
	for i := 0; i < len(l); i++ {
		l[i] = btcPkSkeinHash[i]
		if i > 512 {
			l[i] = ethPkSkeinHash[i]
		}
	}

	masterPrivate := HashSkein1024(l[:128])
	masterPublicKey := HashSkein1024([]byte(masterPrivate[:128]))
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

func (mw MasterWallet) MasterAddressFromBtcEthPrivateKey(btcPk, ethPk string) *MasterWallet {

	masterWallet := new(MasterWallet)
	btcPkSkeinHash := HashSkein1024([]byte(btcPk))
	ethPkSkeinHash := HashSkein1024([]byte(ethPk))

	l := make([]byte, len(ethPkSkeinHash))
	for i := 0; i < len(l); i++ {
		l[i] = btcPkSkeinHash[i]
		if i > 512 {
			l[i] = ethPkSkeinHash[i]
		}
	}

	masterPrivate := HashSkein1024(l[:128])
	masterPublicKey := HashSkein1024([]byte(masterPrivate[:128]))
	masterWallet.PrivateKey = hex.EncodeToString(masterPrivate)[:256]
	masterWallet.PublicKey = hex.EncodeToString(masterPublicKey)[:43]

	return masterWallet
}

func (mw MasterWallet) MasterAddressFromPrivateKey(masterPrivate []byte) *MasterWallet {
	masterWallet := new(MasterWallet)
	masterPublicKey := HashSkein1024([]byte(masterPrivate[:64]))
	masterWallet.PrivateKey = hex.EncodeToString(masterPrivate)[:256]
	masterWallet.PublicKey = hex.EncodeToString(masterPublicKey)[:43]
	return masterWallet
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
	outputBuffer := make([]byte, 1024)
	sk.Final(outputBuffer)
	//return hex.EncodeToString(outputBuffer)
	return outputBuffer
}

// GeneratePrivateKey returns a private key that is suitable for use with
// secp256k1.
func GenerateEcdsaPrivateKey() (*ecdsa.PrivateKey, error) {
	key, err := ecdsa.GenerateKey(crypto.S256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// GeneratePrivateKey returns a private key that is suitable for use with
// secp256k1.
func GeneratePrivateKey(pk *ecdsa.PrivateKey) (*secp256k1.PrivateKey, error) {
	return secp256k1.PrivKeyFromBytes(pk.D.Bytes()), nil
}

func CreateBTCWifFromPk(pk *secp256k1.PrivateKey) *btcutil.WIF {
	wif, err := networks["btc"].CreateWifFromPk(pk)
	if err != nil {
		log.Fatal(err)
	}
	return wif
}

func GenerateBTCWallet() Wallet {

	wif, err := networks["btc"].CreateWif()
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

func GenerateETHWalletFromPrivateKEy(privateKey *ecdsa.PrivateKey) Wallet {
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
