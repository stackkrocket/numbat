package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	"github.com/stackkrocket/numbat/helpers"
	"golang.org/x/crypto/argon2"
)

// KeyPair represents a pair of RSA keys.
type KeyPair struct {
	PrivateKey *rsa.PrivateKey
	PublicKey  *rsa.PublicKey
	Bits       int
}

// GenerateKeyPair generates a new RSA key pair with the specified key size.
func GenerateKeyPair(bits int) (*KeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, fmt.Errorf("failed to generate private key: %v", err)
	}
	return &KeyPair{
		PrivateKey: privateKey,
		PublicKey:  &privateKey.PublicKey,
		Bits:       bits,
	}, nil
}

// SaveKeys saves the private and public keys to files, encrypting the private key.
func (kp *KeyPair) SaveKeys(privatePath, publicPath string, passphrase []byte) error {
	// Encrypt the private key
	encryptedPrivateKey, err := encryptPrivateKey(kp.PrivateKey, passphrase)
	if err != nil {
		return fmt.Errorf("failed to encrypt private key: %v", err)
	}

	// Save the encrypted private key
	if err := savePEMFile(privatePath, "ENCRYPTED PRIVATE KEY", encryptedPrivateKey); err != nil {
		return fmt.Errorf("failed to save private key: %v", err)
	}

	// Marshal the public key
	publicKeyDER, err := x509.MarshalPKIXPublicKey(kp.PublicKey)
	if err != nil {
		return fmt.Errorf("failed to marshal public key: %v", err)
	}

	// Save the public key
	if err := savePEMFile(publicPath, "PUBLIC KEY", publicKeyDER); err != nil {
		return fmt.Errorf("failed to save public key: %v", err)
	}

	return nil
}

// encryptPrivateKey encrypts a private key using a passphrase.
func encryptPrivateKey(privateKey *rsa.PrivateKey, passphrase []byte) ([]byte, error) {
	// Marshal the private key
	privateKeyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal private key: %v", err)
	}

	// Generate a random salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %v", err)
	}

	// Derive a key using Argon2id
	key := argon2.IDKey(passphrase, salt, 1, 64*1024, 4, 32)

	// Encrypt the private key using AES-GCM
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher block: %v", err)
	}
	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %v", err)
	}
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("failed to generate nonce: %v", err)
	}
	ciphertext := aesgcm.Seal(nonce, nonce, privateKeyDER, nil)

	// Prepend the salt to the ciphertext
	encryptedPrivateKey := append(salt, ciphertext...)

	// Zero out sensitive data
	helpers.ZeroBytes(privateKeyDER)
	helpers.ZeroBytes(key)

	return encryptedPrivateKey, nil
}

// savePEMFile saves a PEM block to a file with restricted permissions.
func savePEMFile(path, blockType string, derBytes []byte) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	pemBlock := &pem.Block{
		Type:  blockType,
		Bytes: derBytes,
	}

	if err := pem.Encode(file, pemBlock); err != nil {
		return fmt.Errorf("failed to encode PEM block: %v", err)
	}

	return nil
}
