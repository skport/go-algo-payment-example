package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	// "log"
	"strings"
	"bufio"
	"os"

	"crypto/sha512"
	"encoding/base32"

	"golang.org/x/crypto/ed25519"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/encoding/msgpack"
	// "github.com/algorand/go-algorand-sdk/transaction"
	"github.com/algorand/go-algorand-sdk/types"
)

const (
	minFee      = 1000
	genesisId   = "testnet-v1.0"
	genesisHash = "SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI="
)

func main() {
	FromAccount := crypto.GenerateAccount()
	fmt.Println("Algorand FromAddress", FromAccount.Address.String())

	ToAccount := crypto.GenerateAccount()
	fmt.Println("Algorand TOAddress", ToAccount.Address.String())

	// Validate params for transaction
	var ghDigest  types.Digest
	copy(ghDigest[:], byteFromBase64(genesisHash))

	// TRANSACTION_PAYMENT_CREATE
	unSignedTx := types.Transaction{
		Type: "pay",
		Header: types.Header{
			Sender:      FromAccount.Address,
			FirstValid:  types.Round(10000),
			LastValid:   types.Round(10000 + 1000),
			Note:        nil,
			GenesisID:   genesisId,
			GenesisHash: ghDigest,
		},
		PaymentTxnFields: types.PaymentTxnFields{
			Receiver:         ToAccount.Address,
			Amount:           types.MicroAlgos(1000000),
			CloseRemainderTo: types.Address{},
		},
	}
	unSignedTx.Fee = types.MicroAlgos(minFee)

	// Sign the encoded transaction

	// Encode the transaction as msgpack
	encodedTx := msgpack.Encode(unSignedTx)
	// Prepend the hashable prefix
	msgParts := [][]byte{[]byte("TX"), encodedTx}
	toBeSigned := bytes.Join(msgParts, nil)

	signature := ed25519.Sign(FromAccount.PrivateKey, toBeSigned)
	// Copy the resulting signature into a Signature, and check that it's
	// the expected length
	var algoSignature types.Signature
	n := copy(algoSignature[:], signature)
	if n != len(algoSignature) {
		fmt.Errorf("Failed to copy signature")
	}
	fmt.Println("Signature", algoSignature)

	// Create the SignedTxn
	stx := types.SignedTxn{
		Sig: algoSignature,
		Txn: unSignedTx,
	}

	// Encode the SignedTxn
	stxBytes := msgpack.Encode(stx)

	// Create the Transaction ID
	txid := txIDFromRawTxnBytesToSign(toBeSigned)
	fmt.Println("Transaction ID", txid)

	// Waiting for user input
	fmt.Print("continue? (y/n) >")
	scanner := bufio.NewScanner(os.Stdin)	
	scanner.Scan()
	in := scanner.Text()

	// Broadcast the transaction to the network
	switch in {
	case "y":
		fmt.Println("Sending transaction...")
		pendingTxID, err := sendTransaction(stxBytes)
		if err != nil {
			fmt.Printf("Failed to send transaction: %s\n", err)
		}
		fmt.Printf("Pending Transaction ID: %s\n", pendingTxID)
	default:
		break
	}
}

func sendTransaction(stxBytes []byte) (string, error) {
	var algodAddress = "https://testnet-api.algonode.cloud"
	var algodToken = strings.Repeat("a", 64)
	algodClient, _ := algod.MakeClient(
		algodAddress,
		algodToken,
	)

	pendingTxID, err := algodClient.SendRawTransaction(stxBytes).Do(context.Background())
	if err != nil {
		fmt.Printf("failed to send transaction: %s\n", err)
		return "", err
	}

	return pendingTxID, nil
}

func byteFromBase64(s string) []byte {
	b, _ := base64.StdEncoding.DecodeString(s)
	return b
}

func txIDFromRawTxnBytesToSign(toBeSigned []byte) (txid string) {
	txidBytes := sha512.Sum512_256(toBeSigned)
	txid = base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(txidBytes[:])
	return
}
