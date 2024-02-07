package db

import (
	"interview/internal/config"
	"interview/internal/utils"
	"interview/pkg/log"
	"testing"

	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	runDBTest(t, func(db *gorm.DB) {
		logger, _ := log.NewForTest()
		dbc := New(db, logger)
		assert.NotNil(t, dbc)
		assert.Equal(t, db, dbc.DB())
	})
}

func runDBTest(t *testing.T, f func(db *gorm.DB)) {
	logger, _ := log.NewForTest()
	cfg, err := config.Load("test.yml", logger)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	db, err := utils.GetDBConnection(cfg.DSN)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer func() {
		_ = utils.CloseDBConnection(db)
	}()

	sqls := []string{
		"CREATE TABLE IF NOT EXISTS dbcontexttest (id INT NOT NULL, name varchar(255), PRIMARY KEY (id))",
		"TRUNCATE dbcontexttest",
	}
	for _, s := range sqls {
		tx := db.Exec(s)
		assert.Nil(t, tx.Error)
	}

	f(db)
}
