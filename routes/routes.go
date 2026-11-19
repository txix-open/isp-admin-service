package routes

import (
	"isp-admin-service/controller"
	"isp-admin-service/docs"
	"net/http"

	"github.com/txix-open/isp-kit/cluster"
	"github.com/txix-open/isp-kit/common_endpoints"
	"github.com/txix-open/isp-kit/http/endpoint"
	"github.com/txix-open/isp-kit/http/router"
)

type Controllers struct {
	Auth          controller.Auth
	User          controller.User
	Customization controller.Customization
	Secure        controller.Secure
	Session       controller.Session
	Audit         controller.Audit
	Role          controller.Role
	Permissions   controller.Permissions
	UiSettings    controller.UiSettings
	Download      controller.Download
}

func EndpointDescriptors() []cluster.EndpointDescriptor {
	return concatEndpoints(
		endpointDescriptors(Controllers{}),
		commonEndpoints(),
	)
}

func Handler(wrapper endpoint.Wrapper, c Controllers) *router.Router { // nolint:ireturn
	muxer := router.New()
	for _, descriptor := range endpointDescriptors(c) {
		muxer.Handler(descriptor.HttpMethod, "/"+descriptor.Path, wrapper.Endpoint(descriptor.Handler))
	}
	for _, descriptor := range commonEndpoints() {
		muxer.Handler(descriptor.HttpMethod, "/"+descriptor.Path, wrapper.Endpoint(descriptor.Handler))
	}
	return muxer
}

//nolint:funlen
func endpointDescriptors(c Controllers) []cluster.EndpointDescriptor {
	return []cluster.EndpointDescriptor{
		{
			Path:       "admin/auth/login",
			Inner:      false,
			Handler:    c.Auth.Login,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/auth/login_with_oauth",
			Inner:      false,
			Handler:    c.Auth.LoginWithOauth,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/auth/logout",
			Inner:      true,
			Handler:    c.Auth.Logout,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/auth/logout_with_reason",
			Inner:      true,
			Handler:    c.Auth.LogoutWithReason,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/user/get_profile",
			Inner:      true,
			Handler:    c.User.GetProfile,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/user/get_design",
			Inner:      true,
			Handler:    c.Customization.GetUIDesign,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/user/get_users",
			Inner:      true,
			Handler:    c.User.GetUsers,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/user/get_all_users",
			Inner:      true,
			Handler:    c.User.GetAllUsers,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/user/create_user",
			Inner:      true,
			Extra:      cluster.RequireAdminPermission("user_create"),
			Handler:    c.User.CreateUser,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/user/update_user",
			Inner:      true,
			Extra:      cluster.RequireAdminPermission("user_update"),
			Handler:    c.User.UpdateUser,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/user/delete_user",
			Inner:      true,
			Extra:      cluster.RequireAdminPermission("user_delete"),
			Handler:    c.User.DeleteUser,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/user/block_user",
			Inner:      true,
			Extra:      cluster.RequireAdminPermission("user_block"),
			Handler:    c.User.Block,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/user/get_by_id",
			Inner:      true,
			Handler:    c.User.GetById,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/user/change_password",
			Inner:      true,
			Handler:    c.User.ChangePassword,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/role/all",
			Inner:      true,
			Extra:      cluster.RequireAdminPermission("role_view"),
			Handler:    c.Role.All,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/user/get_permissions",
			Inner:      true,
			Extra:      cluster.RequireAdminPermission("role_view"),
			Handler:    c.Permissions.GetPermissions,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/role/create",
			Inner:      true,
			Extra:      cluster.RequireAdminPermission("role_add"),
			Handler:    c.Role.CreateRole,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/role/update",
			Inner:      true,
			Extra:      cluster.RequireAdminPermission("role_update"),
			Handler:    c.Role.UpdateRole,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/role/delete",
			Inner:      true,
			Extra:      cluster.RequireAdminPermission("role_delete"),
			Handler:    c.Role.DeleteRole,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/session/all",
			Inner:      true,
			Extra:      cluster.RequireAdminPermission("session_view"),
			Handler:    c.Session.All,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/session/revoke",
			Inner:      true,
			Extra:      cluster.RequireAdminPermission("session_revoke"),
			Handler:    c.Session.Revoke,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/log/all",
			Inner:      true,
			Extra:      cluster.RequireAdminPermission("security_log_view"),
			Handler:    c.Audit.All,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/log/events",
			Inner:      true,
			Extra:      cluster.RequireAdminPermission("audit_management_view"),
			Handler:    c.Audit.Events,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/log/set_events",
			Inner:      true,
			Extra:      cluster.RequireAdminPermission("audit_management_view"),
			Handler:    c.Audit.SetEvents,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/secure/authenticate",
			Inner:      true,
			Handler:    c.Secure.Authenticate,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/secure/authorize",
			Inner:      true,
			Handler:    c.Secure.Authorize,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "admin/ui_settings/oauth",
			Inner:      false,
			Handler:    c.UiSettings.OauthSettings,
			HttpMethod: http.MethodPost,
		},
		// download
		{
			Path:       "/user/download",
			Inner:      true,
			Extra:      cluster.RequireAdminPermission("data_download"),
			Handler:    c.Download.Users,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "/session/download",
			Inner:      true,
			Extra:      cluster.RequireAdminPermission("data_download"),
			Handler:    c.Download.Sessions,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "/log/download",
			Inner:      true,
			Extra:      cluster.RequireAdminPermission("data_download"),
			Handler:    c.Download.Audit,
			HttpMethod: http.MethodPost,
		},
		{
			Path:       "/role/download",
			Inner:      true,
			Extra:      cluster.RequireAdminPermission("data_download"),
			Handler:    c.Download.Roles,
			HttpMethod: http.MethodPost,
		},
	}
}

func commonEndpoints() []cluster.EndpointDescriptor {
	return common_endpoints.CommonEndpoints(
		"isp-admin-service",
		common_endpoints.WithSwaggerEndpoint(docs.SwaggerJson),
	)
}

func concatEndpoints(endpoints ...[]cluster.EndpointDescriptor) []cluster.EndpointDescriptor {
	var result []cluster.EndpointDescriptor
	for _, desc := range endpoints {
		result = append(result, desc...)
	}
	return result
}
