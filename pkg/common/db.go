package common

import (
	"crypto/tls"
	"fmt"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/upper/db/v4"
	postgresqladp "github.com/upper/db/v4/adapter/postgresql"
)

type PostgresSQLConfig struct {
	Host     string `mapstructure:"postgres-host"`
	Port     string `mapstructure:"postgres-port"`
	Database string `mapstructure:"postgres-database"`
	Username string `mapstructure:"postgres-username"`
	Password string `mapstructure:"postgres-password"`
}

func (c PostgresSQLConfig) Hostname() string {
	if c.Port == "" {
		return c.Host
	}
	return fmt.Sprintf("%s:%v", c.Host, c.Port)
}

func MakePostgresDBSession(cfg PostgresSQLConfig) (db.Session, error) {
	settings := postgresqladp.ConnectionURL{
		User:     cfg.Username,
		Password: cfg.Password,
		Host:     cfg.Hostname(),
		Database: cfg.Database,
	}

	session, err := postgresqladp.Open(settings)
	if err != nil {
		return nil, err
	}

	db.LC().SetLevel(db.LogLevelError)
	session.SetMaxIdleConns(db.DefaultSettings.MaxIdleConns())
	session.SetMaxOpenConns(db.DefaultSettings.MaxOpenConns())
	session.SetConnMaxLifetime(db.DefaultSettings.ConnMaxLifetime())
	session.SetConnMaxIdleTime(db.DefaultSettings.ConnMaxIdleTime())

	return session, nil
}

func IsDBErrorNoRows(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "no more rows in this result set")
}

// Redis
type RedisCfg struct {
	URI         string `mapstructure:"redis-uri"`
	Password    string `mapstructure:"redis-password"`
	ClusterMode string `mapstructure:"redis-cluster-mode"`
}

func redisOptsToUnivOpts(opts *redis.Options, password string) *redis.UniversalOptions {
	return &redis.UniversalOptions{
		Addrs:    []string{opts.Addr},
		Password: password,
		DB:       opts.DB,
	}
}

// makeRedisClusterClient creates a Redis Cluster client.
// Need to use cluster client for AWS Elasticache with a single configuration URL
// See: https://stackoverflow.com/questions/73907312/i-want-to-connect-to-elasticcache-for-redis-in-which-cluster-mode-is-enabled-i
func makeRedisClusterClient(opts *redis.Options, password string) *redis.ClusterClient {
	return redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:    []string{opts.Addr},
		Password: password,

		TLSConfig: &tls.Config{
			InsecureSkipVerify: true,
		},

		PoolSize:     10,
		MinIdleConns: 10,

		ReadOnly:       false,
		RouteRandomly:  false,
		RouteByLatency: false,
	})
}

// MakeRedisClient creates a redis client for connection to redis cluster.
// It supports creating cluster client if RedisCfg.ClusterMode is non-empty string,
// else it creates a universal client which determines the underlying mode using
// the number of addresses provided.
func MakeRedisClient(cfg RedisCfg) (redis.UniversalClient, error) {
	opts, err := redis.ParseURL(cfg.URI)
	if err != nil {
		return nil, err
	}

	if cfg.ClusterMode != "" {
		// Use cluster mode, make cluster client
		return makeRedisClusterClient(opts, cfg.Password), nil
	}

	// Else, make universal client
	rdb := redis.NewUniversalClient(redisOptsToUnivOpts(opts, cfg.Password))
	return rdb, nil
}
