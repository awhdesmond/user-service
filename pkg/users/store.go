package users

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/awhdesmond/user-service/pkg/common"
	"github.com/redis/go-redis/v9"
	"github.com/upper/db/v4"
	"go.uber.org/zap"
)

const (
	dbtable    = "users"
	loggerName = "users.store"
)

var (
	ErrUserNotFound            = errors.New("username not found")
	ErrUnexpectedDatabaseError = errors.New("unexpected error")

	DefaultCacheTTL = 10 * time.Minute
)

type Store interface {
	Upsert(ctx context.Context, username string, dob time.Time) error
	Read(ctx context.Context, username string) (User, error)
}

type store struct {
	sess   db.Session
	rdb    redis.UniversalClient
	logger *zap.Logger
}

func NewStore(sess db.Session, rdb redis.UniversalClient, logger *zap.Logger) Store {
	return &store{sess, rdb, logger.Named(loggerName)}
}

func (store *store) rdbUserKey(username string) string {
	return fmt.Sprintf("user_service:username:%s", username)
}

func (store *store) writeToCache(ctx context.Context, usr User) error {
	data, err := json.Marshal(usr)
	if err != nil {
		store.logger.Error("redis marshal error", zap.Error(err))
		return err
	}
	cmd := store.rdb.Set(ctx, store.rdbUserKey(usr.Username), data, DefaultCacheTTL)
	if cmd.Err() != nil {
		store.logger.Error("cache error", zap.Error(cmd.Err()))
		return ErrUnexpectedDatabaseError
	}
	return nil
}

// Upsert saves the username along with the date of a birth of a user. It also
// implements the write-through cache policy to save the information to redis.
func (store *store) Upsert(ctx context.Context, username string, dob time.Time) error {
	_, err := store.sess.WithContext(ctx).SQL().Query(`
		INSERT INTO users (username, date_of_birth)
		VALUES (?, ?)
		ON CONFLICT(username)
		DO UPDATE SET
			date_of_birth = EXCLUDED.date_of_birth
	`, username, dob)
	if err != nil {
		store.logger.Error("db error", zap.Error(err))
		return ErrUnexpectedDatabaseError
	}

	return store.writeToCache(ctx, User{Username: username, DoB: dob})
}

// Read retrieves the user from the cache (if it exists), else from the DB.
// It saves the information to the cache when the cache does not have it.
func (store *store) Read(ctx context.Context, username string) (User, error) {
	var usr User

	rdbUserKey := store.rdbUserKey(username)
	cmd := store.rdb.Get(ctx, rdbUserKey)
	if cmd.Err() != nil {
		// unexpected error
		if !errors.Is(cmd.Err(), redis.Nil) {
			store.logger.Error("cache error", zap.Error(cmd.Err()))
			return User{}, ErrUnexpectedDatabaseError
		}
		// else is key not found, so just fallthrough
	} else {
		// Key is found, return from cache
		if err := json.Unmarshal([]byte(cmd.Val()), &usr); err != nil {
			// dirty data in cache, refetch from db
			store.rdb.Del(ctx, rdbUserKey)
		} else {
			return usr, nil
		}
	}

	// Key is not found in cache, fetch from db
	q := store.sess.WithContext(ctx).SQL().SelectFrom(dbtable).Where("username = ?", username)
	err := q.One(&usr)

	if common.IsDBErrorNoRows(err) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		store.logger.Error("db error", zap.Error(err))
		return User{}, ErrUnexpectedDatabaseError
	}

	go func() {
		if err := store.writeToCache(context.Background(), usr); err != nil {
			store.logger.Warn("cache error", zap.Error(err))
		}
	}()
	return usr, nil
}
