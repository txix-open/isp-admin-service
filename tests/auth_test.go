package tests_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"isp-admin-service/assembly"
	"isp-admin-service/conf"
	"isp-admin-service/domain"
	"isp-admin-service/entity"
	"isp-admin-service/repository"

	"github.com/txix-open/isp-kit/errors"
	"github.com/txix-open/isp-kit/test/fake"
	"github.com/txix-open/isp-kit/test/httpt"
	"golang.org/x/sync/errgroup"

	"github.com/stretchr/testify/suite"
	"github.com/txix-open/isp-kit/dbx"
	"github.com/txix-open/isp-kit/http/httpcli"
	"github.com/txix-open/isp-kit/json"
	"github.com/txix-open/isp-kit/test"
	"github.com/txix-open/isp-kit/test/dbt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestAuthTestSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, &AuthTestSuite{})
}

type AuthTestSuite struct {
	suite.Suite

	test            *test.Test
	db              *dbt.TestDb
	httpCli         *httpcli.Client
	oauthCli        *httpcli.Client
	oauthClientName string
}

func (s *AuthTestSuite) SetupTest() {
	testInstance, _ := test.New(s.T())
	s.test = testInstance
	s.db = dbt.New(testInstance, dbx.WithMigrationRunner("../migrations", testInstance.Logger()))
	s.oauthCli = httpcli.New()
	s.oauthClientName = fake.It[string]()

	mocksrv, host := s.initMockOauth()
	oauthHost, _ := url.JoinPath(host, "/blitz/")
	s.oauthCli.GlobalRequestConfig().BaseUrl = oauthHost

	remote := conf.Remote{
		Oauth: &conf.Oauth{
			Host: host,
			Clients: []conf.OauthClientConfig{
				{
					ClientName:   s.oauthClientName,
					ClientId:     "admin",
					ClientSecret: "admin",
					RedirectURI:  "http://localhost",
				},
			},
		},
		ExpireSec: 3600,
		AntiBruteforce: conf.AntiBruteforce{
			MaxInFlightLoginRequests: 3,
			DelayLoginRequestInSec:   1,
		},
	}

	cfg, err := assembly.NewLocator(testInstance.Logger(), s.oauthCli, httpcli.New(), s.db).
		Config(s.T().Context(), remote, time.Minute)
	s.Require().NoError(err)

	server, cli := httpt.TestServer(s.test, cfg.Handler)
	s.httpCli = cli

	testInstance.T().Cleanup(func() {
		server.Close()
		mocksrv.Close()
	})
}

func (s *AuthTestSuite) TestLoginHappyPath() {
	id := InsertUser(s.db, entity.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "a@a.ru",
		Password:  "password",
	})

	response := domain.LoginResponse{}
	err := s.httpCli.Post("admin/auth/login").
		JsonRequestBody(domain.LoginRequest{
			Email:    "a@a.ru",
			Password: "password",
		}).
		JsonResponseBody(&response).
		DoWithoutResponse(s.T().Context())
	s.Require().NoError(err)

	tokenInfo := SelectTokenEntityByToken(s.db, response.Token)
	s.Require().Equal(tokenInfo.UserId, id)

	time.Sleep(1 * time.Second) // wait for go SaveAuditAsync()
}

func (s *AuthTestSuite) TestLoginNotFound() {
	err := s.httpCli.Post("admin/auth/login").
		JsonRequestBody(domain.LoginRequest{
			Email:    "a1@a.ru",
			Password: "password",
		}).
		DoWithoutResponse(s.T().Context())
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.Unauthenticated, st.Code())

	time.Sleep(1 * time.Second) // wait for go SaveAuditAsync()
}

func (s *AuthTestSuite) TestBlockedUser() {
	InsertUser(s.db, entity.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "a@a.ru",
		Password:  "password",
		Blocked:   true,
	})

	err := s.httpCli.Post("admin/auth/login").
		JsonRequestBody(domain.LoginRequest{
			Email:    "a@a.ru",
			Password: "password",
		}).
		DoWithoutResponse(s.T().Context())
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.Unauthenticated, st.Code())

	time.Sleep(1 * time.Second) // wait for go SaveAuditAsync()
}

func (s *AuthTestSuite) TestLoginWrongPassword() {
	InsertUser(s.db, entity.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "a@a.ru",
		Password:  "password",
	})

	err := s.httpCli.Post("admin/auth/login").
		JsonRequestBody(domain.LoginRequest{
			Email:    "a@a.ru",
			Password: "WrongPassword",
		}).
		DoWithoutResponse(s.T().Context())
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.Unauthenticated, st.Code())

	time.Sleep(1 * time.Second) // wait for go SaveAuditAsync()
}

func (s *AuthTestSuite) TestOauthLoginHappyPath() {
	response := domain.LoginResponse{}

	err := s.httpCli.Post("admin/auth/login_with_oauth").
		JsonRequestBody(domain.LoginOauthRequest{
			AuthCode:   "code",
			ClientName: s.oauthClientName,
		}).
		JsonResponseBody(&response).
		DoWithoutResponse(s.T().Context())
	s.Require().NoError(err)
	user := entity.User{}
	s.db.Must().SelectRow(&user, "select id, email, full_name from users where oauth_user_id = $1", "oauthUser1")
	s.Require().Equal("oauth@email.ru", user.Email)
	s.Require().Equal("surname name patronymic", user.FullName)

	tokenInfo := SelectTokenEntityByToken(s.db, response.Token)
	s.Require().Equal(tokenInfo.UserId, user.Id)

	time.Sleep(1 * time.Second) // wait for go SaveAuditAsync()
}

func (s *AuthTestSuite) TestOauthLogin_UnknownClient() {
	response := domain.LoginResponse{}

	err := s.httpCli.Post("admin/auth/login_with_oauth").
		JsonRequestBody(domain.LoginOauthRequest{
			AuthCode:   "code",
			ClientName: s.oauthClientName + fake.It[string](),
		}).
		JsonResponseBody(&response).
		DoWithoutResponse(s.T().Context())
	s.Require().Error(err)
	st, ok := status.FromError(err)
	s.Require().True(ok)
	s.Require().Equal(codes.Unauthenticated, st.Code())
}

func (s *AuthTestSuite) Test_Logout_HappyPath() {
	userId := InsertUser(s.db, entity.User{Email: "suslik@mail.ru"})
	InsertTokenEntity(s.db, entity.Token{
		Token:     "token-841297641213",
		UserId:    userId,
		Status:    entity.TokenStatusAllowed,
		CreatedAt: time.Time{},
		ExpiredAt: time.Time{},
	})
	err := s.httpCli.Post("admin/auth/logout").
		Header(domain.AdminAuthIdHeader, strconv.Itoa(int(userId))).
		DoWithoutResponse(s.T().Context())
	s.Require().NoError(err)

	tokenInfo := SelectTokenEntityByToken(s.db, "token-841297641213")
	s.Require().Equal(entity.TokenStatusRevoked, tokenInfo.Status)

	time.Sleep(1 * time.Second) // wait for go SaveAuditAsync()
}

func (s *AuthTestSuite) Test_Logout_NotFound() {
	err := s.httpCli.Post("admin/auth/logout").
		Header(domain.AdminAuthIdHeader, "0143218411981").
		DoWithoutResponse(s.T().Context())
	s.Require().NoError(err)

	time.Sleep(time.Second)
	audit := repository.NewAudit(s.db)
	auditList, err := audit.AllByRequest(s.T().Context(), domain.AuditPageRequest{
		LimitOffsetParams: domain.LimitOffsetParams{
			Limit:  10,
			Offset: 0,
		},
		Order: &domain.OrderParams{
			Field: "created_at",
			Type:  "desc",
		},
	})
	s.Require().NoError(err)
	s.Require().Len(auditList, 1)
	s.Require().Equal(entity.EventSuccessLogout, auditList[0].Event)

	time.Sleep(1 * time.Second) // wait for go SaveAuditAsync()
}

func (s *AuthTestSuite) Test_Logout_AlreadyRevoke() {
	userId := InsertUser(s.db, entity.User{Email: "suslik@mail.ru"})
	InsertTokenEntity(s.db, entity.Token{
		Token:     "token-148623719462",
		UserId:    userId,
		Status:    entity.TokenStatusRevoked,
		CreatedAt: time.Time{},
		ExpiredAt: time.Time{},
	})
	err := s.httpCli.Post("admin/auth/logout").
		Header(domain.AdminAuthIdHeader, strconv.Itoa(int(userId))).
		DoWithoutResponse(s.T().Context())
	s.Require().NoError(err)

	tokenInfo := SelectTokenEntityByToken(s.db, "token-148623719462")
	s.Require().Equal(entity.TokenStatusRevoked, tokenInfo.Status)

	time.Sleep(1 * time.Second) // wait for go SaveAuditAsync()
}

func (s *AuthTestSuite) TestBruteForceLogin() {
	_ = InsertUser(s.db, entity.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "a@a.ru",
		Password:  "password",
	})

	tooManyRequestsErrorCount := &atomic.Int32{}
	unauthorizedErrorCount := &atomic.Int32{}
	group, ctx := errgroup.WithContext(s.T().Context())
	for index := range 100 {
		group.Go(func() error {
			start := time.Now()
			response := domain.LoginResponse{}
			err := s.httpCli.Post("admin/auth/login").
				JsonRequestBody(domain.LoginRequest{
					Email:    "a@a.ru",
					Password: fmt.Sprintf("password %s", strconv.Itoa(index)),
				}).
				JsonResponseBody(&response).
				DoWithoutResponse(ctx)
			s.Require().Error(err)

			switch status.Code(err) { // nolint:exhaustive
			case codes.ResourceExhausted:
				tooManyRequestsErrorCount.Add(1)
				s.Require().True(time.Since(start) < time.Second)
			case codes.Unauthenticated:
				unauthorizedErrorCount.Add(1)
				s.Require().True(time.Since(start) > time.Second)
			default:
				s.Require().NoError(errors.New("never happen"))
			}

			return nil
		})
	}

	err := group.Wait()
	s.Require().NoError(err)

	s.Require().EqualValues(97, tooManyRequestsErrorCount.Load())
	s.Require().EqualValues(3, unauthorizedErrorCount.Load())

	time.Sleep(2 * time.Second) // wait for go SaveAuditAsync()
}

func (s *AuthTestSuite) initMockOauth() (*httptest.Server, string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/blitz/oauth/te", func(writer http.ResponseWriter, request *http.Request) {
		res := entity.OauthTokenResponse{
			OauthAuthError: nil,
			IdToken:        "1",
			AccessToken:    "token",
		}
		data, err := json.Marshal(res)
		s.Require().NoError(err) //nolint:testifylint
		_, err = writer.Write(data)
		s.NoError(err)
	})
	mux.HandleFunc("/blitz/oauth/me", func(writer http.ResponseWriter, request *http.Request) {
		res := entity.OauthUserResponse{
			OauthAuthError: nil,
			Email:          "oauth@email.ru",
			Groups:         []string{"DIT-KKD-Admins"},
			Sub:            "oauthUser1",
			GivenName:      "name",
			FamilyName:     "surname",
			Name:           "surname name patronymic",
		}
		data, err := json.Marshal(res)
		s.Require().NoError(err) //nolint:testifylint
		_, err = writer.Write(data)
		s.NoError(err)
	})
	srv := httptest.NewServer(mux)
	return srv, srv.URL
}
