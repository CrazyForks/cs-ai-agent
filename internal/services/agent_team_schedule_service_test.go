package services_test

import (
	"strings"
	"testing"
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/pkg/dto/request"
	"cs-agent/internal/pkg/enums"
	"cs-agent/internal/services"

	"github.com/glebarez/sqlite"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func TestAgentTeamScheduleServiceFindCalendarSchedulesReturnsIntersectingSchedules(t *testing.T) {
	db := setupAgentTeamScheduleTestDB(t)
	createAgentTeamScheduleTestData(t, db)

	list, err := services.AgentTeamScheduleService.FindCalendarSchedules(request.AgentTeamScheduleCalendarRequest{
		StartAt: "2026-04-27 00:00:00",
		EndAt:   "2026-05-04 00:00:00",
	})
	if err != nil {
		t.Fatalf("FindCalendarSchedules() error = %v", err)
	}

	if len(list) != 3 {
		t.Fatalf("expected 3 intersecting schedules, got %d: %+v", len(list), list)
	}
	gotIDs := make([]int64, 0, len(list))
	for _, item := range list {
		gotIDs = append(gotIDs, item.ID)
	}
	wantIDs := []int64{1, 2, 3}
	for i, want := range wantIDs {
		if gotIDs[i] != want {
			t.Fatalf("expected ids %v, got %v", wantIDs, gotIDs)
		}
	}
}

func TestAgentTeamScheduleServiceFindCalendarSchedulesFiltersTeamID(t *testing.T) {
	db := setupAgentTeamScheduleTestDB(t)
	createAgentTeamScheduleTestData(t, db)

	list, err := services.AgentTeamScheduleService.FindCalendarSchedules(request.AgentTeamScheduleCalendarRequest{
		StartAt: "2026-04-27 00:00:00",
		EndAt:   "2026-05-04 00:00:00",
		TeamID:  2,
	})
	if err != nil {
		t.Fatalf("FindCalendarSchedules() error = %v", err)
	}

	if len(list) != 1 {
		t.Fatalf("expected 1 schedule for team 2, got %d: %+v", len(list), list)
	}
	if list[0].ID != 3 || list[0].TeamID != 2 {
		t.Fatalf("unexpected schedule: %+v", list[0])
	}
}

func TestAgentTeamScheduleServiceFindCalendarSchedulesValidatesTimeRange(t *testing.T) {
	setupAgentTeamScheduleTestDB(t)

	_, err := services.AgentTeamScheduleService.FindCalendarSchedules(request.AgentTeamScheduleCalendarRequest{
		StartAt: "2026-05-04 00:00:00",
		EndAt:   "2026-04-27 00:00:00",
	})
	if err == nil {
		t.Fatalf("expected invalid time range to fail")
	}
}

func setupAgentTeamScheduleTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbName := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open("file:"+dbName+"?mode=memory&cache=shared"), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "t_",
			SingularTable: true,
		},
	})
	if err != nil {
		t.Fatalf("open sqlite error = %v", err)
	}
	t.Cleanup(func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	})
	if err := db.AutoMigrate(&models.AgentTeam{}, &models.AgentTeamSchedule{}); err != nil {
		t.Fatalf("auto migrate error = %v", err)
	}
	sqls.SetDB(db)
	return db
}

func createAgentTeamScheduleTestData(t *testing.T, db *gorm.DB) {
	t.Helper()

	teams := []models.AgentTeam{
		{ID: 1, Name: "售前组", Status: enums.StatusOk},
		{ID: 2, Name: "售后组", Status: enums.StatusOk},
	}
	if err := db.Create(&teams).Error; err != nil {
		t.Fatalf("create teams error = %v", err)
	}

	parse := func(value string) time.Time {
		t.Helper()
		ret, err := time.ParseInLocation(time.DateTime, value, time.Local)
		if err != nil {
			t.Fatalf("parse time %q error = %v", value, err)
		}
		return ret
	}
	schedules := []models.AgentTeamSchedule{
		{ID: 1, TeamID: 1, StartAt: parse("2026-04-26 20:00:00"), EndAt: parse("2026-04-27 10:00:00"), SourceType: "manual", Status: enums.StatusOk},
		{ID: 2, TeamID: 1, StartAt: parse("2026-04-28 09:00:00"), EndAt: parse("2026-04-28 18:00:00"), SourceType: "manual", Status: enums.StatusOk},
		{ID: 3, TeamID: 2, StartAt: parse("2026-05-03 20:00:00"), EndAt: parse("2026-05-04 08:00:00"), SourceType: "manual", Status: enums.StatusOk},
		{ID: 4, TeamID: 1, StartAt: parse("2026-04-20 09:00:00"), EndAt: parse("2026-04-20 18:00:00"), SourceType: "manual", Status: enums.StatusOk},
		{ID: 5, TeamID: 2, StartAt: parse("2026-05-04 09:00:00"), EndAt: parse("2026-05-04 18:00:00"), SourceType: "manual", Status: enums.StatusOk},
	}
	if err := db.Create(&schedules).Error; err != nil {
		t.Fatalf("create schedules error = %v", err)
	}
}
