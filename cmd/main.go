package main

import (
	internal "github.com/stackkrocket/numbat/internal/keys"
)

func main() {
	// bits := 2048

	// keyPair, err := internal.GenerateKeyPair(bits)
	// if err != nil {
	// 	fmt.Println("Error generating key pair:", err)
	// 	return
	// }

	// // Prompt for passphrase with confirmation
	// passphrase, err := helpers.PromptPassphrase(true)
	// if err != nil {
	// 	fmt.Println("Error reading passphrase:", err)
	// 	return
	// }

	// // Save keys to files
	// err = keyPair.SaveKeys("../keys/admin/private_key.pem", "../keys/admin/public_key.pem", passphrase)
	// if err != nil {
	// 	fmt.Println("Error saving keys:", err)
	// 	return
	// }

	// fmt.Println("Keys generated and saved successfully.")

	// //Zero out passphrase
	// helpers.ZeroBytes(passphrase)

	internal.TestEncryption()

}
