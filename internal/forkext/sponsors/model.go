package sponsors

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

// Record is the local, operator-managed sponsor representation. Localized
// values remain JSON text so adding a locale never requires a schema change.
type Record struct {
	ID             string     `json:"id" gorm:"primaryKey;size:64"`
	Enabled        bool       `json:"enabled" gorm:"not null;index"`
	Name           string     `json:"name" gorm:"size:160;not null"`
	Priority       int        `json:"priority" gorm:"not null;default:0;index"`
	SlotsJSON      string     `json:"-" gorm:"column:slots_json;type:text;not null"`
	StartAt        *time.Time `json:"startAt,omitempty"`
	EndAt          *time.Time `json:"endAt,omitempty"`
	DestinationURL string     `json:"destinationUrl" gorm:"size:2048;not null"`
	LogoURL        string     `json:"logoUrl,omitempty" gorm:"size:2048"`
	TitleJSON      string     `json:"-" gorm:"column:title_json;type:text;not null"`
	TextJSON       string     `json:"-" gorm:"column:text_json;type:text;not null"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`
}

func (Record) TableName() string { return "fork_sponsors" }

func Models() []any { return []any{&Record{}} }

func Migrate(db *gorm.DB) error {
	if db == nil {
		return errors.New("sponsors migration requires database")
	}
	return db.AutoMigrate(Models()...)
}
