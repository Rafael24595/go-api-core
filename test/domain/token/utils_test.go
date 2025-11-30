package token_test

import (
	"testing"

	"github.com/Rafael24595/go-api-core/src/domain/token"
)

func Test_HashToken(t *testing.T) {
	secret := "token:salt+pepper"

	for range 1000 {
		raw := token.GenerateRawToken()
		hash := token.HashToken(secret, raw)

		expected := token.HashToken(secret, raw)
		if hash != expected {
			t.Errorf("Found variable %q but %q expected", hash, expected)
		}
	}
}

func Test_HashToken_Static(t *testing.T) {
	secret := "token:salt+pepper"

	raw := "889a30a1c41da99be9aaa6c575dd5b22accad09262f77d07c0796c7fc5e8c19d"
	hash := token.HashToken(secret, raw)

	expected := token.HashToken(secret, raw)
	if hash != expected {
		t.Errorf("Found variable %q but %q expected", hash, expected)
	}
}

func Test_HashToken_Unique(t *testing.T) {
	secret := "token:salt+pepper"
	seen := make(map[string]bool)

	for range 10_000 {
		raw := token.GenerateRawToken()
		hash := token.HashToken(secret, raw)

		if seen[hash] {
			t.Fatalf("Hash collision found for token %s", raw)
		}

		seen[hash] = true
	}
}

func Test_HashToken_SecretInfluence(t *testing.T) {
	raw := token.GenerateRawToken()

	hash1 := token.HashToken("secret1", raw)
	hash2 := token.HashToken("secret2", raw)

	if hash1 == hash2 {
		t.Errorf("Hash should differ when using different secrets")
	}
}

func Test_HashToken_EmptyValues(t *testing.T) {
	_ = token.HashToken("", "")
	_ = token.HashToken("secret", "")
	_ = token.HashToken("", "rawtoken")
}

func Test_HashToken_Length(t *testing.T) {
	secret := "token:salt+pepper"
	raw := token.GenerateRawToken()
	hash := token.HashToken(secret, raw)

	if len(hash) != 64 {
		t.Errorf("Unexpected hash length: got %d, want 64", len(hash))
	}
}
