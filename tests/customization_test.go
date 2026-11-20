package tests_test

import (
	"testing"
	"time"

	"isp-admin-service/assembly"
	"isp-admin-service/conf"

	"github.com/stretchr/testify/suite"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/dbt"
	"github.com/txix-open/isp-kit/test/httpt"
)

func TestCustomizationTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &CustomizationTestSuite{})
}

type CustomizationTestSuite struct {
	suite.Suite

	test    *test.Test
	db      *dbt.TestDb
	httpCli *httpcli.Client
}

func (s *CustomizationTestSuite) SetupTest() {
	testInstance, _ := test.New(s.T())
	s.test = testInstance
	s.db = dbt.New(testInstance, dbx.WithMigrationRunner("../migrations", testInstance.Logger()))

	remote := conf.Remote{
		UiDesign: conf.UIDesign{
			Name:         "test",
			PrimaryColor: "#ff4d4f",
		},
		ExpireSec: 0,
	}
	cfg, err := assembly.NewLocator(testInstance.Logger(), nil, s.db).
		Config(s.T().Context(), remote, time.Minute)
	s.Require().NoError(err)

	server, cli := httpt.TestServer(s.test, cfg.Handler)
	s.httpCli = cli

	testInstance.T().Cleanup(func() {
		server.Close()
	})
}

func (s *CustomizationTestSuite) TestGetDesign() {
	response := conf.UIDesign{}
	err := s.httpCli.
		Post("admin/user/get_design").
		JsonRequestBody(struct{}{}).
		JsonResponseBody(&response).
		DoWithoutResponse(s.T().Context())
	s.Require().NoError(err)
	expected := conf.UIDesign{
		Name:         "test",
		PrimaryColor: "#ff4d4f",
	}
	s.Require().Equal(expected, response)
}
