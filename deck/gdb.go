package deck

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Columns is a condition struct for query
type Columns map[string]interface{}

// SetupGormDB gets gorm.DB instance. Passed in models will be auto migrated.
func SetupGormDB(t *testing.T, dst ...interface{}) *gorm.DB {
	db, err := gorm.Open(
		sqlite.Open(":memory:"),
		&gorm.Config{Logger: DisabledGormLogger{}})

	assert.Nil(t, err)

	if len(dst) > 0 {
		assert.Nil(t, db.AutoMigrate(dst...))
	}

	return db
}

// DryRunSession gets a gorm session in dry run mode.
func DryRunSession(t *testing.T) *gorm.DB {
	return SetupGormDB(t).Session(&gorm.Session{DryRun: true, Logger: DisabledGormLogger{}})
}

// AssertDBCount asserts db has specific count.
func AssertDBCount(t *testing.T, gdb *gorm.DB, expected int64) {
	var actual int64
	gdb.Count(&actual)

	assert.Equal(t, expected, actual)
}

// AssertDBHas asserts db has data with specific query condition.
func AssertDBHas(t *testing.T, gdb *gorm.DB, cols Columns) {
	var count int64
	gdb.Where(map[string]interface{}(cols)).Count(&count)

	assert.Truef(t, count > 0, "data not found by %+v", cols)
}

// AssertDBMissing asserts db misses data with specific query condition.
func AssertDBMissing(t *testing.T, gdb *gorm.DB, cols Columns) {
	var count int64
	gdb.Where(map[string]interface{}(cols)).Count(&count)

	assert.Truef(t, count == 0, "data found by %+v", cols)
}

// DisabledGormLogger implements gorm logger interface
type DisabledGormLogger struct{}

// LogMode is log mode
func (DisabledGormLogger) LogMode(logger.LogLevel) logger.Interface {
	return DisabledGormLogger{}
}

// Info print info messages
func (DisabledGormLogger) Info(context.Context, string, ...interface{}) {}

// Warn print warn messages
func (DisabledGormLogger) Warn(context.Context, string, ...interface{}) {}

// Error print error messages
func (DisabledGormLogger) Error(context.Context, string, ...interface{}) {}

// Trace print sql message
func (DisabledGormLogger) Trace(context.Context, time.Time, func() (string, int64), error) {}
