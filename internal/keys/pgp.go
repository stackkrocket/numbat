package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"syscall"

	"github.com/stackkrocket/numbat/helpers"
	"golang.org/x/crypto/argon2"
	"golang.org/x/term"
)

// KeyPair represents a pair of RSA keys.

// LoadPrivateKey loads and decrypts the private key from a file.
func LoadPrivateKey(path string, passphrase []byte) (*rsa.PrivateKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %v", err)
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "ENCRYPTED PRIVATE KEY" {
		return nil, errors.New("invalid private key file")
	}

	decryptedDER, err := decryptPrivateKey(block.Bytes, passphrase)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %v", err)
	}

	key, err := x509.ParsePKCS8PrivateKey(decryptedDER)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	privateKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return nil, errors.New("not an RSA private key")
	}

	// Zero out sensitive data
	helpers.ZeroBytes(decryptedDER)

	return privateKey, nil
}

// LoadPublicKey loads the public key from a file.
func LoadPublicKey(path string) (*rsa.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %v", err)
	}

	block, _ := pem.Decode(data)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, errors.New("invalid public key file")
	}

	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	publicKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return nil, errors.New("not an RSA public key")
	}

	return publicKey, nil
}

// decryptPrivateKey decrypts an encrypted private key using the passphrase.
func decryptPrivateKey(encryptedKey, passphrase []byte) ([]byte, error) {
	if len(encryptedKey) < 16 {
		return nil, errors.New("invalid encrypted key")
	}

	salt := encryptedKey[:16]
	ciphertext := encryptedKey[16:]

	key := argon2.IDKey(passphrase, salt, 1, 64*1024, 4, 32)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %v", err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}

	nonceSize := aesgcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := aesgcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %v", err)
	}

	// Zero out sensitive data
	helpers.ZeroBytes(key)

	return plaintext, nil
}

// EncryptWithPublicKey encrypts data using the recipient's public key.
func EncryptWithPublicKey(publicKey *rsa.PublicKey, data []byte) ([]byte, error) {
	label := []byte("")
	hash := sha256.New()
	ciphertext, err := rsa.EncryptOAEP(hash, rand.Reader, publicKey, data, label)
	if err != nil {
		return nil, fmt.Errorf("encryption failed: %v", err)
	}
	return ciphertext, nil
}

// DecryptWithPrivateKey decrypts data using the owner's private key.
func DecryptWithPrivateKey(privateKey *rsa.PrivateKey, ciphertext []byte) ([]byte, error) {
	label := []byte("")
	hash := sha256.New()
	plaintext, err := rsa.DecryptOAEP(hash, rand.Reader, privateKey, ciphertext, label)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %v", err)
	}
	return plaintext, nil
}

// TestEncryption demonstrates the encryption and decryption process.
func TestEncryption() {
	// Load the recipient's public key
	publicKey, err := LoadPublicKey("../keys/admin/public_key.pem")
	if err != nil {
		fmt.Println("Error loading public key:", err)
		return
	}

	// Prompt user for the message
	fmt.Print("Enter your message: ")
	message, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		fmt.Println("Error reading message:", err)
		return
	}

	// Encrypt the message
	encryptedMessage, err := EncryptWithPublicKey(publicKey, message)
	if err != nil {
		fmt.Println("Error encrypting message:", err)
		return
	}
	fmt.Println("Encrypted message:", encryptedMessage)

	// Prompt for passphrase to load private key
	passphrase, err := helpers.PromptPassphrase(true)
	if err != nil {
		fmt.Println("Error reading passphrase:", err)
		return
	}

	// Load the owner's private key
	privateKey, err := LoadPrivateKey("../keys/admin/private_key.pem", passphrase)
	if err != nil {
		fmt.Println("Error loading private key:", err)
		return
	}

	// Decrypt the message
	decryptedMessage, err := DecryptWithPrivateKey(privateKey, encryptedMessage)
	if err != nil {
		fmt.Println("Error decrypting message:", err)
		return
	}
	fmt.Println("Decrypted message:", string(decryptedMessage))

	// Zero out sensitive data
	helpers.ZeroBytes(passphrase)
	helpers.ZeroBytes(message)
}
