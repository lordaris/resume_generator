package security

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "SecureP@ssw0rd"
	params := DefaultArgon2Params()

	// Test with default parameters
	hash, err := HashPassword(password, params)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Check hash format
	if !strings.HasPrefix(hash, "$argon2id$v=") {
		t.Errorf("Invalid hash format: %s", hash)
	}

	// Hash should be different each time due to random salt
	hash2, err := HashPassword(password, params)
	if err != nil {
		t.Fatalf("Second HashPassword failed: %v", err)
	}
	if hash == hash2 {
		t.Errorf("Hashes should be different due to random salt")
	}

	// Test with nil parameters (should use defaults)
	hash3, err := HashPassword(password, nil)
	if err != nil {
		t.Fatalf("HashPassword with nil params failed: %v", err)
	}
	if !strings.HasPrefix(hash3, "$argon2id$v=") {
		t.Errorf("Invalid hash format with nil params: %s", hash3)
	}

	// Test with custom parameters
	customParams := &Argon2Params{
		Memory:      32 * 1024,
		Iterations:  2,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}
	hash4, err := HashPassword(password, customParams)
	if err != nil {
		t.Fatalf("HashPassword with custom params failed: %v", err)
	}
	if !strings.Contains(hash4, "m=32768,t=2,p=2") {
		t.Errorf("Custom parameters not reflected in hash: %s", hash4)
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "SecureP@ssw0rd"
	wrongPassword := "WrongPassword"

	// Hash the password
	hash, err := HashPassword(password, nil)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Verify with correct password
	match, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if !match {
		t.Errorf("VerifyPassword should return true for correct password")
	}

	// Verify with wrong password
	match, err = VerifyPassword(wrongPassword, hash)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}
	if match {
		t.Errorf("VerifyPassword should return false for wrong password")
	}

	// Test with invalid hash format
	_, err = VerifyPassword(password, "invalid-hash-format")
	if err == nil {
		t.Errorf("VerifyPassword should fail with invalid hash format")
	}
}

func TestDecodeHash(t *testing.T) {
	// Create a valid hash
	password := "TestPassword"
	hash, err := HashPassword(password, nil)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Decode the hash
	params, salt, hashBytes, err := decodeHash(hash)
	if err != nil {
		t.Fatalf("decodeHash failed: %v", err)
	}

	// Check decoded parameters
	if params.Memory != DefaultArgon2Params().Memory {
		t.Errorf("Wrong memory parameter: got %d, want %d", params.Memory, DefaultArgon2Params().Memory)
	}
	if params.Iterations != DefaultArgon2Params().Iterations {
		t.Errorf("Wrong iterations parameter: got %d, want %d", params.Iterations, DefaultArgon2Params().Iterations)
	}
	if params.Parallelism != DefaultArgon2Params().Parallelism {
		t.Errorf("Wrong parallelism parameter: got %d, want %d", params.Parallelism, DefaultArgon2Params().Parallelism)
	}
	if len(salt) != int(DefaultArgon2Params().SaltLength) {
		t.Errorf("Wrong salt length: got %d, want %d", len(salt), DefaultArgon2Params().SaltLength)
	}
	if len(hashBytes) != int(DefaultArgon2Params().KeyLength) {
		t.Errorf("Wrong hash bytes length: got %d, want %d", len(hashBytes), DefaultArgon2Params().KeyLength)
	}

	// Test with invalid hash format
	_, _, _, err = decodeHash("invalid-hash-format")
	if err == nil {
		t.Errorf("decodeHash should fail with invalid hash format")
	}

	// Test with invalid version
	_, _, _, err = decodeHash("$argon2id$v=99$m=65536,t=3,p=4$c29tZXNhbHQ$hash")
	if err == nil {
		t.Errorf("decodeHash should fail with invalid version")
	}
}
