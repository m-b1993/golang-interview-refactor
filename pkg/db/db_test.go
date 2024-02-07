package db

import (
	"context"
	"interview/internal/config"
	"interview/internal/utils"
	"interview/pkg/log"
	"net/http"
	"net/http/httptest"
	"testing"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
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

func TestDB_Transactional(t *testing.T) {
	runDBTest(t, func(db *gorm.DB) {
		assert.Zero(t, successfulQueryCount(t, db))
		logger, _ := log.NewForTest()
		dbc := New(db, logger)

		// successful transaction
		err := dbc.Transactional(context.Background(), func(ctx context.Context) error {
			err := dbc.With(ctx).Exec("INSERT INTO dbcontexttest (id, name) VALUES(?, ?)", "1", "name1")
			assert.Nil(t, err.Error)
			err = dbc.With(ctx).Exec("INSERT INTO dbcontexttest (id, name) VALUES(?, ?)", "2", "name2")
			assert.Nil(t, err.Error)
			return nil
		})
		assert.Nil(t, err)
		assert.Equal(t, 2, successfulQueryCount(t, db))

		// failed transaction
		err = dbc.Transactional(context.Background(), func(ctx context.Context) error {
			err := dbc.With(ctx).Exec("INSERT INTO dbcontexttest (id, name) VALUES(?, ?)", "3", "name1")
			assert.Nil(t, err.Error)
			err = dbc.With(ctx).Exec("INSERT INTO dbcontexttest (id, name) VALUES(?, ?)", "4", "name2")
			assert.Nil(t, err.Error)
			return gorm.ErrRecordNotFound
		})
		assert.Equal(t, gorm.ErrRecordNotFound, err)
		assert.Equal(t, 2, successfulQueryCount(t, db))

		// failed transaction, but queries made outside of the transaction
		err = dbc.Transactional(context.Background(), func(ctx context.Context) error {
			err := dbc.With(context.Background()).Exec("INSERT INTO dbcontexttest (id, name) VALUES(?, ?)", "3", "name1")
			assert.Nil(t, err.Error)
			err = dbc.With(context.Background()).Exec("INSERT INTO dbcontexttest (id, name) VALUES(?, ?)", "4", "name2")
			assert.Nil(t, err.Error)
			return gorm.ErrRecordNotFound
		})
		assert.Equal(t, gorm.ErrRecordNotFound, err)
		assert.Equal(t, 4, successfulQueryCount(t, db))
	})
}

func TestDB_TransactionHandler(t *testing.T) {
	runDBTest(t, func(db *gorm.DB) {
		assert.Zero(t, successfulQueryCount(t, db))
		logger, _ := log.NewForTest()
		dbc := New(db, logger)
		txHandler := dbc.TransactionHandler()

		// successful transaction
		{
			w := httptest.NewRecorder()
			_, engine := gin.CreateTestContext(w)
			req, _ := http.NewRequest("GET", "/", nil)
			engine.Use(txHandler)
			engine.Use(func(c *gin.Context) {
				ctx := c.Request.Context()
				err := dbc.With(ctx).Exec("INSERT INTO dbcontexttest (id, name) VALUES(?, ?)", "1", "name1")
				assert.Nil(t, err.Error)
				err = dbc.With(ctx).Exec("INSERT INTO dbcontexttest (id, name) VALUES(?, ?)", "2", "name2")
				assert.Nil(t, err.Error)
				c.Next()
			})
			engine.ServeHTTP(w, req)

			assert.Equal(t, 2, successfulQueryCount(t, db))
		}

		// failed transaction
		{
			w := httptest.NewRecorder()
			_, engine := gin.CreateTestContext(w)
			req, _ := http.NewRequest("GET", "/", nil)
			engine.Use(txHandler)
			engine.Use(func(c *gin.Context) {
				ctx := c.Request.Context()
				err := dbc.With(ctx).Exec("INSERT INTO dbcontexttest (id, name) VALUES(?, ?)", "3", "name1")
				assert.Nil(t, err.Error)
				err = dbc.With(ctx).Exec("INSERT INTO dbcontexttest (id, name) VALUES(?, ?)", "4", "name2")
				assert.Nil(t, err.Error)
				c.Error(gorm.ErrRecordNotFound)
			})
			engine.ServeHTTP(w, req)

			assert.Equal(t, 2, successfulQueryCount(t, db))
		}
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

func successfulQueryCount(t *testing.T, db *gorm.DB) int {
	var count int
	err := db.Raw("SELECT COUNT(*) FROM dbcontexttest").Scan(&count)
	assert.Nil(t, err.Error)
	return count

}
