package repository

import (
	"fmt"
	"time"

	"github.com/LGU-SE-Internal/rcabench/database"
)

// BlacklistedToken represents a blacklisted JWT token
type BlacklistedToken struct {
	ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
	TokenID   string    `gorm:"unique;not null;index" json:"token_id"`   // JWT ID (jti claim)
	UserID    int       `gorm:"index" json:"user_id"`                    // User who owns the token
	ExpiresAt time.Time `gorm:"index" json:"expires_at"`                 // When the token expires
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`        // When it was blacklisted
	Reason    string    `gorm:"type:varchar(255)" json:"reason"`         // Reason for blacklisting
}

// CreateBlacklistedTokenTable creates the blacklisted tokens table if it doesn't exist
func CreateBlacklistedTokenTable() error {
	return database.DB.AutoMigrate(&BlacklistedToken{})
}

// BlacklistToken adds a token to the blacklist
func BlacklistToken(tokenID string, userID int, expiresAt time.Time, reason string) error {
	blacklistedToken := BlacklistedToken{
		TokenID:   tokenID,
		UserID:    userID,
		ExpiresAt: expiresAt,
		Reason:    reason,
	}
	
	if err := database.DB.Create(&blacklistedToken).Error; err != nil {
		return fmt.Errorf("failed to blacklist token: %v", err)
	}
	
	return nil
}

// IsTokenBlacklisted checks if a token is blacklisted
func IsTokenBlacklisted(tokenID string) (bool, error) {
	var count int64
	err := database.DB.Model(&BlacklistedToken{}).
		Where("token_id = ? AND expires_at > ?", tokenID, time.Now()).
		Count(&count).Error
	
	if err != nil {
		return false, fmt.Errorf("failed to check token blacklist status: %v", err)
	}
	
	return count > 0, nil
}

// BlacklistUserTokens blacklists all tokens for a specific user
func BlacklistUserTokens(userID int, reason string) error {
	// Note: This is a simplified approach. In a production system, 
	// you might want to track active tokens per user differently
	now := time.Now()
	futureExpiry := now.Add(24 * time.Hour) // Assume max token lifetime
	
	blacklistedToken := BlacklistedToken{
		TokenID:   fmt.Sprintf("user_%d_%d", userID, now.Unix()),
		UserID:    userID,
		ExpiresAt: futureExpiry,
		Reason:    reason,
	}
	
	if err := database.DB.Create(&blacklistedToken).Error; err != nil {
		return fmt.Errorf("failed to blacklist user tokens: %v", err)
	}
	
	return nil
}

// CleanupExpiredBlacklistedTokens removes expired blacklisted tokens
func CleanupExpiredBlacklistedTokens() error {
	result := database.DB.Where("expires_at < ?", time.Now()).Delete(&BlacklistedToken{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired blacklisted tokens: %v", result.Error)
	}
	
	return nil
}

// GetBlacklistedTokensByUser retrieves blacklisted tokens for a specific user
func GetBlacklistedTokensByUser(userID int, page, pageSize int) ([]BlacklistedToken, int64, error) {
	var tokens []BlacklistedToken
	var total int64
	
	query := database.DB.Model(&BlacklistedToken{}).Where("user_id = ?", userID)
	
	// Get total count
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count blacklisted tokens for user %d: %v", userID, err)
	}
	
	// Get paginated results
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&tokens).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to get blacklisted tokens for user %d: %v", userID, err)
	}
	
	return tokens, total, nil
}

// GetBlacklistedTokensCount returns the total count of blacklisted tokens
func GetBlacklistedTokensCount() (int64, error) {
	var count int64
	err := database.DB.Model(&BlacklistedToken{}).Count(&count).Error
	if err != nil {
		return 0, fmt.Errorf("failed to count blacklisted tokens: %v", err)
	}
	return count, nil
} 