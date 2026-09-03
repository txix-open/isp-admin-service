package service

import (
	"context"

	"msp-admin-service/conf"
	"msp-admin-service/domain"
	"msp-admin-service/entity"

	"github.com/pkg/errors"
)

type oauthRepo interface {
	GetToken(ctx context.Context, creds entity.ClientSetting) (*entity.OauthTokenResponse, error)
	GetUser(ctx context.Context, accessToken string) (*entity.OauthUserResponse, error)
}

type roleRepo interface {
	GetRolesByExternalGroup(ctx context.Context, groups []string) ([]entity.Role, error)
	InsertRole(ctx context.Context, role entity.Role) (*entity.Role, error)
}

type Oauth struct {
	oauthRepo      oauthRepo
	clientSettings map[string]conf.OauthClientConfig
}

func NewOauth(settings []conf.OauthClientConfig, oauthRepo oauthRepo) (Oauth, error) {
	clientSettings := make(map[string]conf.OauthClientConfig, len(settings))
	for i, setting := range settings {
		_, ok := clientSettings[setting.ClientName]
		if ok {
			return Oauth{}, errors.Errorf("client setting [%d] has non-unique client name", i)
		}
		clientSettings[setting.ClientName] = setting
	}

	return Oauth{
		oauthRepo:      oauthRepo,
		clientSettings: clientSettings,
	}, nil
}

func (s Oauth) Authenticate(ctx context.Context, req domain.LoginOauthRequest, roleRepo roleRepo) (*entity.OauthUser, error) {
	if len(s.clientSettings) == 0 {
		return nil, domain.ErrOauthIsMissed
	}

	setting, ok := s.clientSettings[req.ClientName]
	if !ok {
		return nil, domain.ErrUnauthenticated
	}

	creds := entity.ClientSetting{
		AuthCode:     req.AuthCode,
		ClientId:     setting.ClientId,
		ClientSecret: setting.ClientSecret,
		RedirectURI:  setting.RedirectURI,
	}
	tokenResponse, err := s.oauthRepo.GetToken(ctx, creds)
	switch {
	case err != nil:
		return nil, errors.WithMessage(err, "get token")
	case tokenResponse.OauthAuthError != nil:
		return nil, errors.WithMessage(tokenResponse.OauthAuthError, "get token")
	}

	user, err := s.oauthRepo.GetUser(ctx, tokenResponse.AccessToken)
	switch {
	case err != nil:
		return nil, errors.WithMessage(err, "get user")
	case user.OauthAuthError != nil:
		return nil, errors.WithMessage(user.OauthAuthError, "get user")
	}
	email := user.Email
	if email == "" {
		email = user.Sub
	}

	rolesIds := make([]int, 0)
	if len(user.Groups) > 0 {
		roles, err := roleRepo.GetRolesByExternalGroup(ctx, user.Groups)
		if err != nil {
			return nil, errors.WithMessage(err, "get roles by external groups")
		}
		for _, role := range roles {
			rolesIds = append(rolesIds, role.Id)
		}
	}

	return &entity.OauthUser{
		RoleIds:     rolesIds,
		OauthUserId: user.Sub,
		FirstName:   user.GivenName,
		LastName:    user.FamilyName,
		FullName:    user.Name,
		Email:       email,
	}, nil
}
