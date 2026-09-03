package tests_test

import (
	"msp-admin-service/assembly"
	"msp-admin-service/conf"
	"msp-admin-service/domain"
	"net/url"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/txix-open/isp-kit/grpc/client"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/fake"
	"github.com/txix-open/isp-kit/test/grpct"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UiSettingsSuite struct {
	suite.Suite

	test               *test.Test
	apiCli             *client.Client
	oauthClientSetting conf.OauthClientConfig
	baseOauthUrl       string
}

func TestUiSettingsSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &UiSettingsSuite{
		baseOauthUrl: "http://127.0.0.1/oauth",
		oauthClientSetting: conf.OauthClientConfig{
			ClientName:   fake.It[string](),
			ClientId:     fake.It[string](),
			ClientSecret: fake.It[string](),
			RedirectURI:  "https://127.0.0.1/test/redirected",
			UiSetting: &conf.UiOauthSetting{
				LoginEndpoint:  "/login?someQueryParam=1&test=true",
				LogoutEndpoint: "/logout?redirect=url",
			},
		},
	})
}

func (s *UiSettingsSuite) SetupTest() {
	s.test, _ = test.New(s.T())
	remote := conf.Remote{
		Oauth: &conf.Oauth{
			Host:    s.baseOauthUrl,
			Clients: []conf.OauthClientConfig{s.oauthClientSetting},
		},
	}
	cfg, err := assembly.NewLocator(s.test.Logger(), nil, nil).
		Config(s.T().Context(), remote, 0)
	s.Require().NoError(err)

	_, s.apiCli = grpct.TestServer(s.test, cfg.Handler)
}

func (s *UiSettingsSuite) Test_HappyPath() {
	var resp domain.UiOauthSetting

	err := s.apiCli.Invoke("admin/ui_settings/oauth").
		JsonRequestBody(domain.ClientNameRequest{
			ClientName: s.oauthClientSetting.ClientName,
		}).
		JsonResponseBody(&resp).
		Do(s.T().Context())
	s.Require().NoError(err)

	loginURL, err := url.Parse(resp.LoginUrl)
	s.Require().NoError(err)

	s.Require().Equal(
		"http://127.0.0.1/oauth/login",
		loginURL.Scheme+"://"+loginURL.Host+loginURL.Path,
	)

	query := loginURL.Query()
	s.Require().Equal(s.oauthClientSetting.ClientId, query.Get("client_id"))
	s.Require().Equal(s.oauthClientSetting.RedirectURI, query.Get("redirect_uri"))
	s.Require().Equal("1", query.Get("someQueryParam"))
	s.Require().Equal("true", query.Get("test"))

	logoutURL, err := url.Parse(resp.LogoutUrl)
	s.Require().NoError(err)

	s.Require().Equal(
		"http://127.0.0.1/oauth/logout",
		logoutURL.Scheme+"://"+logoutURL.Host+logoutURL.Path,
	)
	s.Require().Equal("url", logoutURL.Query().Get("redirect"))
}

func (s *UiSettingsSuite) Test_UnknownClient() {
	err := s.apiCli.Invoke("admin/ui_settings/oauth").
		JsonRequestBody(domain.ClientNameRequest{
			ClientName: s.oauthClientSetting.ClientName + fake.It[string](),
		}).
		Do(s.T().Context())
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.Unauthenticated, st.Code())
}
