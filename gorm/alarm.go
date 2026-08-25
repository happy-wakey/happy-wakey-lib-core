package happywakeygorm

import "time"

// Alarm is the persistence model. OwnerID is internal authorization state and
// must never be accepted from a public request body.
type Alarm struct {
	ID             string    `gorm:"column:id;type:uuid;primaryKey"`
	OwnerID        string    `gorm:"column:owner_id;not null;index:happy_wakey_alarms_owner_enabled,priority:1"`
	Label          string    `gorm:"column:label;not null"`
	LocalTime      string    `gorm:"column:local_time;not null"`
	TimeZone       string    `gorm:"column:time_zone;not null"`
	WeekdaysJSON   []byte    `gorm:"column:weekdays;type:jsonb;not null"`
	Enabled        bool      `gorm:"column:enabled;not null;default:true;index:happy_wakey_alarms_owner_enabled,priority:2"`
	Sound          string    `gorm:"column:sound;not null"`
	Volume         float64   `gorm:"column:volume;not null"`
	GradualSeconds uint32    `gorm:"column:gradual_seconds;not null"`
	TagsJSON       []byte    `gorm:"column:tags;type:jsonb;not null"`
	Generation     uint64    `gorm:"column:generation;not null"`
	CreatedAt      time.Time `gorm:"column:created_at;not null"`
	UpdatedAt      time.Time `gorm:"column:updated_at;not null;index:happy_wakey_alarms_owner_enabled,priority:3,sort:desc"`
}

func (Alarm) TableName() string { return "happy_wakey_alarms" }
