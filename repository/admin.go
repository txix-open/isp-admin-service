package repository

import (
	"context"
	"isp-admin-service/domain"
	"isp-admin-service/entity"

	"github.com/txix-open/isp-kit/errors"
	"github.com/txix-open/isp-kit/http/httpcli"
)

const (
	getAllLogs     = "/admin/log/all"
	getAllSessions = "/admin/session/all"
	getAllUsers    = "/admin/user/get_users"
	getAllRoles    = "/admin/role/all"
	getPermissions = "/admin/user/get_permissions"
)

type ReqGetById struct {
	UserId int `json:"userId"`
}

type getRoleResp struct {
	Roles []int `json:"roles"`
}

type Admin struct {
	cli *httpcli.Client
}

func NewAdmin(cli *httpcli.Client) Admin {
	return Admin{cli: cli}
}

func (r Admin) GetAudit(ctx context.Context, req domain.AuditDownloadRequest) (*entity.AuditResponse, error) {
	var response entity.AuditResponse

	err := r.cli.Post(getAllLogs).
		JsonRequestBody(req).
		JsonResponseBody(&response).
		StatusCodeToError().
		DoWithoutResponse(ctx)
	if err != nil {
		return nil, errors.WithMessagef(err, "http client request %s", getAllLogs)
	}

	return &response, nil
}

func (r Admin) GetSessions(ctx context.Context, req domain.SessionDownloadRequest) (*entity.SessionResponse, error) {
	var response entity.SessionResponse

	err := r.cli.Post(getAllSessions).
		JsonRequestBody(req).
		JsonResponseBody(&response).
		StatusCodeToError().
		DoWithoutResponse(ctx)
	if err != nil {
		return nil, errors.WithMessagef(err, "http client request %s", getAllSessions)
	}

	return &response, nil
}

func (r Admin) GetUsers(ctx context.Context, req domain.UsersDownloadRequest) ([]entity.User, error) {
	var response entity.UsersResponse

	err := r.cli.Post(getAllUsers).
		JsonRequestBody(req).
		JsonResponseBody(&response).
		StatusCodeToError().
		DoWithoutResponse(ctx)
	if err != nil {
		return nil, errors.WithMessagef(err, "http client request %s", getAllUsers)
	}

	return response.Items, nil
}

func (r Admin) GetRoles(ctx context.Context) ([]entity.Role, error) {
	var response []entity.Role

	err := r.cli.Post(getAllRoles).
		JsonResponseBody(&response).
		StatusCodeToError().
		DoWithoutResponse(ctx)
	if err != nil {
		return nil, errors.WithMessagef(err, "http client request %s", getAllRoles)
	}

	return response, nil
}

func (r Admin) GetPermissions(ctx context.Context) ([]entity.Permission, error) {
	var response []entity.Permission

	err := r.cli.Post(getPermissions).
		JsonResponseBody(&response).
		StatusCodeToError().
		DoWithoutResponse(ctx)
	if err != nil {
		return nil, errors.WithMessagef(err, "http client request %s", getPermissions)
	}

	return response, nil
}
