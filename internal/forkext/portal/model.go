package portal

import "time"

type Credential struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ClientID  int       `json:"clientId" gorm:"uniqueIndex;not null"`
	TokenHash string    `json:"-" gorm:"size:64;not null"`
	Version   int64     `json:"-" gorm:"not null;default:1"`
	Enabled   bool      `json:"enabled" gorm:"not null;default:true"`
	ExpiresAt int64     `json:"expiresAt" gorm:"not null;default:0"`
	LastUsed  int64     `json:"lastUsed" gorm:"not null;default:0"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (Credential) TableName() string { return "fork_portal_credentials" }

type HostGrant struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	SubjectType string    `json:"subjectType" gorm:"size:16;not null;uniqueIndex:idx_portal_host_grant,priority:1"`
	SubjectID   int       `json:"subjectId" gorm:"not null;uniqueIndex:idx_portal_host_grant,priority:2"`
	HostID      int       `json:"hostId" gorm:"not null;uniqueIndex:idx_portal_host_grant,priority:3;index"`
	Enabled     bool      `json:"enabled" gorm:"not null;default:true"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func (HostGrant) TableName() string { return "fork_portal_host_grants" }

type PortalSettings struct {
	Enabled bool `json:"enabled"`
}

type CredentialView struct {
	ID        uint  `json:"id"`
	ClientID  int   `json:"clientId"`
	Enabled   bool  `json:"enabled"`
	ExpiresAt int64 `json:"expiresAt"`
	LastUsed  int64 `json:"lastUsed"`
	CreatedAt int64 `json:"createdAt"`
}

type IssuedToken struct {
	CredentialView
	Token string `json:"token"`
}
