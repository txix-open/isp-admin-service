// nolint: funlen
package tests_test

import (
	"encoding/csv"
	"net/http"
	"testing"
	"time"

	"isp-admin-service/assembly"
	"isp-admin-service/conf"
	"isp-admin-service/domain"
	"isp-admin-service/entity"
	"isp-admin-service/repository"

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
	cfg, err := assembly.NewLocator(testInstance.Logger(), nil, s.db).
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
		JsonRequestBody(domain.UsersPageRequest{
			Limit:  1000,
			Offset: 0,
			Order: &domain.OrderParams{
				Field: "id",
				Type:  "desc",
			},
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

	s.Require().Len(records, 4)

	row := records[1]
	s.Require().EqualValues("3", row[idIdx])
	s.Require().EqualValues("alice@a.ru", row[emailIdx])
	s.Require().EqualValues("Bob Alice", row[fullNameIdx])
}

func (s *DownloadSuite) TestSessionsHappyPath() {
	timeNow := time.Now().UTC()
	userId := InsertUser(s.db, entity.User{Email: "sess1@a.ru", FullName: "Bob Alice"})
	roleId := InsertRole(s.db, entity.Role{Name: "admin", Permissions: s.permissions})
	InsertUserRole(s.db, entity.UserRole{RoleId: int(roleId), UserId: int(userId)})

	InsertTokenEntity(s.db, entity.Token{
		Token:     "test_token_1",
		UserId:    userId,
		Status:    entity.TokenStatusAllowed,
		ExpiredAt: timeNow.Add(1 * time.Hour),
		CreatedAt: timeNow.Add(-1 * time.Hour),
		UpdatedAt: timeNow,
	})
	InsertTokenEntity(s.db, entity.Token{
		Token:     "test_token_2",
		UserId:    userId,
		Status:    entity.TokenStatusRevoked,
		ExpiredAt: timeNow.Add(1 * time.Hour),
		CreatedAt: timeNow.Add(-2 * time.Hour),
		UpdatedAt: timeNow,
	})
	InsertTokenEntity(s.db, entity.Token{
		Token:     "test_token_3",
		UserId:    userId,
		Status:    entity.TokenStatusExpired,
		ExpiredAt: timeNow.Add(-1 * time.Hour),
		CreatedAt: timeNow.Add(-3 * time.Hour),
		UpdatedAt: timeNow,
	})

	resp, err := s.httpCli.Post("/session/download").
		JsonRequestBody(domain.SessionPageRequest{
			Limit:  1000,
			Offset: 0,
			Order: &domain.OrderParams{
				Field: "user_id",
				Type:  "desc",
			},
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

	s.Require().Len(records, 4)
	row := records[1]
	s.Require().EqualValues("Bob Alice", row[userIdx])
	s.Require().EqualValues("1", row[idIdx])
}

func (s *DownloadSuite) TestAuditHappyPath() {
	timeNow := time.Now().UTC()
	userId := InsertUser(s.db, entity.User{Email: "audit1@a.ru", FullName: "Bob Alice"})
	roleId := InsertRole(s.db, entity.Role{Name: "admin", Permissions: s.permissions})
	InsertUserRole(s.db, entity.UserRole{RoleId: int(roleId), UserId: int(userId)})

	auditRepo := repository.NewAudit(s.db)
	_, err := auditRepo.Insert(s.T().Context(), entity.Audit{
		Id:        1,
		UserId:    int(userId),
		Message:   "audit1@a.ru",
		Event:     "123456",
		CreatedAt: timeNow,
	})
	s.Require().NoError(err)

	resp, err := s.httpCli.Post("log/download").
		JsonRequestBody(domain.AuditPageRequest{
			Limit:  1000,
			Offset: 0,
			Order: &domain.OrderParams{
				Field: "user_id",
				Type:  "desc",
			},
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
	moveIdx, ok := col["Действие"]
	s.Require().True(ok)

	s.Require().Len(records, 2)
	row := records[1]
	s.Require().EqualValues("Bob Alice", row[idIdx])
	s.Require().EqualValues("audit1@a.ru", row[moveIdx])
}

func (s *DownloadSuite) TestRolesHappyPath() {
	resp, err := s.httpCli.Post("role/download").
		StatusCodeToError().
		Do(s.T().Context())
	s.Require().NoError(err)

	records, err := csv.NewReader(resp.Raw.Body).ReadAll()
	s.Require().NoError(err)
	s.Require().NotEmpty(records)
}
