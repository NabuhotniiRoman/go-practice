package migrations

import (
	"gorm.io/gorm"
)

// AddFriendshipsUniqueConstraint додає унікальний індекс на (user_id, friend_id)
func AddFriendshipsUniqueConstraint(tx *gorm.DB) error {
	// Додаємо унікальний індекс для запобігання дублюванню дружніх зв'язків
	return tx.Exec(`
		CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS idx_friendships_user_friend 
		ON friendships (user_id, friend_id)
	`).Error
}

// DropFriendshipsUniqueConstraint видаляє унікальний індекс
func DropFriendshipsUniqueConstraint(tx *gorm.DB) error {
	return tx.Exec(`DROP INDEX IF EXISTS idx_friendships_user_friend`).Error
}
