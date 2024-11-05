package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/awhdesmond/user-service/pkg/common"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

const (
	testPort       = "18080"
	testMetricPort = "19090"
)

var (
	envPort       = ""
	envMetricPort = ""
)

type testSuite struct {
	suite.Suite
}

func (ts *testSuite) SetupSuite() {
	envPort = os.Getenv("USERS_SVC_PORT")
	envMetricPort = os.Getenv("USERS_SVC_METRICS_PORT")

	os.Setenv("USERS_SVC_PORT", testPort)
	os.Setenv("USERS_SVC_METRICS_PORT", testMetricPort)
	os.Setenv("USERS_SVC_POSTGRES_HOST", "localhost")
	os.Setenv("USERS_SVC_POSTGRES_PORT", "5432")
	os.Setenv("USERS_SVC_POSTGRES_USERNAME", "postgres")
	os.Setenv("USERS_SVC_POSTGRES_PASSWORD", "postgres")
	os.Setenv("USERS_SVC_POSTGRES_DATABASE", common.GetPostgresTestDb())
	os.Setenv("USERS_SVC_REDIS_URI", "redis://localhost:6379/10")

	go func() {
		main()
	}()
	time.Sleep(2 * time.Second)
}

func (ts *testSuite) TearDownSuite() {
	var srvCfg ServerConfig
	if err := viper.Unmarshal(&srvCfg); err != nil {
		ts.T().Fatalf("got = %v, want = %v", err, nil)
	}

	pgSess, err := common.MakePostgresDBSession(srvCfg.PostgresSQLConfig)
	if err != nil {
		ts.T().Fatalf("got = %v, want = %v", err, nil)
	}
	rdb, err := common.MakeRedisClient(srvCfg.RedisCfg)
	if err != nil {
		ts.T().Fatalf("got = %v, want = %v", err, nil)
	}

	_, err = pgSess.SQL().Exec(common.TruncateAllTablesSQL)
	if err != nil {
		ts.T().Log(err)
	}
	rdb.FlushDB(context.TODO())

	// Set back env vars
	os.Setenv("USERS_SVC_PORT", envPort)
	os.Setenv("USERS_SVC_METRICS_PORT", envMetricPort)
}

type ApiTestSuite struct {
	testSuite
}

func TestApiTestSuite(t *testing.T) {
	suite.Run(t, new(ApiTestSuite))
}

func (ts *ApiTestSuite) Test() {
	ts.testUpsertAndRead()
	ts.testHealth()
}

func (ts *ApiTestSuite) testUpsertAndRead() {
	client := http.Client{}

	// Upsert

	jsonBody := []byte(`{"dateOfBirth": "2020-04-01"}`)
	bodyReader := bytes.NewReader(jsonBody)
	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("http://localhost:%s/hello/apple", testPort),
		bodyReader,
	)
	if err != nil {
		ts.T().Fatalf("got %v, want = %v", err, nil)
	}

	res, err := client.Do(req)
	if err != nil {
		ts.T().Fatalf("got %v, want = %v", err, nil)
	}
	if res.StatusCode != http.StatusNoContent {
		ts.T().Fatalf("got %v, want = %v", res.StatusCode, http.StatusNoContent)
	}

	// Read

	req, err = http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("http://localhost:%s/hello/apple", testPort),
		nil,
	)
	if err != nil {
		ts.T().Fatalf("got %v, want = %v", err, nil)
	}

	res, err = client.Do(req)
	if err != nil {
		ts.T().Fatalf("got %v, want = %v", err, nil)
	}
	if res.StatusCode != http.StatusOK {
		ts.T().Fatalf("got %v, want = %v", res.StatusCode, http.StatusNoContent)
	}

	var bodyResp struct {
		Message string `json:"message"`
		Error   string `json:"error,omitempty"`
	}
	if err := json.NewDecoder(res.Body).Decode(&bodyResp); err != nil {
		ts.T().Fatalf("got %v, want = %v", err, nil)
	}

	if bodyResp.Error != "" {
		ts.T().Fatalf("got %v, want = %v", bodyResp.Error, "")
	}
}

func (ts *ApiTestSuite) testHealth() {
	client := http.Client{}

	req, err := http.NewRequest(
		http.MethodGet,
		fmt.Sprintf("http://localhost:%s/healthz", testPort),
		nil,
	)
	if err != nil {
		ts.T().Fatalf("got %v, want = %v", err, nil)
	}

	res, err := client.Do(req)
	if err != nil {
		ts.T().Fatalf("got %v, want = %v", err, nil)
	}
	if res.StatusCode != http.StatusOK {
		ts.T().Fatalf("got %v, want = %v", res.StatusCode, http.StatusNoContent)
	}
}
