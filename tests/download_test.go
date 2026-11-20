package tests_test

import (
	"encoding/csv"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"isp-admin-service/assembly"
	"isp-admin-service/conf"
	"isp-admin-service/domain"
	"isp-admin-service/entity"

	"github.com/stretchr/testify/suite"

	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/dbt"
	"github.com/txix-open/isp-kit/test/httpt"
)

func TestDownloadSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &DownloadSuite{})
}

type DownloadSuite struct {
	suite.Suite

	test        *test.Test
	db          *dbt.TestDb
	httpCli     *httpcli.Client
	downloadCli *httpcli.Client
	permissions entity.PermList
}

func (s *DownloadSuite) SetupTest() {
	testInstance, _ := test.New(s.T())
	s.test = testInstance
	s.db = dbt.New(testInstance, dbx.WithMigrationRunner("../migrations", testInstance.Logger()))

	remote := conf.Remote{
		Permissions: []conf.Permission{
			{
				Key:  "data_download",
				Name: "123",
			},
		},
	}
	s.initMockDownload()
	cfg, err := assembly.NewLocator(testInstance.Logger(), nil, s.downloadCli, s.db).
		Config(s.T().Context(), remote, time.Minute)
	s.Require().NoError(err)

	server, cli := httpt.TestServer(s.test, cfg.Handler)
	s.httpCli = cli

	s.permissions = entity.PermList{
		"data_download",
	}

	testInstance.T().Cleanup(func() {
		server.Close()
	})
}

func (s *DownloadSuite) TestUsersHappyPath() {
	id1 := InsertUser(s.db, entity.User{
		FirstName: "Alice",
		LastName:  "Bob",
		FullName:  "Bob Alice",
		Email:     "alice@a.ru",
		Password:  "password",
	})
	roleId := InsertRole(s.db, entity.Role{Name: "admin", Permissions: s.permissions})
	InsertUserRole(s.db, entity.UserRole{RoleId: int(roleId), UserId: int(id1)})

	resp, err := s.httpCli.Post("user/download").
		JsonRequestBody(domain.UsersDownloadRequest{
			LimitOffsetParams: domain.LimitOffsetParams{Limit: 1000, Offset: 0},
		}).
		Do(s.T().Context())
	s.Require().NoError(err)
	defer resp.Raw.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode())

	records, err := csv.NewReader(resp.Raw.Body).ReadAll()
	s.Require().NoError(err)
	s.Require().NotEmpty(records)

	header := records[0]
	col := make(map[string]int, len(header))
	for i, name := range header {
		col[name] = i
	}

	idIdx, ok := col["id"]
	s.Require().True(ok)
	emailIdx, ok := col["Email"]
	s.Require().True(ok)
	fullNameIdx, ok := col["Пользователь"]
	s.Require().True(ok)

	s.Require().Len(records, 2)

	row := records[1]
	s.Require().EqualValues("1", row[idIdx])
	s.Require().EqualValues("alice@a.ru", row[emailIdx])
	s.Require().EqualValues("Bob Alice", row[fullNameIdx])
}

func (s *DownloadSuite) TestSessionsHappyPath() {
	id1 := InsertUser(s.db, entity.User{Email: "sess1@a.ru"})
	roleId := InsertRole(s.db, entity.Role{Name: "admin", Permissions: s.permissions})
	InsertUserRole(s.db, entity.UserRole{RoleId: int(roleId), UserId: int(id1)})

	resp, err := s.httpCli.Post("/session/download").
		JsonRequestBody(domain.SessionDownloadRequest{
			LimitOffsetParams: domain.LimitOffsetParams{Limit: 1000, Offset: 0},
		}).
		Do(s.T().Context())
	s.Require().NoError(err)
	defer resp.Raw.Body.Close()
	s.Require().Equal(http.StatusOK, resp.StatusCode())

	records, err := csv.NewReader(resp.Raw.Body).ReadAll()
	s.Require().NoError(err)
	s.Require().NotEmpty(records)

	header := records[0]
	col := make(map[string]int, len(header))
	for i, name := range header {
		col[name] = i
	}

	userIdx, ok := col["Пользователь"]
	s.Require().True(ok)
	idIdx, ok := col["Идентификатор"]
	s.Require().True(ok)

	s.Require().Len(records, 2)
	row := records[1]
	s.Require().EqualValues("Bob Alice", row[userIdx])
	s.Require().EqualValues("1", row[idIdx])
}

func (s *DownloadSuite) TestAuditHappyPath() {
	id1 := InsertUser(s.db, entity.User{Email: "audit1@a.ru"})
	roleId := InsertRole(s.db, entity.Role{Name: "admin", Permissions: s.permissions})
	InsertUserRole(s.db, entity.UserRole{RoleId: int(roleId), UserId: int(id1)})

	resp, err := s.httpCli.Post("log/download").
		JsonRequestBody(domain.AuditDownloadRequest{
			LimitOffsetParams: domain.LimitOffsetParams{Limit: 1000, Offset: 0},
		}).
		StatusCodeToError().
		Do(s.T().Context())
	s.Require().NoError(err)

	records, err := csv.NewReader(resp.Raw.Body).ReadAll()
	s.Require().NoError(err)
	s.Require().NotEmpty(records)

	header := records[0]
	col := make(map[string]int, len(header))
	for i, name := range header {
		col[name] = i
	}

	idIdx, ok := col["Пользователь"]
	s.Require().True(ok)
	moveIdx, ok := col["Действие"] // уточнить
	s.Require().True(ok)

	s.Require().Len(records, 2)
	row := records[1]
	s.Require().EqualValues("Bob Alice", row[idIdx])
	s.Require().EqualValues("audit1@a.ru", row[moveIdx])
}

func (s *DownloadSuite) TestRolesHappyPath() {
	err := s.httpCli.Post("role/download").
		StatusCodeToError().
		DoWithoutResponse(s.T().Context())
	s.Require().NoError(err)
}

func (s *DownloadSuite) initMockDownload() {
	mocksrv := httpt.NewMock(s.test)
	mocksrv.POST("/user/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"Items": []domain.User{{Id: 0, FullName: "Bob Alice", Email: "alice@a.ru"}},
		})
	})
	mocksrv.POST("/session/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"Items": []domain.User{{Id: 0, FullName: "Sess1", Email: "sess1@a.ru"}},
		})
	})
	mocksrv.POST("/log/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"TotalCount": 1,
			"Items":      []domain.Audit{{Id: 0, UserId: 0, Message: "audit1@a.ru"}},
		})
	})
	mocksrv.POST("/role/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"Items": []domain.User{{Id: 0, Roles: []int{0}, Email: "admin@a.ru"}},
		})
	})

	mocksrv.POST("/admin/user/get_users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"Items": []domain.User{{
				Id:       1,
				FullName: "Bob Alice",
				Email:    "alice@a.ru",
			}},
			"TotalCount": 1,
		})
	})

	mocksrv.POST("/admin/user/get_permissions", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]domain.Permission{{
			Key:  "data_download",
			Name: "123",
		}})

	})

	mocksrv.POST("/admin/session/all", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"Items": []domain.Session{{
				Id:     1,
				UserId: 1,
			}},
			"TotalCount": 1,
		})
	})

	mocksrv.POST("/admin/log/all", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]any{
			"Items": []domain.Audit{{
				Id:      1,
				UserId:  1,
				Message: "audit1@a.ru",
			}},
			"TotalCount": 1,
		})
	})

	mocksrv.POST("/admin/role/all", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]domain.Role{{
			Id:   1,
			Name: "admin",
		}})
	})

	s.downloadCli = mocksrv.Client()
}
