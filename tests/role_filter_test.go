package tests_test

import (
	"net/http"
	"strconv"
	"testing"
	"time"

	"isp-admin-service/assembly"
	"isp-admin-service/conf"
	"isp-admin-service/domain"
	"isp-admin-service/entity"

	"github.com/stretchr/testify/suite"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/errors"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/dbt"
	"github.com/txix-open/isp-kit/test/httpt"
)

func TestRoleTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &RoleTestSuite{})
}

type RoleTestSuite struct {
	suite.Suite

	test    *test.Test
	db      *dbt.TestDb
	httpCli *httpcli.Client
	adminId int64
}

func (s *RoleTestSuite) SetupTest() {
	testInstance, _ := test.New(s.T())
	s.test = testInstance
	s.db = dbt.New(testInstance, dbx.WithMigrationRunner("../migrations", testInstance.Logger()))

	remote := conf.Remote{
		ExpireSec: 0,
	}
	cfg, err := assembly.NewLocator(testInstance.Logger(), httpcli.New(), s.db).
		Config(s.T().Context(), remote, time.Minute)
	s.Require().NoError(err)

	server, apiCli := httpt.TestServer(testInstance, cfg.Handler)
	s.httpCli = apiCli

	testInstance.T().Cleanup(func() {
		server.Close()
	})

	roleId := InsertRole(s.db, entity.Role{
		Name: "admin_test",
		Permissions: []string{
			"role_view",
			"role_add",
			"role_update",
		},
	})
	s.adminId = InsertUser(s.db, entity.User{
		Email: "admin@test.ru",
	})
	InsertUserRole(s.db, entity.UserRole{
		UserId: int(s.adminId),
		RoleId: int(roleId),
	})
}

func (s *RoleTestSuite) TestAllRolesWithoutFilter() {
	InsertRole(s.db, entity.Role{Name: "empty"})
	InsertRole(s.db, entity.Role{Name: "groupA_1", ApplicationGroup: "groupA"})
	InsertRole(s.db, entity.Role{Name: "groupA_2", ApplicationGroup: "groupA"})
	InsertRole(s.db, entity.Role{Name: "groupB", ApplicationGroup: "groupB"})

	roles := []domain.Role{}
	_, err := s.request("admin/role/all", domain.RoleAllRequest{}, &roles)
	s.Require().NoError(err)

	groups := make(map[string]string, len(roles))
	for _, role := range roles {
		groups[role.Name] = role.ApplicationGroup
	}

	s.Require().Empty(groups["empty"])
	s.Require().Equal("groupA", groups["groupA_1"])
	s.Require().Equal("groupA", groups["groupA_2"])
	s.Require().Equal("groupB", groups["groupB"])
}

func (s *RoleTestSuite) TestAllRolesInvalidArgument() {
	roles := []domain.Role{}
	statusCode, err := s.request("admin/role/all", domain.RoleAllRequest{ApplicationGroup: new("")}, &roles)

	s.Require().Error(err)
	s.Require().Equal(http.StatusBadRequest, statusCode)
	s.Require().Empty(roles)
}

func (s *RoleTestSuite) TestAllRolesFilterByApplicationGroup() {
	InsertRole(s.db, entity.Role{Name: "groupA", ApplicationGroup: "groupA"})
	InsertRole(s.db, entity.Role{Name: "groupB", ApplicationGroup: "groupB"})

	roles := []domain.Role{}
	_, err := s.request("admin/role/all", domain.RoleAllRequest{ApplicationGroup: new("groupA")}, &roles)
	s.Require().NoError(err)

	s.Require().Len(roles, 1)
	s.Require().Equal("groupA", roles[0].ApplicationGroup)

	roles = []domain.Role{}
	_, err = s.request("admin/role/all", domain.RoleAllRequest{ApplicationGroup: new("groupC")}, &roles)
	s.Require().NoError(err)

	s.Require().Empty(roles)
}

func (s *RoleTestSuite) TestCreateRoleApplicationGroup() {
	role := domain.Role{}
	_, err := s.request("admin/role/create", domain.CreateRoleRequest{
		Name:             "role_new",
		ApplicationGroup: "group",
		Permissions:      []string{},
	}, &role)
	s.Require().NoError(err)

	roleEmpty := domain.Role{}
	_, err = s.request("admin/role/create", domain.CreateRoleRequest{
		Name:        "role_empty",
		Permissions: []string{},
	}, &roleEmpty)
	s.Require().NoError(err)

	s.Require().Equal("role_new", role.Name)
	s.Require().Equal("group", role.ApplicationGroup)
	s.Require().Equal("role_empty", roleEmpty.Name)
	s.Require().Empty(roleEmpty.ApplicationGroup)
}

func (s *RoleTestSuite) TestUpdateRoleApplicationGroup() {
	roleId := InsertRole(s.db, entity.Role{
		Name:             "role_old",
		ApplicationGroup: "group_old",
	})

	role := domain.Role{}
	_, err := s.request("admin/role/update", domain.UpdateRoleRequest{
		Id:               int(roleId),
		Name:             "role_new",
		ApplicationGroup: "group_new",
		Permissions:      []string{},
	}, &role)
	s.Require().NoError(err)
	s.Require().Equal("role_new", role.Name)
	s.Require().Equal("group_new", role.ApplicationGroup)
}

func (s *RoleTestSuite) request(path string, req, resp any) (int, error) {
	err := s.httpCli.Post(path).
		JsonRequestBody(req).
		JsonResponseBody(resp).
		StatusCodeToError().
		Header(domain.AdminAuthIdHeader, strconv.Itoa(int(s.adminId))).
		DoWithoutResponse(s.T().Context())
	if err != nil {
		return http.StatusBadRequest, errors.WithMessagef(err, "request %s", path)
	}

	return http.StatusOK, nil
}
