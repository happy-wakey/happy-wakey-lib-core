package happywakeygorm

import (
	"context"
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"gorm.io/gorm"
)

var ErrInvalidSubject = errors.New("verified subject must be non-empty, bounded, and whitespace-free")

type ReadRepository struct {
	db *gorm.DB
}

func NewReadRepository(db *gorm.DB) (*ReadRepository, error) {
	if db == nil {
		return nil, errors.New("database is required")
	}
	return &ReadRepository{db: db}, nil
}

func (repository *ReadRepository) AlarmsForSubject(ctx context.Context, subject string) ([]Alarm, error) {
	if err := validateSubject(subject); err != nil {
		return nil, err
	}
	var alarms []Alarm
	result := repository.db.WithContext(ctx).
		Where("owner_id = ?", subject).
		Order("created_at ASC").
		Find(&alarms)
	return alarms, result.Error
}

func validateSubject(subject string) error {
	if subject == "" || len(subject) > 512 || !utf8.ValidString(subject) || strings.IndexFunc(subject, func(r rune) bool { return r == '\uFFFD' || unicode.IsSpace(r) }) >= 0 {
		return ErrInvalidSubject
	}
	return nil
}
