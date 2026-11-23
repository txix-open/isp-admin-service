package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"isp-admin-service/domain"
	"isp-admin-service/entity"

	"github.com/txix-open/isp-kit/errors"
)

const (
	requestLimit   = 1000
	foundSymbol    = "+"
	notFoundSymbol = "-"
)

type Download struct {
	auditService      Audit
	userService       User
	tokenService      Token
	roleService       Role
	permissionService Permission
	loc               *time.Location
}

func NewDownload(
	auditService Audit,
	userService User,
	tokenService Token,
	roleService Role,
	permissionService Permission,
) Download {
	loc, _ := time.LoadLocation("Europe/Moscow")
	return Download{
		auditService:      auditService,
		userService:       userService,
		tokenService:      tokenService,
		roleService:       roleService,
		permissionService: permissionService,
		loc:               loc,
	}
}

func (s Download) DownloadAudit(ctx context.Context, req domain.AuditPageRequest, w http.ResponseWriter) error {
	req.Limit = requestLimit
	req.Offset = 0

	response, err := s.auditService.All(ctx, req)
	if err != nil {
		return errors.WithMessage(err, "get audit")
	}

	if response.TotalCount == 0 {
		return nil
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", getCsvFileName("Audit")))
	w.Header().Set("content-type", "application/octet-stream; charset=utf-8")
	w.Header().Set("transfer-encoding", "chunked")

	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	err = writeCsvLine(csvWriter, entity.AuditColumnNames())
	if err != nil {
		return errors.WithMessage(err, "write auditColumns to csv")
	}

	userById, err := s.getUsersList(ctx)
	if err != nil {
		return errors.WithMessage(err, "get users list")
	}

	itemsCount := len(response.Items)
	for {
		for _, audit := range response.Items {
			username := getUsername(userById, audit.UserId)

			row := []string{username, s.formatTime(&audit.CreatedAt, false), audit.Message}
			err = writeCsvLine(csvWriter, row)
			if err != nil {
				return errors.WithMessage(err, "write audit row to csv")
			}
		}

		if itemsCount >= response.TotalCount {
			break
		}

		req.Offset += requestLimit
		response, err = s.auditService.All(ctx, req)
		if err != nil {
			return errors.WithMessage(err, "get audit")
		}

		itemsCount += len(response.Items)
	}

	return nil
}

func (s Download) DownloadSessions(ctx context.Context, req domain.SessionPageRequest, w http.ResponseWriter) error {
	req.Limit = requestLimit
	req.Offset = 0

	response, err := s.tokenService.All(ctx, req)
	if err != nil {
		return errors.WithMessage(err, "get sessions")
	}

	if response.TotalCount == 0 {
		return nil
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", getCsvFileName("Sessions")))
	w.Header().Set("content-type", "application/octet-stream; charset=utf-8")
	w.Header().Set("transfer-encoding", "chunked")

	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	err = writeCsvLine(csvWriter, entity.SessionColumnNames())
	if err != nil {
		return errors.WithMessage(err, "write sessionColumns to csv")
	}

	userById, err := s.getUsersList(ctx)
	if err != nil {
		return errors.WithMessage(err, "get users list")
	}

	itemsCount := len(response.Items)
	for {
		for _, session := range response.Items {
			row := []string{
				getUsername(userById, session.UserId),
				strconv.Itoa(session.Id),
				s.formatTime(&session.CreatedAt, false),
				s.formatTime(&session.ExpiredAt, false),
				session.StatusRus()}

			err = writeCsvLine(csvWriter, row)
			if err != nil {
				return errors.WithMessage(err, "write session row to csv")
			}
		}

		if itemsCount >= response.TotalCount {
			break
		}

		req.Offset += requestLimit
		response, err = s.tokenService.All(ctx, req)
		if err != nil {
			return errors.WithMessage(err, "get sessions")
		}

		itemsCount += len(response.Items)
	}

	return nil
}

func (s Download) DownloadUsers(ctx context.Context, req domain.UsersPageRequest, w http.ResponseWriter) error {
	req.Limit = requestLimit
	req.Offset = 0

	users, err := s.userService.GetUsers(ctx, req)
	if err != nil {
		return errors.WithMessage(err, "get users")
	}

	roles, err := s.roleService.All(ctx, domain.RoleAllRequest{})
	if err != nil {
		return errors.WithMessage(err, "get roles")
	}

	if len(users.Items) == 0 {
		return nil
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", getCsvFileName("Users")))
	w.Header().Set("content-type", "application/octet-stream; charset=utf-8")
	w.Header().Set("transfer-encoding", "chunked")

	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	err = writeCsvLine(csvWriter, entity.UserColumnNames())
	if err != nil {
		return errors.WithMessage(err, "write userColumns to csv")
	}

	for _, user := range users.Items {
		row := []string{strconv.Itoa(int(user.Id)),
			user.FullName,
			user.Description,
			user.Email,
			getRoleNamesByIds(roles, user.Roles),
			s.formatTime(user.LastSessionCreatedAt, true)}

		err = writeCsvLine(csvWriter, row)
		if err != nil {
			return errors.WithMessage(err, "write user row to csv")
		}
	}

	return nil
}

func (s Download) DownloadRoles(ctx context.Context, w http.ResponseWriter) error {
	roles, err := s.roleService.All(ctx, domain.RoleAllRequest{})
	if err != nil {
		return errors.WithMessage(err, "get roles")
	}

	perms := s.permissionService.All(ctx)

	if len(perms) == 0 {
		return nil
	}

	columns := entity.RoleColumnNames()
	for _, role := range roles {
		columns = append(columns, role.Name)
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", getCsvFileName("Roles")))
	w.Header().Set("content-type", "application/octet-stream; charset=utf-8")
	w.Header().Set("transfer-encoding", "chunked")

	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	err = writeCsvLine(csvWriter, columns)
	if err != nil {
		return errors.WithMessage(err, "write role columns to csv")
	}

	for _, perm := range perms {
		row := []string{perm.Name, perm.Key}

		for _, role := range roles {
			if slices.Contains(role.Permissions, perm.Key) {
				row = append(row, foundSymbol)
			} else {
				row = append(row, notFoundSymbol)
			}
		}

		err = writeCsvLine(csvWriter, row)
		if err != nil {
			return errors.WithMessage(err, "write role row to csv")
		}
	}

	return nil
}

func (s Download) getUsersList(ctx context.Context) (map[int]domain.User, error) {
	users, err := s.userService.GetAllUsers(ctx)
	if err != nil {
		return nil, errors.WithMessage(err, "get users")
	}

	userById := make(map[int]domain.User)
	for _, user := range users.Items {
		userById[int(user.Id)] = user
	}

	return userById, nil
}

func getUsername(userById map[int]domain.User, userId int) string {
	user, ok := userById[userId]
	if !ok {
		return ""
	}

	return user.FullName
}

func getRoleNamesByIds(roles []domain.Role, roleIds []int) string {
	names := make([]string, 0)
	for _, role := range roles {
		if slices.Contains(roleIds, role.Id) {
			names = append(names, role.Name)
		}
	}

	return strings.Join(names, ", ")
}

func writeCsvLine(csvWriter *csv.Writer, line []string) error {
	err := csvWriter.Write(line)
	if err != nil {
		return errors.WithMessage(err, "write line to csv")
	}

	return nil
}

func getCsvFileName(pageName string) string {
	return fmt.Sprintf("%s_%s.%s", pageName, time.Now().Format("02-01-2006_15:04:05"), "csv")
}

func (s Download) formatTime(t *time.Time, onlyDate bool) string {
	if t == nil {
		return ""
	}

	if onlyDate {
		return t.In(s.loc).Format("02-01-2006")
	}

	return t.In(s.loc).Format("02-01-2006 15:04:05")
}
