package migration

import (
	"cs-agent/internal/models"

	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

func init() {
	register(8, "sync lightweight ticket permissions and reset ticket data", func() error {
		return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
			if err := resetLightweightTicketData(ctx.Tx); err != nil {
				return err
			}

			permissions, err := ensurePermissions(ctx.Tx)
			if err != nil {
				return err
			}

			roles, err := ensureRoles(ctx.Tx)
			if err != nil {
				return err
			}

			return ensureRolePermissions(ctx.Tx, roles, permissions)
		})
	})
}

func resetLightweightTicketData(tx *gorm.DB) error {
	for _, table := range lightweightTicketResetTables(tx) {
		if !tx.Migrator().HasTable(table) {
			continue
		}
		if err := tx.Exec("DELETE FROM " + table).Error; err != nil {
			return err
		}
	}
	return nil
}

func lightweightTicketResetTables(tx *gorm.DB) []string {
	return []string{
		"t_ticket_sla_record",
		"t_ticket_resolution_code",
		"t_ticket_priority_config",
		"t_ticket_watcher",
		"t_ticket_collaborator",
		"t_ticket_mention",
		"t_ticket_event_log",
		"t_ticket_relation",
		"t_ticket_comment",
		tableName(tx, &models.TicketProgress{}, "t_ticket_progress"),
		tableName(tx, &models.TicketTag{}, "t_ticket_tag"),
		tableName(tx, &models.Ticket{}, "t_ticket"),
		tableName(tx, &models.TicketNoSequence{}, "t_ticket_no_sequence"),
		tableName(tx, &models.TicketView{}, "t_ticket_view"),
	}
}

func tableName(tx *gorm.DB, model any, fallback string) string {
	stmt := &gorm.Statement{DB: tx}
	if err := stmt.Parse(model); err != nil {
		return fallback
	}
	return stmt.Schema.Table
}
