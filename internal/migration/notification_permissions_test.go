package migration

import "testing"

func TestNotificationPermissionMigrationRegistered(t *testing.T) {
	migration, ok := migrationFuncs[7]
	if !ok {
		t.Fatalf("expected migration version 7 to be registered")
	}
	if migration.Remark != "sync notification permissions" {
		t.Fatalf("unexpected migration remark: %q", migration.Remark)
	}
}

func TestLightweightTicketPermissionMigrationRegistered(t *testing.T) {
	migration, ok := migrationFuncs[8]
	if !ok {
		t.Fatalf("expected migration version 8 to be registered")
	}
	if migration.Remark != "sync lightweight ticket permissions" {
		t.Fatalf("unexpected migration remark: %q", migration.Remark)
	}
}
