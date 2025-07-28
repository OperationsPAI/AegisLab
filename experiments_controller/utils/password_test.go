package utils

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		shouldError bool
		description string
	}{
		{
			name:        "Valid Password",
			password:    "password123",
			shouldError: false,
			description: "Should hash a valid password",
		},
		{
			name:        "Short Password",
			password:    "123",
			shouldError: true,
			description: "Should fail with password too short",
		},
		{
			name:        "Empty Password",
			password:    "",
			shouldError: true,
			description: "Should fail with empty password",
		},
		{
			name:        "Very Long Password",
			password:    strings.Repeat("a", 150),
			shouldError: true,
			description: "Should fail with password too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)

			if tt.shouldError {
				assert.Error(t, err, tt.description)
				assert.Empty(t, hash)
			} else {
				assert.NoError(t, err, tt.description)
				assert.NotEmpty(t, hash)
				assert.Contains(t, hash, ":", "Hash should contain salt separator")

				// Verify we can verify the password
				isValid := VerifyPassword(tt.password, hash)
				assert.True(t, isValid, "Should be able to verify the password")
			}
		})
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "testpassword123"
	hash, err := HashPassword(password)
	assert.NoError(t, err)

	tests := []struct {
		name        string
		password    string
		hash        string
		expected    bool
		description string
	}{
		{
			name:        "Correct Password",
			password:    password,
			hash:        hash,
			expected:    true,
			description: "Should verify correct password",
		},
		{
			name:        "Wrong Password",
			password:    "wrongpassword",
			hash:        hash,
			expected:    false,
			description: "Should fail with wrong password",
		},
		{
			name:        "Empty Password",
			password:    "",
			hash:        hash,
			expected:    false,
			description: "Should fail with empty password",
		},
		{
			name:        "Invalid Hash Format",
			password:    password,
			hash:        "invalid-hash",
			expected:    false,
			description: "Should fail with invalid hash format",
		},
		{
			name:        "Empty Hash",
			password:    password,
			hash:        "",
			expected:    false,
			description: "Should fail with empty hash",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifyPassword(tt.password, tt.hash)
			assert.Equal(t, tt.expected, result, tt.description)
		})
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name             string
		password         string
		expectedStrength PasswordStrength
		shouldError      bool
		description      string
	}{
		{
			name:             "Very Strong Password",
			password:         "MyStr0ngP@ssw0rd!",
			expectedStrength: VeryStrongPassword,
			shouldError:      false,
			description:      "Should be very strong with all character types",
		},
		{
			name:             "Strong Password",
			password:         "StrongPass123!",
			expectedStrength: VeryStrongPassword, // 6分：长度8+, 长度12+, 小写, 大写, 数字, 特殊字符
			shouldError:      false,
			description:      "Should be very strong with all character types",
		},
		{
			name:             "Moderate Password",
			password:         "Password123",
			expectedStrength: ModeratePassword,
			shouldError:      false,
			description:      "Should be moderate with basic complexity",
		},
		{
			name:             "Weak Password",
			password:         "password",
			expectedStrength: WeakPassword,
			shouldError:      false,
			description:      "Should be weak with no complexity",
		},
		{
			name:             "Too Short",
			password:         "123",
			expectedStrength: WeakPassword,
			shouldError:      true,
			description:      "Should fail with password too short",
		},
		{
			name:             "Common Pattern",
			password:         "password123",
			expectedStrength: ModeratePassword, // 3分-1分=2分：长度8+, 小写, 数字, 减去常见模式penalty
			shouldError:      false,
			description:      "Should be moderate despite common pattern penalty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strength, suggestions, err := ValidatePasswordStrength(tt.password)

			if tt.shouldError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
				assert.Equal(t, tt.expectedStrength, strength, tt.description)
				assert.NotEmpty(t, suggestions, "Should provide suggestions")
			}
		})
	}
}

func TestPasswordStrengthString(t *testing.T) {
	tests := []struct {
		strength PasswordStrength
		expected string
	}{
		{WeakPassword, "Weak"},
		{ModeratePassword, "Moderate"},
		{StrongPassword, "Strong"},
		{VeryStrongPassword, "Very Strong"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.strength.String()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGenerateRandomPassword(t *testing.T) {
	tests := []struct {
		name        string
		length      int
		description string
	}{
		{
			name:        "Minimum Length",
			length:      8,
			description: "Should generate password with minimum length",
		},
		{
			name:        "Medium Length",
			length:      16,
			description: "Should generate password with medium length",
		},
		{
			name:        "Long Password",
			length:      32,
			description: "Should generate long password",
		},
		{
			name:        "Too Short (Auto-corrected)",
			length:      4,
			description: "Should auto-correct to minimum length",
		},
		{
			name:        "Too Long (Auto-corrected)",
			length:      200,
			description: "Should auto-correct to maximum length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			password, err := GenerateRandomPassword(tt.length)

			assert.NoError(t, err, tt.description)
			assert.NotEmpty(t, password, "Password should not be empty")

			// Check minimum length constraint
			expectedLength := tt.length
			if expectedLength < MinPasswordLength {
				expectedLength = MinPasswordLength
			}
			if expectedLength > MaxPasswordLength {
				expectedLength = MaxPasswordLength
			}

			assert.Equal(t, expectedLength, len(password), "Password length should match expected")

			// Verify password contains different character types
			hasLower := false
			hasUpper := false
			hasDigit := false
			hasSpecial := false

			for _, char := range password {
				if char >= 'a' && char <= 'z' {
					hasLower = true
				} else if char >= 'A' && char <= 'Z' {
					hasUpper = true
				} else if char >= '0' && char <= '9' {
					hasDigit = true
				} else {
					hasSpecial = true
				}
			}

			assert.True(t, hasLower, "Should contain lowercase letters")
			assert.True(t, hasUpper, "Should contain uppercase letters")
			assert.True(t, hasDigit, "Should contain digits")
			assert.True(t, hasSpecial, "Should contain special characters")
		})
	}
}

func TestIsPasswordValid(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		shouldError bool
		description string
	}{
		{
			name:        "Valid Password",
			password:    "password123",
			shouldError: false,
			description: "Should be valid with letters and numbers",
		},
		{
			name:        "Too Short",
			password:    "pass1",
			shouldError: true,
			description: "Should fail when too short",
		},
		{
			name:        "No Numbers",
			password:    "passwordonly",
			shouldError: true,
			description: "Should fail without numbers",
		},
		{
			name:        "No Letters",
			password:    "12345678",
			shouldError: true,
			description: "Should fail without letters",
		},
		{
			name:        "Too Long",
			password:    strings.Repeat("password123", 15),
			shouldError: true,
			description: "Should fail when too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := IsPasswordValid(tt.password)

			if tt.shouldError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

func TestConstantTimeCompare(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected bool
	}{
		{
			name:     "Equal Strings",
			a:        "hello",
			b:        "hello",
			expected: true,
		},
		{
			name:     "Different Strings",
			a:        "hello",
			b:        "world",
			expected: false,
		},
		{
			name:     "Different Lengths",
			a:        "hello",
			b:        "hi",
			expected: false,
		},
		{
			name:     "Empty Strings",
			a:        "",
			b:        "",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := constantTimeCompare(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests
func BenchmarkHashPassword(b *testing.B) {
	password := "testpassword123"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = HashPassword(password)
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "testpassword123"
	hash, _ := HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VerifyPassword(password, hash)
	}
}

func BenchmarkGenerateRandomPassword(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = GenerateRandomPassword(16)
	}
}
