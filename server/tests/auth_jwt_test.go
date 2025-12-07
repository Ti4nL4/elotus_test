package tests

import (
	"testing"
	"time"

	"elotus_test/server/models/auth"
)

func TestNewJWTService_DefaultConfig(t *testing.T) {
	service := auth.NewJWTService(nil, nil)
	if service == nil {
		t.Fatal("Expected service to be created with default config")
	}
}

func TestNewJWTService_CustomConfig(t *testing.T) {
	config := &auth.Config{
		SecretKey:     []byte("custom-secret"),
		TokenDuration: 2 * time.Hour,
	}
	service := auth.NewJWTService(config, nil)

	if service == nil {
		t.Fatal("Expected service to be created")
	}
}

func TestGenerateToken_Success(t *testing.T) {
	config := &auth.Config{
		SecretKey:     []byte("test-secret-key"),
		TokenDuration: time.Hour,
	}
	service := auth.NewJWTService(config, nil)

	token, expiresAt, err := service.GenerateToken(123, "testuser")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Error("Expected non-empty token")
	}

	expectedExpiry := time.Now().Add(time.Hour)
	if expiresAt.Before(time.Now()) {
		t.Error("Token should expire in the future")
	}
	if expiresAt.After(expectedExpiry.Add(time.Minute)) {
		t.Error("Token expiry time is too far in the future")
	}
}

func TestValidateToken_Success(t *testing.T) {
	config := &auth.Config{
		SecretKey:     []byte("test-secret-key"),
		TokenDuration: time.Hour,
	}
	service := auth.NewJWTService(config, nil)

	// Generate a token
	token, _, err := service.GenerateToken(123, "testuser")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Validate it
	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.UserID != 123 {
		t.Errorf("Expected UserID 123, got %d", claims.UserID)
	}
	if claims.Username != "testuser" {
		t.Errorf("Expected Username 'testuser', got '%s'", claims.Username)
	}
	if claims.Issuer != "elotus-auth" {
		t.Errorf("Expected Issuer 'elotus-auth', got '%s'", claims.Issuer)
	}
}

func TestValidateToken_InvalidToken(t *testing.T) {
	config := &auth.Config{
		SecretKey:     []byte("test-secret-key"),
		TokenDuration: time.Hour,
	}
	service := auth.NewJWTService(config, nil)

	_, err := service.ValidateToken("invalid.token.string")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	config1 := &auth.Config{
		SecretKey:     []byte("secret-key-1"),
		TokenDuration: time.Hour,
	}
	service1 := auth.NewJWTService(config1, nil)

	config2 := &auth.Config{
		SecretKey:     []byte("secret-key-2"),
		TokenDuration: time.Hour,
	}
	service2 := auth.NewJWTService(config2, nil)

	// Generate with service1
	token, _, err := service1.GenerateToken(123, "testuser")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Try to validate with service2 (different secret)
	_, err = service2.ValidateToken(token)
	if err == nil {
		t.Error("Expected error when validating with wrong secret")
	}
}

func TestValidateToken_ExpiredToken(t *testing.T) {
	config := &auth.Config{
		SecretKey:     []byte("test-secret-key"),
		TokenDuration: -time.Hour, // Negative duration = already expired
	}
	service := auth.NewJWTService(config, nil)

	token, _, err := service.GenerateToken(123, "testuser")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	_, err = service.ValidateToken(token)
	if err == nil {
		t.Error("Expected error for expired token")
	}
}

func TestValidateToken_EmptyToken(t *testing.T) {
	config := &auth.Config{
		SecretKey:     []byte("test-secret-key"),
		TokenDuration: time.Hour,
	}
	service := auth.NewJWTService(config, nil)

	_, err := service.ValidateToken("")
	if err == nil {
		t.Error("Expected error for empty token")
	}
}

func TestTokenClaims_RegisteredClaims(t *testing.T) {
	config := &auth.Config{
		SecretKey:     []byte("test-secret-key"),
		TokenDuration: time.Hour,
	}
	service := auth.NewJWTService(config, nil)

	// JWT timestamps are truncated to seconds, so use second-precision for comparison
	beforeGen := time.Now().Truncate(time.Second)
	token, _, err := service.GenerateToken(123, "testuser")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}
	afterGen := time.Now().Add(time.Second).Truncate(time.Second)

	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	// Check IssuedAt (with 1 second tolerance for JWT truncation)
	issuedAt := claims.IssuedAt.Time
	if issuedAt.Before(beforeGen) || issuedAt.After(afterGen) {
		t.Errorf("IssuedAt should be between %v and %v, got %v", beforeGen, afterGen, issuedAt)
	}

	// Check NotBefore (with 1 second tolerance for JWT truncation)
	notBefore := claims.NotBefore.Time
	if notBefore.Before(beforeGen) || notBefore.After(afterGen) {
		t.Errorf("NotBefore should be between %v and %v, got %v", beforeGen, afterGen, notBefore)
	}

	// Check ExpiresAt
	expiresAt := claims.ExpiresAt.Time
	expectedExpiry := issuedAt.Add(time.Hour)
	if expiresAt.Before(expectedExpiry.Add(-time.Second)) || expiresAt.After(expectedExpiry.Add(time.Second)) {
		t.Errorf("ExpiresAt should be around %v, got %v", expectedExpiry, expiresAt)
	}

	// Check Subject
	if claims.Subject != "testuser" {
		t.Errorf("Expected Subject 'testuser', got '%s'", claims.Subject)
	}
}

func TestRevokeUserTokens_NoRevocationStore(t *testing.T) {
	config := &auth.Config{
		SecretKey:     []byte("test-secret-key"),
		TokenDuration: time.Hour,
	}
	service := auth.NewJWTService(config, nil)

	// Should not error when revocation store is nil
	err := service.RevokeUserTokens(123)
	if err != nil {
		t.Errorf("RevokeUserTokens should not error without revocation store: %v", err)
	}
}

func TestRevokeUserTokensBefore_NoRevocationStore(t *testing.T) {
	config := &auth.Config{
		SecretKey:     []byte("test-secret-key"),
		TokenDuration: time.Hour,
	}
	service := auth.NewJWTService(config, nil)

	// Should not error when revocation store is nil
	err := service.RevokeUserTokensBefore(123, time.Now())
	if err != nil {
		t.Errorf("RevokeUserTokensBefore should not error without revocation store: %v", err)
	}
}

func TestDefaultConfig(t *testing.T) {
	config := auth.DefaultConfig()

	if config == nil {
		t.Fatal("DefaultConfig should return a config")
	}

	if len(config.SecretKey) == 0 {
		t.Error("Default SecretKey should not be empty")
	}

	if config.TokenDuration != 24*time.Hour {
		t.Errorf("Expected default TokenDuration of 24h, got %v", config.TokenDuration)
	}
}

// Benchmark tests
func BenchmarkGenerateToken(b *testing.B) {
	config := &auth.Config{
		SecretKey:     []byte("test-secret-key-for-benchmarks"),
		TokenDuration: time.Hour,
	}
	service := auth.NewJWTService(config, nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = service.GenerateToken(123, "testuser")
	}
}

func BenchmarkValidateToken(b *testing.B) {
	config := &auth.Config{
		SecretKey:     []byte("test-secret-key-for-benchmarks"),
		TokenDuration: time.Hour,
	}
	service := auth.NewJWTService(config, nil)
	token, _, _ := service.GenerateToken(123, "testuser")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = service.ValidateToken(token)
	}
}

