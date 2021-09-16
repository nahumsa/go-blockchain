package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"golang.org/x/crypto/ripemd160"
)

const (
	checksumLenght = 4
	version        = byte(0x00)
)

// Wallet keeps a public and a private key
type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func (w Wallet) Address() []byte {
	pubHash := publicKeyHash(w.PublicKey)

	versionedHash := append([]byte{version}, pubHash...)
	checksum := checkSum(versionedHash)

	fullHash := append(versionedHash, checksum...)

	address := base58Encode(fullHash)

	return address
}

// NewKeyPair generates a public and a private key
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256()

	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	pub := append(private.X.Bytes(), private.Y.Bytes()...)
	return *private, pub
}

// MakeWallet creates a wallet with a public and a private key
func MakeWallet() *Wallet {
	private, public := NewKeyPair()

	wallet := Wallet{private, public}

	return &wallet
}

func publicKeyHash(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)

	hasher := ripemd160.New()
	_, err := hasher.Write(pubHash[:])
	if err != nil {
		log.Panic(err)
	}

	publicRipMD := hasher.Sum(nil)
	return publicRipMD
}

func checkSum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:checksumLenght]
}
