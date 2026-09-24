package analytics

import "gorm.io/gorm"

// Migrate adds only fork-owned analytics tables. It never alters or deletes
// upstream tables, so disabling analytics leaves the upstream schema untouched.
// This is intentionally the only schema integration point for the package.
func Migrate(db *gorm.DB) error {
	if db == nil {
		return nil
	}
	return db.AutoMigrate(&DestinationObservation{}, &DNSObservation{}, &NetworkSession{}, &ServiceCategoryAggregate{}, &AccessLogCursor{})
}
