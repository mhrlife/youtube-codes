package repository

import (
	"TestSeparation/internal/entity"
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ory/dockertest/v3"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestGormUsersRepository_Create_And_Find(t *testing.T) {
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")

	pool.MaxWait = time.Minute * 2

	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	// uses pool to try to connect to Docker
	err = pool.Client.Ping()
	if err != nil {
		logrus.WithError(err).Fatal("Could not connect to Docker")
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("mysql", "5.7", []string{"MYSQL_ROOT_PASSWORD=secret"})
	if err != nil {
		logrus.WithError(err).Fatal("Could not start resource")
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		dsn := "root:secret@tcp(localhost:%s)/mysql?charset=utf8mb4&parseTime=True&loc=Local"
		dsn = fmt.Sprintf(dsn, resource.GetPort("3306/tcp"))
		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

		if err != nil {
			return err
		}

		return db.Exec("SELECT 1").Error

	}); err != nil {
		logrus.WithError(err).Fatal("Could not connect to database")
		t.Fatal(err)
	}

	dsn := "root:secret@tcp(localhost:%s)/mysql?charset=utf8mb4&parseTime=True&loc=Local"
	dsn = fmt.Sprintf(dsn, resource.GetPort("3306/tcp"))
	db, _ := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	// migrate user table
	assert.NoError(t, db.AutoMigrate(&entity.User{}))

	repository := NewGormUsersRepository(db)
	_, err = repository.ByField(context.Background(), "phone_number", "09211231231")
	assert.ErrorIs(t, err, ErrNotFound)

	user := &entity.User{
		DisplayName: "Mohammad",
		PhoneNumber: "09211231231",
	}

	assert.NoError(t, repository.Create(context.Background(), user))
	assert.Equal(t, user.ID, uint(1))

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		logrus.WithError(err).Error("Could not connect to database")
	}

}
