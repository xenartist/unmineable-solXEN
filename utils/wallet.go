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
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gagliardetto/solana-go"
	"github.com/rivo/tview"
)

// var GLOBAL_PASSWORD string = ""
// var GLOBAL_PUBLIC_KEY string = ""
// var GLOBAL_PRIVATE_KEY string = ""

var (
	encryptedPassword  []byte
	encryptedPublicKey []byte
	// encryptedPrivateKey []byte
	encryptionKey []byte
	mutex         sync.Mutex
)

func PasswordProtectionInit() {
	// Generate a random encryption key
	encryptionKey = make([]byte, 32)
	_, err := rand.Read(encryptionKey)
	if err != nil {
		panic("Failed to generate encryption key")
	}
}

func xorEncrypt(data []byte, key []byte) []byte {
	encrypted := make([]byte, len(data))
	for i := 0; i < len(data); i++ {
		encrypted[i] = data[i] ^ key[i%len(key)]
	}
	return encrypted
}

func SetGlobalPassword(password string) {
	mutex.Lock()
	defer mutex.Unlock()
	encryptedPassword = xorEncrypt([]byte(password), encryptionKey)
}

func GetGlobalPassword() string {
	mutex.Lock()
	defer mutex.Unlock()
	if encryptedPassword == nil {
		return ""
	}
	return string(xorEncrypt(encryptedPassword, encryptionKey))
}

func SetGlobalPublicKey(publicKey string) {
	mutex.Lock()
	defer mutex.Unlock()
	encryptedPublicKey = xorEncrypt([]byte(publicKey), encryptionKey)
}

func GetGlobalPublicKey() string {
	mutex.Lock()
	defer mutex.Unlock()
	if encryptedPublicKey == nil {
		return ""
	}
	return string(xorEncrypt(encryptedPublicKey, encryptionKey))
}

// func SetGlobalPrivateKey(privateKey string) {
// 	mutex.Lock()
// 	defer mutex.Unlock()
// 	encryptedPrivateKey = xorEncrypt([]byte(privateKey), encryptionKey)
// }

// func GetGlobalPrivateKey() string {
// 	mutex.Lock()
// 	defer mutex.Unlock()
// 	if encryptedPrivateKey == nil {
// 		return ""
// 	}
// 	return string(xorEncrypt(encryptedPrivateKey, encryptionKey))
// }

func ClearGlobalKeys() {
	mutex.Lock()
	defer mutex.Unlock()
	encryptedPassword = nil
	encryptedPublicKey = nil
	// encryptedPrivateKey = nil
}

func CheckExistingWallet() string {
	walletDir := filepath.Join(GetExecutablePath(), "wallet")

	// Check if wallet directory exists
	if _, err := os.Stat(walletDir); os.IsNotExist(err) {
		return ""
	}

	// Read wallet directory
	files, err := os.ReadDir(walletDir)
	if err != nil {
		return ""
	}

	// Look for .solXENwallet file
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".solXENwallet") {
			return strings.TrimSuffix(file.Name(), ".solXENwallet")
		}
	}

	return ""
}

func CreateNewWallet(app *tview.Application, logView *tview.TextView, logMessage LogMessageFunc, password string) (string, error) {
	// Check for existing wallet
	files, err := os.ReadDir(filepath.Join(GetExecutablePath(), "wallet"))
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
	err = os.MkdirAll(filepath.Join(GetExecutablePath(), "wallet"), 0700)
	if err != nil {
		logMessage(logView, "Error creating wallet directory: "+err.Error())
		return "", err
	}
	fileName := fmt.Sprintf("%s.solXENwallet", publicKey[:8])
	err = os.WriteFile(filepath.Join(GetExecutablePath(), "wallet", fileName), ciphertext, 0600)
	if err != nil {
		logMessage(logView, "Error saving encrypted data: "+err.Error())
		return "", err
	}

	// Set global variables
	SetGlobalPassword(password)
	// SetGlobalPrivateKey(base64.StdEncoding.EncodeToString(privateKey))
	SetGlobalPublicKey(publicKey)

	logMessage(logView, "Wallet created successfully: "+publicKey[:8]+"********")
	return publicKey, nil
}

func VerifyPassword(password string) bool {
	LogToFile("Starting password verification for public key")

	// Check for existing wallet
	files, err := os.ReadDir(filepath.Join(GetExecutablePath(), "wallet"))
	if err != nil && !os.IsNotExist(err) {
		LogToFile("Error reading wallet directory: " + err.Error())
		return false
	}

	walletFilePath := ""
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".solXENwallet") {
			walletFilePath = filepath.Join("wallet", file.Name())
		}
	}

	// Read the wallet file
	encryptedData, err := os.ReadFile(walletFilePath)
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

	// Decode the private key
	privateKeyBytes, err := base64.StdEncoding.DecodeString(walletData.PrivateKey)
	if err != nil {
		LogToFile("Private key decoding failed: " + err.Error())
		return false
	}

	// Generate the public key from the private key
	account := solana.PrivateKey(privateKeyBytes)
	generatedPublicKey := account.PublicKey().String()

	// Verify public key and store private key if successful
	if generatedPublicKey == walletData.PublicKey {
		SetGlobalPassword(password)
		// SetGlobalPrivateKey(walletData.PrivateKey)
		SetGlobalPublicKey(walletData.PublicKey)
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

func ExportPrivateKey() error {
	LogToFile("Starting private key export")

	// Check if global password is set
	password := GetGlobalPassword()
	if password == "" {
		LogToFile("Global password is not set")
		return errors.New("global password is not set")
	}

	// Find the wallet file
	walletDir := filepath.Join(GetExecutablePath(), "wallet")
	files, err := os.ReadDir(walletDir)
	if err != nil {
		LogToFile("Error reading wallet directory: " + err.Error())
		return err
	}

	var walletFile string
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".solXENwallet") {
			walletFile = filepath.Join(walletDir, file.Name())
			break
		}
	}

	if walletFile == "" {
		LogToFile("Wallet file not found")
		return errors.New("wallet file not found")
	}

	// Read and decrypt the wallet file
	encryptedData, err := os.ReadFile(walletFile)
	if err != nil {
		LogToFile("Error reading wallet file: " + err.Error())
		return err
	}

	decryptedData, err := decrypt(encryptedData, []byte(password))
	if err != nil {
		LogToFile("Error decrypting wallet file: " + err.Error())
		return err
	}

	var walletData struct {
		PublicKey  string `json:"public_key"`
		PrivateKey string `json:"private_key"`
	}
	err = json.Unmarshal(decryptedData, &walletData)
	if err != nil {
		LogToFile("Error unmarshalling wallet data: " + err.Error())
		return err
	}

	// Prepare data for export
	exportData := map[string]string{
		"public_key":  walletData.PublicKey,
		"private_key": walletData.PrivateKey,
	}
	exportJSON, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		LogToFile("Error marshalling export data: " + err.Error())
		return err
	}

	// Save export data to file
	exportFilePath := filepath.Join(walletDir, "solXEN-private-key-exported.json")
	err = os.WriteFile(exportFilePath, exportJSON, 0600)
	if err != nil {
		LogToFile("Error saving export data: " + err.Error())
		return err
	}

	LogToFile("Wallet exported successfully to " + exportFilePath)
	return nil
}

func ExportPublicKey() error {
	LogToFile("Starting public key export")

	// Check if global public key is set
	if GetGlobalPublicKey() == "" {
		LogToFile("Global public key is not set")
		return errors.New("global public key is not set")
	}

	// Prepare data for export
	exportData := map[string]string{
		"public_key": GetGlobalPublicKey(),
	}
	exportJSON, err := json.MarshalIndent(exportData, "", "  ")
	if err != nil {
		LogToFile("Error marshalling export data: " + err.Error())
		return err
	}

	// Save export data to file
	exportFilePath := filepath.Join("wallet", "solXEN-public-key-exported.json")
	err = os.WriteFile(exportFilePath, exportJSON, 0600)
	if err != nil {
		LogToFile("Error saving export data: " + err.Error())
		return err
	}

	LogToFile("Public key exported successfully to " + exportFilePath)
	return nil
}
