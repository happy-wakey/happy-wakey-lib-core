package happywakeygorm

import (
	"context"
	"errors"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAlarmsForSubjectIsOwnerScoped(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&Alarm{}); err != nil {
		t.Fatal(err)
	}
	now := time.Date(2026, 8, 25, 12, 0, 0, 0, time.UTC)
	rows := []Alarm{
		{ID: "a", OwnerID: "subject:a", Label: "A", LocalTime: "07:00", TimeZone: "UTC", WeekdaysJSON: []byte("[]"), Enabled: true, Sound: "bell", TagsJSON: []byte("[]"), CreatedAt: now, UpdatedAt: now},
		{ID: "b", OwnerID: "subject:b", Label: "B", LocalTime: "08:00", TimeZone: "UTC", WeekdaysJSON: []byte("[]"), Enabled: true, Sound: "bell", TagsJSON: []byte("[]"), CreatedAt: now, UpdatedAt: now},
	}
	if err := db.Create(&rows).Error; err != nil {
		t.Fatal(err)
	}
	repository, err := NewReadRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	alarms, err := repository.AlarmsForSubject(context.Background(), "subject:a")
	if err != nil {
		t.Fatal(err)
	}
	if len(alarms) != 1 || alarms[0].OwnerID != "subject:a" {
		t.Fatalf("owner boundary failed: %#v", alarms)
	}
}

func TestSubjectValidationFailsClosed(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	repository, err := NewReadRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	for _, subject := range []string{"", "has spaces", string(make([]byte, 513))} {
		if _, err := repository.AlarmsForSubject(context.Background(), subject); !errors.Is(err, ErrInvalidSubject) {
			t.Fatalf("expected invalid subject for %q, got %v", subject, err)
		}
	}
}
