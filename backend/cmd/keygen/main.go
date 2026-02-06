package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

func main() {
	// Parse command-line flags
	keySize := flag.Int("size", 2048, "RSA key size in bits (2048 or 4096)")
	outputDir := flag.String("output", "./keys", "Output directory for keys")
	keyID := flag.String("kid", "dev_key_001", "Key ID for JWT signing")
	flag.Parse()

	// Validate key size
	if *keySize != 2048 && *keySize != 4096 {
		fmt.Println("Error: Key size must be 2048 or 4096 bits")
		os.Exit(1)
	}

	// Create output directory if it doesn't exist
if err := os.MkdirAll(*outputDir, 0755); err != nil {
fmt.Printf("Error creating output directory: %v\n", err)
os.Exit(1)
}

// Generate RSA private key
fmt.Printf("Generating %d-bit RSA key pair...\n", *keySize)
privateKey, err := rsa.GenerateKey(rand.Reader, *keySize)
if err != nil {
fmt.Printf("Error generating private key: %v\n", err)
os.Exit(1)
}

// Save private key
privateKeyPath := filepath.Join(*outputDir, "private.pem")
privateKeyFile, err := os.Create(privateKeyPath)
if err != nil {
fmt.Printf("Error creating private key file: %v\n", err)
os.Exit(1)
}
defer privateKeyFile.Close()

privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
privateKeyPEM := &pem.Block{
Type:  "RSA PRIVATE KEY",
Bytes: privateKeyBytes,
}

if err := pem.Encode(privateKeyFile, privateKeyPEM); err != nil {
fmt.Printf("Error encoding private key: %v\n", err)
os.Exit(1)
}

// Set restrictive permissions on private key
if err := os.Chmod(privateKeyPath, 0600); err != nil {
fmt.Printf("Warning: Could not set permissions on private key: %v\n", err)
}
fmt.Printf("✓ Private key saved to: %s\n", privateKeyPath)

// Save public key
publicKeyPath := filepath.Join(*outputDir, "public.pem")
publicKeyFile, err := os.Create(publicKeyPath)
if err != nil {
fmt.Printf("Error creating public key file: %v\n", err)
os.Exit(1)
}
defer publicKeyFile.Close()

publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
if err != nil {
fmt.Printf("Error marshaling public key: %v\n", err)
os.Exit(1)
}

publicKeyPEM := &pem.Block{
Type:  "RSA PUBLIC KEY",
Bytes: publicKeyBytes,
}

if err := pem.Encode(publicKeyFile, publicKeyPEM); err != nil {
fmt.Printf("Error encoding public key: %v\n", err)
os.Exit(1)
}
fmt.Printf("✓ Public key saved to: %s\n", publicKeyPath)

// Display key information
fmt.Println("\n=== Key Information ===")
fmt.Printf("Key ID: %s\n", *keyID)
fmt.Printf("Algorithm: RS256\n")
fmt.Printf("Key Size: %d bits\n", *keySize)

fmt.Println("\nUpdate your .env file with:")
fmt.Printf("JWT_KEY_ID=%s\n", *keyID)
fmt.Printf("JWT_SIGNING_KEY_PATH=%s\n", privateKeyPath)
fmt.Printf("JWT_PUBLIC_KEY_PATH=%s\n", publicKeyPath)

fmt.Println("\n⚠️  WARNING: Keep the private key secure and never commit it to version control!")
fmt.Printf("Add %s to your .gitignore file.\n", *outputDir)
}
