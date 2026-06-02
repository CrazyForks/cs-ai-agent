package services

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"agent-desk/internal/ai/rag"
	"agent-desk/internal/models"
	"agent-desk/internal/pkg/dto"
	"agent-desk/internal/pkg/dto/request"
	"agent-desk/internal/pkg/enums"
	"agent-desk/internal/pkg/errorsx"
	"agent-desk/internal/pkg/utils"
	"agent-desk/internal/repositories"

	"agent-desk/internal/pkg/httpx/params"

	"github.com/mlogclub/simple/sqls"
)

var KnowledgeFAQService = newKnowledgeFAQService()

func newKnowledgeFAQService() *knowledgeFAQService {
	return &knowledgeFAQService{}
}

type knowledgeFAQService struct{}

func (s *knowledgeFAQService) Get(id int64) *models.KnowledgeFAQ {
	return repositories.KnowledgeFAQRepository.Get(sqls.DB(), id)
}

func (s *knowledgeFAQService) FindPageByCnd(cnd *sqls.Cnd) (list []models.KnowledgeFAQ, paging *sqls.Paging) {
	return repositories.KnowledgeFAQRepository.FindPageByCnd(sqls.DB(), cnd)
}

func (s *knowledgeFAQService) FindPageByParams(queryParams *params.QueryParams) (list []models.KnowledgeFAQ, paging *sqls.Paging) {
	return repositories.KnowledgeFAQRepository.FindPageByParams(sqls.DB(), queryParams)
}

func (s *knowledgeFAQService) Count(cnd *sqls.Cnd) int64 {
	return repositories.KnowledgeFAQRepository.Count(sqls.DB(), cnd)
}

func (s *knowledgeFAQService) CreateKnowledgeFAQ(req request.CreateKnowledgeFAQRequest, operator *dto.AuthPrincipal) (*models.KnowledgeFAQ, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	kb, err := s.requireFAQKnowledgeBase(req.KnowledgeBaseID)
	if err != nil {
		return nil, err
	}
	if _, err := KnowledgeDirectoryService.RequireUsableDirectory(req.KnowledgeBaseID, req.DirectoryID); err != nil {
		return nil, err
	}
	item, err := s.buildKnowledgeFAQModel(req)
	if err != nil {
		return nil, err
	}
	item.Status = kb.Status
	item.IndexStatus = enums.KnowledgeDocumentIndexStatusPending
	item.IndexError = ""
	item.IndexedAt = nil
	item.AuditFields = utils.BuildAuditFields(operator)
	if err := repositories.KnowledgeFAQRepository.Create(sqls.DB(), item); err != nil {
		return nil, err
	}
	if err := rag.Index.IndexFAQByID(context.Background(), item.ID); err != nil {
		slog.Error("failed to index created knowledge faq", "faq_id", item.ID, "error", err)
		return item, nil
	}
	return s.Get(item.ID), nil
}

func (s *knowledgeFAQService) UpdateKnowledgeFAQ(req request.UpdateKnowledgeFAQRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	current := s.Get(req.ID)
	if current == nil {
		return errorsx.InvalidParam("FAQ不存在")
	}
	if _, err := s.requireFAQKnowledgeBase(req.KnowledgeBaseID); err != nil {
		return err
	}
	if _, err := KnowledgeDirectoryService.RequireUsableDirectory(req.KnowledgeBaseID, req.DirectoryID); err != nil {
		return err
	}
	item, err := s.buildKnowledgeFAQModel(req.CreateKnowledgeFAQRequest)
	if err != nil {
		return err
	}
	if err := repositories.KnowledgeFAQRepository.Updates(sqls.DB(), req.ID, map[string]any{
		"knowledge_base_id": item.KnowledgeBaseID,
		"directory_id":      item.DirectoryID,
		"question":          item.Question,
		"answer":            item.Answer,
		"similar_questions": item.SimilarQuestions,
		"index_status":      enums.KnowledgeDocumentIndexStatusPending,
		"indexed_at":        nil,
		"index_error":       "",
		"remark":            item.Remark,
		"update_user_id":    operator.UserID,
		"update_user_name":  operator.Username,
		"updated_at":        time.Now(),
	}); err != nil {
		return err
	}
	return rag.Index.IndexFAQByID(context.Background(), req.ID)
}

func (s *knowledgeFAQService) DeleteKnowledgeFAQ(id int64) error {
	current := s.Get(id)
	if current == nil {
		return errorsx.InvalidParam("FAQ不存在")
	}
	if err := repositories.KnowledgeFAQRepository.Delete(sqls.DB(), id); err != nil {
		return err
	}
	return rag.Index.RemoveFAQIndex(context.Background(), id)
}

func (s *knowledgeFAQService) BatchMoveKnowledgeFAQs(req request.BatchMoveKnowledgeFAQRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	ids := uniquePositiveIDs(req.IDs)
	if len(ids) == 0 {
		return errorsx.InvalidParam("请选择要移动的FAQ")
	}
	if _, err := s.requireFAQKnowledgeBase(req.KnowledgeBaseID); err != nil {
		return err
	}
	if _, err := KnowledgeDirectoryService.RequireUsableDirectory(req.KnowledgeBaseID, req.DirectoryID); err != nil {
		return err
	}
	for _, id := range ids {
		current := s.Get(id)
		if current == nil {
			return errorsx.InvalidParam("FAQ不存在")
		}
		if current.KnowledgeBaseID != req.KnowledgeBaseID {
			return errorsx.InvalidParam("只能移动当前知识库下的FAQ")
		}
	}
	now := time.Now()
	if err := sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		for _, id := range ids {
			if err := repositories.KnowledgeFAQRepository.Updates(ctx.Tx, id, map[string]any{
				"directory_id":     req.DirectoryID,
				"index_status":     enums.KnowledgeDocumentIndexStatusPending,
				"indexed_at":       nil,
				"index_error":      "",
				"update_user_id":   operator.UserID,
				"update_user_name": operator.Username,
				"updated_at":       now,
			}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	for _, id := range ids {
		if err := rag.Index.IndexFAQByID(context.Background(), id); err != nil {
			slog.Error("failed to reindex moved knowledge faq", "faq_id", id, "error", err)
		}
	}
	return nil
}

func (s *knowledgeFAQService) BatchDeleteKnowledgeFAQs(req request.BatchDeleteKnowledgeFAQRequest) error {
	ids := uniquePositiveIDs(req.IDs)
	if len(ids) == 0 {
		return errorsx.InvalidParam("请选择要删除的FAQ")
	}
	for _, id := range ids {
		if err := s.DeleteKnowledgeFAQ(id); err != nil {
			return err
		}
	}
	return nil
}

func (s *knowledgeFAQService) buildKnowledgeFAQModel(req request.CreateKnowledgeFAQRequest) (*models.KnowledgeFAQ, error) {
	if req.KnowledgeBaseID <= 0 {
		return nil, errorsx.InvalidParam("知识库不存在")
	}
	if req.Question == "" {
		return nil, errorsx.InvalidParam("问题不能为空")
	}
	if req.Answer == "" {
		return nil, errorsx.InvalidParam("答案不能为空")
	}
	similarQuestions, err := json.Marshal(normalizeSimilarQuestions(req.SimilarQuestions))
	if err != nil {
		return nil, errorsx.InvalidParam("相似问格式不合法")
	}
	return &models.KnowledgeFAQ{
		KnowledgeBaseID:  req.KnowledgeBaseID,
		DirectoryID:      req.DirectoryID,
		Question:         req.Question,
		Answer:           req.Answer,
		SimilarQuestions: string(similarQuestions),
		Remark:           req.Remark,
	}, nil
}

func (s *knowledgeFAQService) requireFAQKnowledgeBase(knowledgeBaseID int64) (*models.KnowledgeBase, error) {
	kb := KnowledgeBaseService.Get(knowledgeBaseID)
	if kb == nil {
		return nil, errorsx.InvalidParam("知识库不存在")
	}
	if kb.KnowledgeType != "faq" {
		return nil, errorsx.InvalidParam("当前知识库不是FAQ知识库")
	}
	return kb, nil
}

func normalizeSimilarQuestions(values []string) []string {
	items := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, item := range values {
		value := item
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		items = append(items, value)
	}
	return items
}
