package main

import (
	//"github.com/algorand/go-algorand-sdk/v2/client/v2/algod"
	"context"
	"fmt"
	"log"

	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/transaction"
)

func main() {
	FromAccount := crypto.GenerateAccount()
	fmt.Println("Algorand FromAddress", FromAccount.Address)

	ToAccount := crypto.GenerateAccount()
	fmt.Println("Algorand TOAddress", ToAccount.Address)

	// example: TRANSACTION_PAYMENT_CREATE
	sp, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		log.Fatalf("failed to get suggested params: %s", err)
	}

	// payment from account to itself
	ptxn, err := transaction.MakePaymentTxn(
		FromAccount.Address.String(), // From
		ToAccount.Address.String(),   // To
		sp.Fee,                       // Fee (MicroAlgos)
		100000,                       // Amount (MicroAlgos)
		sp.FirstRoundValid,           // FirstRound
		sp.LastRoundValid,            // LastRound
		nil,                          // Note
		"",                           // CloseRemainderTo
		"",                           // GenesisID
		sp.GenesisHash)
	if err != nil {
		log.Fatalf("failed creating transaction: %s", err)
	}

	// example: TRANSACTION_PAYMENT_SIGN
	_, sptxn, err := crypto.SignTransaction(acct.PrivateKey, ptxn)
	if err != nil {
		fmt.Printf("Failed to sign transaction: %s\n", err)
		return
	}
	// example: TRANSACTION_PAYMENT_SIGN
}
