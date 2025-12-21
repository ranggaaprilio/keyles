package services

// PasswordService defines the interface for password hashing and verification
type PasswordService interface {
	// Hash generates a bcrypt hash of the password with cost factor 12
	Hash(password string) (string, error)

	// Verify checks if the provided password matches the hash
	Verify(password, hash string) error
}
