package migrations

import "gorm.io/gorm"

// DummyMigration is a no-op migration for testing purposes.
func DummyMigration(tx *gorm.DB) error {
	// This migration does nothing.
	return nil
}

// RollbackDummyMigration is a no-op rollback for testing purposes.
func RollbackDummyMigration(tx *gorm.DB) error {
	// This rollback does nothing.
	return nil
}
