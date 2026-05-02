package services

import (
	"time"

	"cs-agent/internal/models"
	"cs-agent/internal/repositories"

	"github.com/mlogclub/simple/sqls"
	"github.com/mlogclub/simple/web/params"
	"gorm.io/gorm"
)

var TicketNoSequenceService = newTicketNoSequenceService()

func newTicketNoSequenceService() *ticketNoSequenceService {
	return &ticketNoSequenceService{}
}

type ticketNoSequenceService struct {
}

func (s *ticketNoSequenceService) Get(id int64) *models.TicketNoSequence {
	return repositories.TicketNoSequenceRepository.Get(sqls.DB(), id)
}

func (s *ticketNoSequenceService) Take(where ...any) *models.TicketNoSequence {
	return repositories.TicketNoSequenceRepository.Take(sqls.DB(), where...)
}

func (s *ticketNoSequenceService) Find(cnd *sqls.Cnd) []models.TicketNoSequence {
	return repositories.TicketNoSequenceRepository.Find(sqls.DB(), cnd)
}

func (s *ticketNoSequenceService) FindOne(cnd *sqls.Cnd) *models.TicketNoSequence {
	return repositories.TicketNoSequenceRepository.FindOne(sqls.DB(), cnd)
}

func (s *ticketNoSequenceService) FindPageByParams(params *params.QueryParams) (list []models.TicketNoSequence, paging *sqls.Paging) {
	return repositories.TicketNoSequenceRepository.FindPageByParams(sqls.DB(), params)
}

func (s *ticketNoSequenceService) FindPageByCnd(cnd *sqls.Cnd) (list []models.TicketNoSequence, paging *sqls.Paging) {
	return repositories.TicketNoSequenceRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *ticketNoSequenceService) Count(cnd *sqls.Cnd) int64 {
	return repositories.TicketNoSequenceRepository.Count(sqls.DB(), cnd)
}

func (s *ticketNoSequenceService) Create(t *models.TicketNoSequence) error {
	return repositories.TicketNoSequenceRepository.Create(sqls.DB(), t)
}

func (s *ticketNoSequenceService) Update(t *models.TicketNoSequence) error {
	return repositories.TicketNoSequenceRepository.Update(sqls.DB(), t)
}

func (s *ticketNoSequenceService) Updates(id int64, columns map[string]any) error {
	return repositories.TicketNoSequenceRepository.Updates(sqls.DB(), id, columns)
}

func (s *ticketNoSequenceService) UpdateColumn(id int64, name string, value any) error {
	return repositories.TicketNoSequenceRepository.UpdateColumn(sqls.DB(), id, name, value)
}

func (s *ticketNoSequenceService) Delete(id int64) {
	repositories.TicketNoSequenceRepository.Delete(sqls.DB(), id)
}

func (s *ticketNoSequenceService) Next(db *gorm.DB, now time.Time) (string, error) {
	return TicketNoService.Next(db, now)
}
