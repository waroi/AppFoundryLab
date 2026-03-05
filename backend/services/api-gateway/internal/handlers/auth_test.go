package handlers

import "testing"

func TestResolveRoleDemoModeAllowsDefaults(t *testing.T) {
	t.Setenv("LOCAL_AUTH_MODE", "demo")
	t.Setenv("BOOTSTRAP_ADMIN_USER", "admin")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "admin_dev_password")
	t.Setenv("BOOTSTRAP_USER", "developer")
	t.Setenv("BOOTSTRAP_USER_PASSWORD", "developer_dev_password")

	role, ok := resolveRole("developer", "developer_dev_password")
	if !ok || role != "user" {
		t.Fatalf("expected demo credentials to work, got ok=%v role=%s", ok, role)
	}
}

func TestResolveRoleGeneratedModeRejectsDefaults(t *testing.T) {
	t.Setenv("LOCAL_AUTH_MODE", "generated")
	t.Setenv("BOOTSTRAP_ADMIN_USER", "admin")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "admin_dev_password")
	t.Setenv("BOOTSTRAP_USER", "developer")
	t.Setenv("BOOTSTRAP_USER_PASSWORD", "developer_dev_password")

	if role, ok := resolveRole("developer", "developer_dev_password"); ok || role != "" {
		t.Fatalf("expected generated mode to reject default credentials, got ok=%v role=%s", ok, role)
	}
}

func TestResolveRoleGeneratedModeAllowsCustomCredentials(t *testing.T) {
	t.Setenv("LOCAL_AUTH_MODE", "generated")
	t.Setenv("BOOTSTRAP_ADMIN_USER", "admin")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "local-admin-secret")
	t.Setenv("BOOTSTRAP_USER", "developer")
	t.Setenv("BOOTSTRAP_USER_PASSWORD", "local-user-secret")

	role, ok := resolveRole("developer", "local-user-secret")
	if !ok || role != "user" {
		t.Fatalf("expected generated mode custom credentials to work, got ok=%v role=%s", ok, role)
	}
}

func TestResolveRoleDisabledModeRejectsBootstrapAuth(t *testing.T) {
	t.Setenv("LOCAL_AUTH_MODE", "disabled")
	t.Setenv("BOOTSTRAP_ADMIN_USER", "admin")
	t.Setenv("BOOTSTRAP_ADMIN_PASSWORD", "local-admin-secret")

	if role, ok := resolveRole("admin", "local-admin-secret"); ok || role != "" {
		t.Fatalf("expected disabled mode to reject bootstrap auth, got ok=%v role=%s", ok, role)
	}
}
