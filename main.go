package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	"crypto/sha512"
	"encoding/base32"

	"golang.org/x/crypto/ed25519"

	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/encoding/msgpack"
	"github.com/algorand/go-algorand-sdk/types"
)

const (
	// 最低手数料 1000 MicroAlgos
	minFee = 1000
	// ネットワークのバージョン
	genesisId = "testnet-v1.0"
	// 最初のブロックハッシュ (Base64)
	genesisHash = "SGO1GKSzyE7IEPItTxCByw9x8FmnrCDexi9/cOUJOiI="
	// 開始ラウンド（開始ブロックナンバー）
	lastRound = 100000
)

func main() {
	// 1. 送金元のアカウント作成
	FromAccount := crypto.GenerateAccount()
	fmt.Println("Algorand FromAddress", FromAccount.Address.String())
	fmt.Println("Algorand FromAddress PrivateKey", hex.EncodeToString(FromAccount.PrivateKey))
	fmt.Println("Algorand FromAddress PublicKey", hex.EncodeToString(FromAccount.PublicKey))

	// 2. 送金先のアドレス構造体を作成
	ToAccount := crypto.GenerateAccount()
	fmt.Println("Algorand ToAddress", ToAccount.Address.String())

	// GenesisHashを byte[] に変換
	var ghDigest types.Digest
	copy(ghDigest[:], byteFromBase64(genesisHash))

	// 3. 署名前のトランザクション作成
	unSignedTx := types.Transaction{
		Type: "pay",
		Header: types.Header{
			Sender:      FromAccount.Address,
			FirstValid:  types.Round(lastRound),
			LastValid:   types.Round(lastRound + 1000),
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

	// Encode the transaction as msgpack
	encodedTx := msgpack.Encode(unSignedTx)
	// Prepend the hashable prefix
	msgParts := [][]byte{[]byte("TX"), encodedTx}
	toBeSigned := bytes.Join(msgParts, nil)

	// 4. 署名の作成
	signature := ed25519.Sign(FromAccount.PrivateKey, toBeSigned)
	// Copy the resulting signature into a Signature,
	// and check that it's the expected length
	var algoSignature types.Signature
	n := copy(algoSignature[:], signature)
	if n != len(algoSignature) {
		log.Fatal("Failed to copy signature")
	}
	fmt.Println("Signature", algoSignature)

	// 5. 署名済みトランザクションの作成
	stx := types.SignedTxn{
		Sig: algoSignature,
		Txn: unSignedTx,
	}

	// Encode the SignedTxn
	stxBytes := msgpack.Encode(stx)

	// 6. TxIDの確認
	txid := txIDFromRawTxnBytesToSign(toBeSigned)
	fmt.Println("Transaction ID", txid)

	// ブロードキャストする前に送信内容を確認するため、ユーザーの入力を待つ
	fmt.Print("\nContinue? (y/n) >")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	in := scanner.Text()

	switch in {
	case "y":
		// 7. "y" を入力したら、署名済みトランザクションデータをブロードキャスト
		fmt.Println("Sending transaction...")
		pendingTxID, err := sendTransaction(stxBytes)
		if err != nil {
			log.Fatalf("Failed to send transaction: %s\n", err)
		}
		fmt.Printf("Pending Transaction ID: %s\n", pendingTxID)
	default:
		// y 以外の入力なら終了
		break
	}
}

func sendTransaction(stxBytes []byte) (string, error) {
	algodAddress := "https://testnet-api.algonode.cloud"
	algodToken := strings.Repeat("a", 64)
	algodClient, _ := algod.MakeClient(
		algodAddress,
		algodToken,
	)

	pendingTxID, err := algodClient.SendRawTransaction(stxBytes).Do(context.Background())
	if err != nil {
		return "", err
	}

	return pendingTxID, nil
}

func byteFromBase64(s string) []byte {
	b, _ := base64.StdEncoding.DecodeString(s)
	return b
}

func txIDFromRawTxnBytesToSign(toBeSigned []byte) string {
	txidBytes := sha512.Sum512_256(toBeSigned)
	return base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(txidBytes[:])
}
