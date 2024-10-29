package utils

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gagliardetto/solana-go"
	spl_token "github.com/gagliardetto/solana-go/programs/token"
	"github.com/gagliardetto/solana-go/rpc"
)

// BurnToken burns a specified amount of a given token
func BurnToken(amount string, token string) (string, error) {
	// Initialize Solana client
	client := rpc.New("https://api.mainnet-beta.solana.com")

	// Define the token mint address and the associated token account
	// Get token mint address from tokenAddresses map
	mintAddress, exists := tokenAddresses[token]
	if !exists {
		LogToFile(fmt.Sprintf("Error: token %s not found in supported tokens", token))
		return "", fmt.Errorf("token %s not found in supported tokens", token)
	}
	tokenMintAddress := solana.MustPublicKeyFromBase58(mintAddress)

	// Get owner's public key
	owner, err := solana.PrivateKeyFromBase58(getPrivateKey())
	if err != nil {

		return "", fmt.Errorf("invalid private key: %v", err)
	}

	// Find associated token account
	tokenAccountAddress, _, err := solana.FindAssociatedTokenAddress(
		owner.PublicKey(),
		tokenMintAddress,
	)
	if err != nil {
		return "", fmt.Errorf("failed to find associated token account: %v", err)
	}

	// Check token balance before burning
	// tokenBalance, err := client.GetTokenAccountBalance(
	// 	context.TODO(),
	// 	tokenAccountAddress,
	// 	rpc.CommitmentFinalized,
	// )
	// if err != nil {
	// 	LogToFile(fmt.Sprintf("Error: failed to get token balance: %v", err))
	// 	return "", fmt.Errorf("failed to get token balance: %v", err)
	// }

	// Convert amount to uint64
	amountFloat, err := strconv.ParseFloat(amount, 64)
	if err != nil {
		LogToFile(fmt.Sprintf("Error: invalid amount: %v", err))
		return "", fmt.Errorf("invalid amount: %v", err)
	}

	// Convert to smallest unit (multiply by 10^6)
	amountToBurn := uint64(amountFloat * 1_000_000)

	// LogToFile("7777")
	// // Convert balance to uint64 for comparison
	// currentBalance, err := strconv.ParseFloat(tokenBalance.Value.Amount, 64)
	// if err != nil {
	// 	LogToFile(fmt.Sprintf("Error: failed to parse current balance: %v", err))
	// 	return "", fmt.Errorf("failed to parse current balance: %v", err)
	// }

	// LogToFile("8888")
	// // Check if balance is sufficient
	// if uint64(currentBalance) < amountToBurn {
	// 	LogToFile(fmt.Sprintf("Error: insufficient balance. Required: %d, Available: %d", amountToBurn, uint64(currentBalance)))
	// 	return "", fmt.Errorf("insufficient token balance")
	// }

	// Fetch a recent blockhash
	var recentBlockhash *rpc.GetLatestBlockhashResult
	maxRetries := 2
	for i := 0; i < maxRetries; i++ {
		var err error
		recentBlockhash, err = client.GetLatestBlockhash(context.TODO(), rpc.CommitmentFinalized)
		if err != nil {
			LogToFile(fmt.Sprintf("Attempt %d: Error getting blockhash: %v", i+1, err))
			if i == maxRetries-1 {
				return "", fmt.Errorf("failed to get recent blockhash after %d attempts: %v", maxRetries, err)
			}
			time.Sleep(time.Second * 2)
			continue
		}
		break
	}

	// Create the burn instruction
	burnInstruction := spl_token.NewBurnCheckedInstruction(
		amountToBurn, // amount
		6,            // decimals
		tokenAccountAddress,
		tokenMintAddress,
		owner.PublicKey(),
		[]solana.PublicKey{}, // multisigSigners (empty if not using multisig)
	)

	// Create and send the transaction
	var tx *solana.Transaction
	for i := 0; i < maxRetries; i++ {
		var err error
		tx, err = solana.NewTransaction(
			[]solana.Instruction{burnInstruction.Build()},
			recentBlockhash.Value.Blockhash,
			solana.TransactionPayer(owner.PublicKey()),
		)
		if err != nil {
			LogToFile(fmt.Sprintf("Attempt %d: Error creating transaction: %v", i+1, err))
			if i == maxRetries-1 {
				return "", fmt.Errorf("failed to create transaction after %d attempts: %v", maxRetries, err)
			}
			time.Sleep(time.Second * 2)
			continue
		}
		break
	}

	// Sign the transaction
	_, err = tx.Sign(
		func(key solana.PublicKey) *solana.PrivateKey {
			if key.Equals(owner.PublicKey()) {
				return &owner
			}
			return nil
		},
	)
	if err != nil {
		LogToFile(fmt.Sprintf("Error: failed to sign transaction: %v", err))
		return "", fmt.Errorf("failed to sign transaction: %v", err)
	}

	// Send the transaction
	var sig solana.Signature
	for i := 0; i < maxRetries; i++ {
		var err error
		var maxRetriesUint uint = uint(maxRetries)
		sig, err = client.SendTransactionWithOpts(context.TODO(), tx, rpc.TransactionOpts{
			SkipPreflight:       false,
			PreflightCommitment: rpc.CommitmentFinalized,
			MaxRetries:          &maxRetriesUint,
		})
		if err != nil {
			LogToFile(fmt.Sprintf("Attempt %d: Error sending transaction: %v", i+1, err))
			if i == maxRetries-1 {
				return "", fmt.Errorf("failed to send transaction after %d attempts: %v", maxRetries, err)
			}
			// Get new blockhash and recreate transaction if needed
			recentBlockhash, err = client.GetLatestBlockhash(context.TODO(), rpc.CommitmentFinalized)
			if err == nil {
				tx.Message.RecentBlockhash = recentBlockhash.Value.Blockhash
			}
			time.Sleep(time.Second * 2)
			continue
		}
		break
	}

	LogToFile(fmt.Sprintf("Burned %s of %s, transaction signature: %s", amount, token, sig))
	return fmt.Sprintf("Burned %s of %s", amount, token), nil
}
