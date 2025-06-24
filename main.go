package main

import (
	"context"
	"crypto/ecdsa"
	"encoding/csv"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/joho/godotenv"
)

type stats struct {
	SentAt          time.Time
	TxnHash         string
	IncludedInBlock uint64
	InclusionDelay  time.Duration
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	region := os.Getenv("REGION")
	if region == "" {
		log.Fatal("REGION environment variable not set")
	}

	key := os.Getenv("PRIVATE_KEY")
	if key == "" {
		log.Fatal("PRIVATE_KEY environment variable not set")
	}

	toAddressRaw := os.Getenv("TO_ADDRESS")
	if toAddressRaw == "" {
		log.Fatal("TO_ADDRESS environment variable not set")
	}

	toAddress := common.HexToAddress(toAddressRaw)
	if toAddress == (common.Address{}) {
		log.Fatal("TO_ADDRESS environment variable not set")
	}

	flashblocksUrl := os.Getenv("FLASHBLOCKS_URL")
	if flashblocksUrl == "" {
		log.Fatal("FLASHBLOCKS_URL environment variable not set")
	}

	baseUrl := os.Getenv("BASE_URL")
	if baseUrl == "" {
		log.Fatal("BASE_URL environment variable not set")
	}

	sendTxnSync := os.Getenv("SEND_TXN_SYNC") == "true"

	pollingIntervalMs := 100
	if pollingEnv := os.Getenv("POLLING_INTERVAL_MS"); pollingEnv != "" {
		if parsed, err := strconv.Atoi(pollingEnv); err == nil {
			pollingIntervalMs = parsed
		}
	}

	numberOfTransactions := 100
	if txnCountEnv := os.Getenv("NUMBER_OF_TRANSACTIONS"); txnCountEnv != "" {
		if parsed, err := strconv.Atoi(txnCountEnv); err == nil {
			numberOfTransactions = parsed
		}
	}

	flashblocksClient, err := ethclient.Dial(flashblocksUrl)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	baseClient, err := ethclient.Dial(baseUrl)
	if err != nil {
		log.Fatalf("Failed to connect to the Ethereum client: %v", err)
	}

	privateKey, err := crypto.HexToECDSA(key)
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}

	publicKey := privateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		log.Fatal("Failed to cast public key to ECDSA")
	}
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)

	var flashblockTimings []stats
	var baseTimings []stats

	chainId, err := baseClient.NetworkID(context.Background())
	if err != nil {
		log.Fatalf("Failed to get network ID: %v", err)
	}

	flashblockErrors := 0
	baseErrors := 0

	log.Printf("Starting flashblock transactions")
	for i := 0; i < numberOfTransactions; i++ {
		timing, err := timeTransaction(chainId, privateKey, fromAddress, toAddress, flashblocksClient, sendTxnSync, pollingIntervalMs)
		if err != nil {
			flashblockErrors += 1
			log.Printf("Failed to send transaction: %v", err)
		}

		flashblockTimings = append(flashblockTimings, timing)

		// wait for it to be mined -- sleep a random amount between 600ms and 1s
		time.Sleep(time.Duration(rand.Int63n(600)+600) * time.Millisecond)
	}

	// wait for the final fb transaction to land
	time.Sleep(5 * time.Second)

	log.Printf("Starting regular transactions")
	for i := 0; i < numberOfTransactions; i++ {
		// Currently not supported on non-flashblock endpoints
		timing, err := timeTransaction(chainId, privateKey, fromAddress, toAddress, baseClient, false, pollingIntervalMs)
		if err != nil {
			baseErrors += 1
			log.Printf("Failed to send transaction: %v", err)
		}

		baseTimings = append(baseTimings, timing)

		// wait for it to be mined -- sleep a random amount between 4s and 3s
		time.Sleep(time.Duration(rand.Int63n(1000)+4000) * time.Millisecond)
	}

	if err := writeToFile(fmt.Sprintf("./data/flashblocks-%s.csv", region), flashblockTimings); err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}

	if err := writeToFile(fmt.Sprintf("./data/base-%s.csv", region), baseTimings); err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}

	log.Printf("Completed test with %d transactions", numberOfTransactions)
	log.Printf("Flashblock errors: %v", flashblockErrors)
	log.Printf("BaseErrors: %v", baseErrors)
}

func writeToFile(filename string, data []stats) error {
	file, err := os.Create(filename)
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	header := []string{"sent_at", "txn_hash", "included_in_block", "inclusion_delay_ms"}
	if err := writer.Write(header); err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}

	for _, d := range data {
		row := []string{
			d.SentAt.String(),
			d.TxnHash,
			strconv.FormatUint(d.IncludedInBlock, 10),
			strconv.FormatInt(d.InclusionDelay.Milliseconds(), 10),
		}
		if err := writer.Write(row); err != nil {
			log.Fatalf("Failed to write to file: %v", err)
		}
	}

	return nil
}

func timeTransaction(chainId *big.Int, privateKey *ecdsa.PrivateKey, fromAddress common.Address, toAddress common.Address, client *ethclient.Client, useSyncRPC bool, pollingIntervalMs int) (stats, error) {
	nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
	if err != nil {
		return stats{}, fmt.Errorf("unable to get nonce: %v", err)
	}

	gasPrice, err := client.SuggestGasPrice(context.Background())
	if err != nil {
		return stats{}, fmt.Errorf("unable to get gas price: %v", err)
	}
	gasLimit := uint64(21000)
	value := big.NewInt(100)

	tip, err := client.SuggestGasTipCap(context.Background())
	if err != nil {
		return stats{}, fmt.Errorf("unable to get gas price: %v", err)
	}

	tx := types.NewTx(&types.DynamicFeeTx{
		ChainID:   chainId,
		Nonce:     nonce,
		GasTipCap: tip,
		GasFeeCap: gasPrice,
		Gas:       gasLimit,
		To:        &toAddress,
		Value:     value,
		Data:      nil,
	})

	signedTx, err := types.SignTx(tx, types.NewPragueSigner(chainId), privateKey)
	if err != nil {
		return stats{}, fmt.Errorf("unable to sign transaction: %v", err)
	}

	sentAt := time.Now()

	if useSyncRPC {
		return sendTransactionSync(client, signedTx, sentAt)
	}

	return sendTransactionAsync(client, signedTx, sentAt, pollingIntervalMs)
}

func sendTransactionSync(client *ethclient.Client, signedTx *types.Transaction, sentAt time.Time) (stats, error) {
	rawTx, err := signedTx.MarshalBinary()
	if err != nil {
		return stats{}, fmt.Errorf("unable to marshal transaction: %v", err)
	}

	txnData := "0x" + hex.EncodeToString(rawTx)

	var receipt *types.Receipt
	err = client.Client().CallContext(context.Background(), &receipt, "eth_sendRawTransactionSync", txnData)
	if err != nil {
		return stats{}, fmt.Errorf("unable to send sync transaction: %v", err)
	}

	log.Println("Transaction sent sync: ", signedTx.Hash().Hex())
	now := time.Now()
	return stats{
		SentAt:          sentAt,
		InclusionDelay:  now.Sub(sentAt),
		TxnHash:         signedTx.Hash().Hex(),
		IncludedInBlock: receipt.BlockNumber.Uint64(),
	}, nil
}

func sendTransactionAsync(client *ethclient.Client, signedTx *types.Transaction, sentAt time.Time, pollingIntervalMs int) (stats, error) {
	err := client.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return stats{}, fmt.Errorf("unable to send transaction: %v", err)
	}

	log.Println("Transaction sent: ", signedTx.Hash().Hex())

	for i := 0; i < 1000; i++ {
		receipt, err := client.TransactionReceipt(context.Background(), signedTx.Hash())
		if err != nil {
			time.Sleep(time.Duration(pollingIntervalMs) * time.Millisecond)
		} else {
			now := time.Now()
			return stats{
				SentAt:          sentAt,
				InclusionDelay:  now.Sub(sentAt),
				TxnHash:         signedTx.Hash().Hex(),
				IncludedInBlock: receipt.BlockNumber.Uint64(),
			}, nil
		}
	}

	return stats{}, fmt.Errorf("failed to get transaction")
}
