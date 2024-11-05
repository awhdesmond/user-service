package common

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func GetPostgresTestDb() string {
	testDB := os.Getenv("USERS_SVC_POSTGRES_TEST_DATABASE")
	if testDB == "" {
		testDB = "postgres_test"
	}
	return testDB
}

var (
	TestPgCfg = PostgresSQLConfig{
		Host:     "localhost",
		Port:     "5432",
		Username: "postgres",
		Password: "postgres",
		Database: GetPostgresTestDb(),
	}
	TestRedisCfg = RedisCfg{
		URI: "redis://localhost:6379/10",
	}
	TruncateAllTablesSQL = `TRUNCATE TABLE users;`
)

func TestSendReq(req interface{}, path, method string, handler http.Handler) *httptest.ResponseRecorder {
	var httpReq *http.Request
	if req != nil {
		data, _ := json.Marshal(req)
		httpReq = httptest.NewRequest(method, path, bytes.NewReader(data))
	} else {
		httpReq = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httpReq)
	return w
}

func TestIsResponseEmptyErr(w *httptest.ResponseRecorder, t *testing.T) {
	var data []byte
	_, err := w.Body.Read(data)
	if err != nil {
		t.Fatalf("got = %v, want = %v", err, nil)
	}

	if string(data) != "" {
		t.Fatalf("got = %v, want = %v", string(data), "")
	}
}

func TestIsResponseErrorExpected(w *httptest.ResponseRecorder, t *testing.T, wantErr string) {
	res := struct {
		Err string `json:"error,omitempty"`
	}{}
	if err := json.NewDecoder(w.Body).Decode(&res); err != nil {
		t.Fatalf("got = % v, want = %v", err, nil)
	}
	if res.Err != wantErr {
		t.Fatalf("got = %v, want = %v", res.Err, wantErr)
	}
}
