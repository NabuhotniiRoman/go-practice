package migrations

import (
	"time"

	"gorm.io/gorm"
)

// Friendship модель для міграції
type Friendship struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    string    `gorm:"type:uuid;not null;index" json:"user_id"`
	FriendID  string    `gorm:"type:uuid;not null;index" json:"friend_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TableName явно задає ім'я таблиці для GORM
func (Friendship) TableName() string {
	return "friendships"
}

// CreateFriendshipsTable створює таблицю friendships
func CreateFriendshipsTable(tx *gorm.DB) error {
	return tx.AutoMigrate(&Friendship{})
}

// DropFriendshipsTable видаляє таблицю friendships
func DropFriendshipsTable(tx *gorm.DB) error {
	return tx.Migrator().DropTable("friendships")
}
