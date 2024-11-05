package users

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/awhdesmond/user-service/pkg/common"
	"github.com/google/go-cmp/cmp"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"github.com/upper/db/v4"
)

const (
	apiPrefix = "/hello"
)

var (
	testTimeFn = func() time.Time {
		year := time.Now().Year()
		return time.Date(year, 6, 1, 0, 0, 0, 0, time.UTC)
	}
)

type apiTestSuite struct {
	suite.Suite

	pgSess  db.Session
	rdb     redis.UniversalClient
	store   Store
	svc     Service
	handler http.Handler
}

func (ts *apiTestSuite) SetupSuite() {
	logger, _ := common.InitZap("debug")
	pgSess, err := common.MakePostgresDBSession(common.TestPgCfg)
	if err != nil {
		ts.T().Fatalf("got = %v, want = %v", err, nil)
	}
	rdb, err := common.MakeRedisClient(common.TestRedisCfg)
	if err != nil {
		ts.T().Fatalf("got = %v, want = %v", err, nil)
	}

	store := NewStore(pgSess, rdb, logger)
	svc := NewService(store, testTimeFn)

	ts.pgSess = pgSess
	ts.rdb = rdb
	ts.store = store
	ts.svc = svc
	ts.handler = MakeHandler(ts.svc)
}

func (ts *apiTestSuite) TearDownSuite() {
	_, err := ts.pgSess.SQL().Exec(common.TruncateAllTablesSQL)
	if err != nil {
		ts.T().Log(err)
	}
	ts.rdb.FlushDB(context.TODO())
	time.Sleep(3 * time.Second)
}

type UpsertApiTestSuite struct {
	apiTestSuite
}

func TestUpsertApiTestSuite(t *testing.T) {
	suite.Run(t, new(UpsertApiTestSuite))
}

func (ts *UpsertApiTestSuite) Test() {
	cases := []struct {
		name     string
		username string
		dob      string
		want     User
	}{
		{
			name:     "basic",
			username: "apple",
			dob:      "2000-01-02",
			want:     User{Username: "apple", DoB: time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)},
		},
		{
			name:     "really old person",
			username: "oldapple",
			dob:      "1900-01-02",
			want:     User{Username: "oldapple", DoB: time.Date(1900, 1, 2, 0, 0, 0, 0, time.UTC)},
		},
		{
			name:     "basic can update",
			username: "apple",
			dob:      "2001-02-03",
			want:     User{Username: "apple", DoB: time.Date(2001, 2, 3, 0, 0, 0, 0, time.UTC)},
		},
	}

	for _, tt := range cases {
		tt := tt
		ts.T().Run(tt.name, func(t *testing.T) {
			req := UpsertRequest{DoB: tt.dob}
			w := common.TestSendReq(
				req,
				fmt.Sprintf("%s/%s", apiPrefix, tt.username),
				http.MethodPut,
				ts.handler,
			)

			// status code should be HTTP 204
			if w.Code != http.StatusNoContent {
				t.Fatalf("got = %v, want = %v", w.Code, http.StatusNoContent)
			}

			// should not contain any errors
			common.TestIsResponseEmptyErr(w, t)

			usr, err := ts.store.Read(context.Background(), tt.username)
			if err != nil {
				t.Fatalf("got = %v, want = %v", err, nil)
			}

			if !cmp.Equal(usr, tt.want) {
				t.Fatalf("got = %v, want = %v", usr, tt.want)
			}
		})
	}
}

func (ts *UpsertApiTestSuite) TestErrorCases() {
	cases := []struct {
		name     string
		username string
		dob      string
		want     error
	}{
		{
			name:     "username contains non letters 1",
			username: "123aaa",
			dob:      "2000-01-02",
			want:     ErrUsernameContainsNonLetters,
		},
		{
			name:     "username contains non letters 2",
			username: "aaa@#",
			dob:      "2000-01-02",
			want:     ErrUsernameContainsNonLetters,
		},
		{
			name:     "dob in the future",
			username: "ok",
			dob:      "9000-01-02",
			want:     ErrDoBFutureUsed,
		},
		{
			name:     "dob is too old",
			username: "ok",
			dob:      "1800-01-02",
			want:     ErrDoBTooOld,
		},
		{
			name:     "dob invalid year",
			username: "ok",
			dob:      "abcd-01-02",
			want:     ErrDoBInvalid,
		},
		{
			name:     "dob invalid month",
			username: "ok",
			dob:      "2000-13-02",
			want:     ErrDoBInvalid,
		},
		{
			name:     "dob invalid day",
			username: "ok",
			dob:      "2000-01-44",
			want:     ErrDoBInvalid,
		},
		{
			name:     "dob invalid leap day in non-leap year",
			username: "ok",
			dob:      "2013-02-29",
			want:     ErrDoBInvalid,
		},
	}

	for _, tt := range cases {
		tt := tt
		ts.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req := UpsertRequest{DoB: tt.dob}
			w := common.TestSendReq(
				req,
				fmt.Sprintf("%s/%s", apiPrefix, tt.username),
				http.MethodPut,
				ts.handler,
			)

			common.TestIsResponseErrorExpected(w, ts.T(), tt.want.Error())
		})
	}

}

type ReadApiTestSuite struct {
	apiTestSuite
}

func (ts *ReadApiTestSuite) SetupSuite() {
	ts.apiTestSuite.SetupSuite()

	// bootstrap some data
	err := ts.svc.Upsert(context.Background(), "apple", "2000-03-03")
	if err != nil {
		ts.T().Fatalf("got = %v, want = %v", err, nil)
	}
	err = ts.svc.Upsert(context.Background(), "mango", "2000-06-01")
	if err != nil {
		ts.T().Fatalf("got = %v, want = %v", err, nil)
	}
	err = ts.svc.Upsert(context.Background(), "pear", "2000-07-03")
	if err != nil {
		ts.T().Fatalf("got = %v, want = %v", err, nil)
	}
}

func TestReadApiTestSuite(t *testing.T) {
	suite.Run(t, new(ReadApiTestSuite))
}

func (ts *ReadApiTestSuite) Test() {
	// NOTE: we need to handle this edge case
	// in order for the test to always work in both leap and non-leap years.
	daysToAddForLeapYear := 0
	if common.IsLeapYear(testTimeFn().Year() + 1) {
		daysToAddForLeapYear += 1
	}

	cases := []struct {
		name     string
		username string
		want     string
	}{
		{
			name:     "birthday has passed",
			username: "apple",
			want:     fmt.Sprintf("Hello, apple! Your birthday is in %d day(s)", 275+daysToAddForLeapYear),
		},
		{
			name:     "birthday has passed + read from cache",
			username: "apple",
			want:     fmt.Sprintf("Hello, apple! Your birthday is in %d day(s)", 275+daysToAddForLeapYear),
		},
		{
			name:     "birthday has not passed",
			username: "pear",
			want:     fmt.Sprintf("Hello, pear! Your birthday is in %d day(s)", 32),
		},
		{
			name:     "birthday is today",
			username: "mango",
			want:     "Hello, mango! Happy birthday!",
		},
	}

	for _, tt := range cases {
		tt := tt
		ts.T().Run(tt.name, func(t *testing.T) {
			w := common.TestSendReq(
				nil,
				fmt.Sprintf("%s/%s", apiPrefix, tt.username),
				http.MethodGet,
				ts.handler,
			)

			// status code should be HTTP 200
			if w.Code != http.StatusOK {
				t.Fatalf("got = %v, want = %v", w.Code, http.StatusOK)
			}

			var resp ReadResponse
			if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
				t.Fatalf("got = %v, want = %v", err, nil)
			}

			if !cmp.Equal(resp.Message, tt.want) {
				t.Fatalf("got = %v, want = %v", resp.Message, tt.want)
			}
			time.Sleep(1000 * time.Millisecond)
		})
	}
}

func (ts *ReadApiTestSuite) TestErrors() {
	cases := []struct {
		name     string
		username string
		want     error
	}{
		{
			name:     "username contains non letters 1",
			username: "123aaa",
			want:     ErrUsernameContainsNonLetters,
		},
		{
			name:     "username not found",
			username: "grape",
			want:     ErrUserNotFound,
		},
	}
	for _, tt := range cases {
		tt := tt
		ts.T().Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := common.TestSendReq(
				nil,
				fmt.Sprintf("%s/%s", apiPrefix, tt.username),
				http.MethodGet,
				ts.handler,
			)

			common.TestIsResponseErrorExpected(w, ts.T(), tt.want.Error())
		})
	}
}
