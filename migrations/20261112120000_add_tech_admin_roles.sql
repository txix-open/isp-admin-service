-- +goose Up
INSERT INTO roles (name, permissions, external_group, immutable, exclusive, created_at, updated_at)
VALUES ('DIT_KKD_TECH_ADMIN_PROD',
        '["profile_view", "profile_change_password", "application_group_view", "application_group_add", "application_group_edit", "application_group_delete", "application_group_app_add", "application_group_app_edit", "application_group_app_delete", "application_group_token_view", "application_group_token_add", "application_group_token_delete", "app_access_view", "app_access_edit", "variable_view", "variable_add", "variable_edit", "variable_delete", "module_view", "module_delete", "module_configuration_save_unsafe", "module_configuration_set_active", "module_history_set", "module_history_delete_version", "module_configuration_edit", "module_configuration_add"]'::jsonb,
        'cn=sudir_prod_int_kkd_tech_admin,cn=groups,cn=accounts,dc=data,dc=corp', true, true, NOW(), NOW())
ON CONFLICT (name) DO UPDATE
SET permissions = EXCLUDED.permissions;

INSERT INTO roles (name, permissions, external_group, immutable, exclusive, created_at, updated_at)
VALUES ('DIT_KKD_TECH_ADMIN_STAGE',
        '["profile_view", "profile_change_password", "application_group_view", "application_group_add", "application_group_edit", "application_group_delete", "application_group_app_add", "application_group_app_edit", "application_group_app_delete", "application_group_token_view", "application_group_token_add", "application_group_token_delete", "app_access_view", "app_access_edit", "variable_view", "variable_add", "variable_edit", "variable_delete", "module_view", "module_delete", "module_configuration_save_unsafe", "module_configuration_set_active", "module_history_set", "module_history_delete_version", "module_configuration_edit", "module_configuration_add"]'::jsonb,
        'cn=sudir_stage_int_kkd_tech_admin,cn=groups,cn=accounts,dc=data,dc=corp', true, true, NOW(), NOW())
ON CONFLICT (name) DO UPDATE
SET permissions = EXCLUDED.permissions;

INSERT INTO roles (name, permissions, external_group, immutable, exclusive, created_at, updated_at)
VALUES ('DIT_KKD_TECH_ADMIN_IB_PROD',
        '["session_view", "user_view", "user_block", "user_create", "user_update", "user_delete", "session_revoke", "session_delete", "security_log_view", "profile_view", "profile_edit", "role_view", "role_add", "role_update", "role_delete", "role_link_user", "role_link_external_group", "supervisor_view"]'::jsonb,
        'cn=sudir_prod_int_kkd_tech_admin_ib,cn=groups,cn=accounts,dc=data,dc=corp', true, true, NOW(), NOW())
ON CONFLICT (name) DO UPDATE
SET permissions = EXCLUDED.permissions;

INSERT INTO roles (name, permissions, external_group, immutable, exclusive, created_at, updated_at)
VALUES ('DIT_KKD_TECH_ADMIN_IB_STAGE',
        '["session_view", "user_view", "user_block", "user_create", "user_update", "user_delete", "session_revoke", "session_delete", "security_log_view", "profile_view", "profile_edit", "role_view", "role_add", "role_update", "role_delete", "role_link_user", "role_link_external_group", "supervisor_view"]'::jsonb,
        'cn=sudir_stage_int_kkd_tech_admin_ib,cn=groups,cn=accounts,dc=data,dc=corp', true, true, NOW(), NOW())
ON CONFLICT (name) DO UPDATE
SET permissions = EXCLUDED.permissions;

INSERT INTO roles (name, permissions, external_group, immutable, exclusive, created_at, updated_at)
VALUES ('DIT_KKD_TECH_ADMIN_READ_ONLY_PROD',
        '["profile_view", "application_group_view", "application_group_token_view", "app_access_view", "module_view"]'::jsonb,
        'cn=sudir_prod_int_kkd_tech_operator_ro,cn=groups,cn=accounts,dc=data,dc=corp', true, true, NOW(), NOW())
ON CONFLICT (name) DO UPDATE
SET permissions = EXCLUDED.permissions;

INSERT INTO roles (name, permissions, external_group, immutable, exclusive, created_at, updated_at)
VALUES ('DIT_KKD_TECH_ADMIN_READ_ONLY_STAGE',
        '["profile_view", "application_group_view", "application_group_token_view", "app_access_view", "module_view"]'::jsonb,
        'cn=sudir_stage_int_kkd_tech_operator_ro,cn=groups,cn=accounts,dc=data,dc=corp', true, true, NOW(), NOW())
ON CONFLICT (name) DO UPDATE
SET permissions = EXCLUDED.permissions;

-- +goose Down
