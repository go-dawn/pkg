package deck

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Fake struct {
	gorm.Model
	F string
}

func Test_Deck_GDB_SetupGdb(t *testing.T) {
	gdb := SetupGormDB(t, &Fake{})

	assert.True(t, gdb.Migrator().HasTable(&Fake{}))
}

func Test_Deck_GDB_DryRunSession(t *testing.T) {
	s := DryRunSession(t)

	stat := s.Find(&Fake{}).Statement

	assert.Equal(t, "SELECT * FROM `fakes` WHERE `fakes`.`deleted_at` IS NULL", stat.SQL.String())
}

func Test_Deck_GDB_AssertDBCount(t *testing.T) {
	gdb := SetupGormDB(t, &Fake{})

	AssertDBCount(t, gdb.Model(&Fake{}), int64(0))
}

func Test_Deck_GDB_AssertDBHas(t *testing.T) {
	gdb := SetupGormDB(t, &Fake{})

	assert.Nil(t, gdb.Create(&Fake{F: "f"}).Error)

	AssertDBHas(t, gdb.Model(&Fake{}), Columns{"F": "f"})
}

func Test_Deck_GDB_AssertDBMissing(t *testing.T) {
	gdb := SetupGormDB(t, &Fake{})

	AssertDBMissing(t, gdb.Model(&Fake{}), Columns{"F": "f"})
}

func Test_Deck_GDB_DisabledGormLogger(t *testing.T) {
	var (
		l   DisabledGormLogger
		ctx = context.Background()
	)

	l.LogMode(logger.Info)

	l.Info(ctx, "info")
	l.Warn(ctx, "warn")
	l.Error(ctx, "error")
	l.Trace(ctx, time.Now(), func() (string, int64) { return "", 0 }, nil)
}
