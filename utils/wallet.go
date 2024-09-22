package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/gagliardetto/solana-go"
	"github.com/rivo/tview"
)

var PRIVATE_KEY string

func CreateNewWallet(app *tview.Application, logView *tview.TextView, logMessage LogMessageFunc, password string) (string, error) {
	// Check for existing wallet
	files, err := os.ReadDir("wallet")
	if err != nil && !os.IsNotExist(err) {
		logMessage(logView, "Error reading wallet directory: "+err.Error())
		return "", err
	}

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".solXENwallet") {
			publicKey := strings.TrimSuffix(file.Name(), ".solXENwallet")
			logMessage(logView, "Wallet already exists with public key: "+publicKey)
			return publicKey, nil
		}
	}

	// Generate Solana keypair
	logMessage(logView, "Generating new wallet...")
	account := solana.NewWallet()
	publicKey := account.PublicKey().String()
	privateKey := account.PrivateKey

	// Prepare data for encryption
	logMessage(logView, "Preparing data for encryption...")
	data := map[string]string{
		"password":    password,
		"public_key":  publicKey,
		"private_key": base64.StdEncoding.EncodeToString(privateKey),
	}
	plaintext, err := json.Marshal(data)
	if err != nil {
		logMessage(logView, "Error marshalling data: "+err.Error())
		return "", err
	}

	// Encrypt data
	logMessage(logView, "Encrypting data...")
	ciphertext, err := encrypt(plaintext, []byte(password))
	if err != nil {
		logMessage(logView, "Error encrypting data: "+err.Error())
		return "", err
	}

	// Save encrypted data to file
	logMessage(logView, "Saving encrypted data to file...")
	err = os.MkdirAll("wallet", 0700)
	if err != nil {
		logMessage(logView, "Error creating wallet directory: "+err.Error())
		return "", err
	}
	fileName := fmt.Sprintf("%s.solXENwallet", publicKey)
	err = os.WriteFile(filepath.Join("wallet", fileName), ciphertext, 0600)
	if err != nil {
		logMessage(logView, "Error saving encrypted data: "+err.Error())
		return "", err
	}

	logMessage(logView, "Wallet created successfully: "+publicKey)
	return publicKey, nil
}

func VerifyPassword(publicKey string, password string) bool {
	LogToFile("Starting password verification for public key: " + publicKey)

	// Construct the wallet file name
	walletFileName := publicKey + ".solXENwallet"
	walletFilePath := filepath.Join("wallet", walletFileName)
	LogToFile("Wallet file path: " + walletFilePath)

	// Read the wallet file
	encryptedData, err := ioutil.ReadFile(walletFilePath)
	if err != nil {
		LogToFile("Error reading wallet file: " + err.Error())
		return false
	}
	LogToFile("Wallet file read successfully")

	// Attempt to decrypt the wallet file
	decryptedData, err := decrypt(encryptedData, []byte(password))
	if err != nil {
		LogToFile("Decryption failed: " + err.Error())
		return false
	}
	LogToFile("Wallet decrypted successfully")

	// Parse the decrypted JSON data
	var walletData struct {
		PublicKey  string `json:"public_key"`
		PrivateKey string `json:"private_key"`
	}
	err = json.Unmarshal(decryptedData, &walletData)
	if err != nil {
		LogToFile("JSON parsing failed: " + err.Error())
		return false
	}
	LogToFile("JSON parsed successfully")

	// Verify public key and store private key if successful
	if walletData.PublicKey == publicKey {
		PRIVATE_KEY = walletData.PrivateKey
		LogToFile("Public key verified successfully")
		return true
	}

	LogToFile("Public key verification failed")
	return false
}

func encrypt(plaintext, password []byte) ([]byte, error) {
	key := sha256.Sum256(password)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, aes.BlockSize+len(plaintext))
	iv := ciphertext[:aes.BlockSize]
	if _, err := rand.Read(iv); err != nil {
		return nil, err
	}

	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)

	return ciphertext, nil
}

func decrypt(ciphertext []byte, password []byte) ([]byte, error) {
	key := sha256.Sum256(password)
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}

	if len(ciphertext) < aes.BlockSize {
		return nil, errors.New("ciphertext too short")
	}
	iv := ciphertext[:aes.BlockSize]
	ciphertext = ciphertext[aes.BlockSize:]

	plaintext := make([]byte, len(ciphertext))
	stream := cipher.NewCTR(block, iv)
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext, nil
}
