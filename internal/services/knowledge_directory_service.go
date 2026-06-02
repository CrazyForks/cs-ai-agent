package services

import (
	"strings"
	"time"

	"agent-desk/internal/models"
	"agent-desk/internal/pkg/dto"
	"agent-desk/internal/pkg/dto/request"
	"agent-desk/internal/pkg/enums"
	"agent-desk/internal/pkg/errorsx"
	"agent-desk/internal/pkg/utils"
	"agent-desk/internal/repositories"

	"github.com/mlogclub/simple/sqls"
)

var KnowledgeDirectoryService = newKnowledgeDirectoryService()

func newKnowledgeDirectoryService() *knowledgeDirectoryService {
	return &knowledgeDirectoryService{}
}

type knowledgeDirectoryService struct {
}

func (s *knowledgeDirectoryService) Get(id int64) *models.KnowledgeDirectory {
	return repositories.KnowledgeDirectoryRepository.Get(sqls.DB(), id)
}

func (s *knowledgeDirectoryService) Find(cnd *sqls.Cnd) []models.KnowledgeDirectory {
	return repositories.KnowledgeDirectoryRepository.Find(sqls.DB(), cnd)
}

func (s *knowledgeDirectoryService) FindAllByKnowledgeBaseID(knowledgeBaseID int64) []models.KnowledgeDirectory {
	return s.Find(sqls.NewCnd().Eq("knowledge_base_id", knowledgeBaseID).Asc("parent_id").Asc("sort_no").Asc("id"))
}

func (s *knowledgeDirectoryService) Count(cnd *sqls.Cnd) int64 {
	return repositories.KnowledgeDirectoryRepository.Count(sqls.DB(), cnd)
}

func (s *knowledgeDirectoryService) CreateDirectory(req request.CreateKnowledgeDirectoryRequest, operator *dto.AuthPrincipal) (*models.KnowledgeDirectory, error) {
	if operator == nil {
		return nil, errorsx.Unauthorized("未登录或登录已过期")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errorsx.InvalidParam("目录名称不能为空")
	}
	if req.KnowledgeBaseID <= 0 || KnowledgeBaseService.Get(req.KnowledgeBaseID) == nil {
		return nil, errorsx.InvalidParam("知识库不存在")
	}
	if err := s.validateParent(req.KnowledgeBaseID, req.ParentID, 0); err != nil {
		return nil, err
	}
	if existing := s.findByName(req.KnowledgeBaseID, req.ParentID, name); existing != nil {
		return nil, errorsx.InvalidParam("同级下已存在相同名称的目录")
	}
	item := &models.KnowledgeDirectory{
		KnowledgeBaseID: req.KnowledgeBaseID,
		ParentID:        req.ParentID,
		Name:            name,
		SortNo:          s.NextSortNo(req.KnowledgeBaseID, req.ParentID),
		Status:          enums.StatusOk,
		Remark:          strings.TrimSpace(req.Remark),
		AuditFields:     utils.BuildAuditFields(operator),
	}
	if err := repositories.KnowledgeDirectoryRepository.Create(sqls.DB(), item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *knowledgeDirectoryService) UpdateDirectory(req request.UpdateKnowledgeDirectoryRequest, operator *dto.AuthPrincipal) error {
	if operator == nil {
		return errorsx.Unauthorized("未登录或登录已过期")
	}
	item := s.Get(req.ID)
	if item == nil {
		return errorsx.InvalidParam("目录不存在")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errorsx.InvalidParam("目录名称不能为空")
	}
	if req.KnowledgeBaseID <= 0 {
		req.KnowledgeBaseID = item.KnowledgeBaseID
	}
	if req.KnowledgeBaseID != item.KnowledgeBaseID {
		return errorsx.InvalidParam("目录不能移动到其他知识库")
	}
	if err := s.validateParent(item.KnowledgeBaseID, req.ParentID, req.ID); err != nil {
		return err
	}
	if req.ParentID > 0 && s.Count(sqls.NewCnd().Eq("parent_id", req.ID)) > 0 {
		return errorsx.InvalidParam("存在子目录的目录不能移动到二级目录")
	}
	if existing := s.findByName(item.KnowledgeBaseID, req.ParentID, name); existing != nil && existing.ID != req.ID {
		return errorsx.InvalidParam("同级下已存在相同名称的目录")
	}
	return repositories.KnowledgeDirectoryRepository.Updates(sqls.DB(), req.ID, map[string]any{
		"parent_id":        req.ParentID,
		"name":             name,
		"remark":           strings.TrimSpace(req.Remark),
		"update_user_id":   operator.UserID,
		"update_user_name": operator.Username,
		"updated_at":       time.Now(),
	})
}

func (s *knowledgeDirectoryService) DeleteDirectory(id int64) error {
	item := s.Get(id)
	if item == nil {
		return errorsx.InvalidParam("目录不存在")
	}
	if s.Count(sqls.NewCnd().Eq("parent_id", id)) > 0 {
		return errorsx.InvalidParam("该目录下存在子目录，无法删除")
	}
	if KnowledgeDocumentService.Count(sqls.NewCnd().Eq("directory_id", id)) > 0 {
		return errorsx.InvalidParam("该目录下存在文档，无法删除")
	}
	if KnowledgeFAQService.Count(sqls.NewCnd().Eq("directory_id", id)) > 0 {
		return errorsx.InvalidParam("该目录下存在FAQ，无法删除")
	}
	return repositories.KnowledgeDirectoryRepository.Delete(sqls.DB(), id)
}

func (s *knowledgeDirectoryService) UpdateSort(knowledgeBaseID int64, parentID int64, ids []int64) error {
	return sqls.WithTransaction(func(ctx *sqls.TxContext) error {
		for i, id := range ids {
			item := repositories.KnowledgeDirectoryRepository.Get(ctx.Tx, id)
			if item == nil {
				return errorsx.InvalidParam("目录不存在")
			}
			if item.KnowledgeBaseID != knowledgeBaseID || item.ParentID != parentID {
				return errorsx.InvalidParam("只能调整同知识库同级目录排序")
			}
			if err := repositories.KnowledgeDirectoryRepository.UpdateColumn(ctx.Tx, id, "sort_no", i+1); err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *knowledgeDirectoryService) RequireUsableDirectory(knowledgeBaseID int64, directoryID int64) (*models.KnowledgeDirectory, error) {
	if directoryID <= 0 {
		return nil, nil
	}
	item := s.Get(directoryID)
	if item == nil {
		return nil, errorsx.InvalidParam("知识库目录不存在")
	}
	if item.KnowledgeBaseID != knowledgeBaseID {
		return nil, errorsx.InvalidParam("知识库目录不属于当前知识库")
	}
	if item.Status != enums.StatusOk {
		return nil, errorsx.InvalidParam("知识库目录不可用")
	}
	return item, nil
}

func (s *knowledgeDirectoryService) NextSortNo(knowledgeBaseID int64, parentID int64) int {
	if temp := repositories.KnowledgeDirectoryRepository.FindOne(sqls.DB(), sqls.NewCnd().Eq("knowledge_base_id", knowledgeBaseID).Eq("parent_id", parentID).Desc("sort_no").Desc("id")); temp != nil {
		return temp.SortNo + 1
	}
	return 1
}

func (s *knowledgeDirectoryService) PathMap(knowledgeBaseID int64) map[int64]string {
	items := s.FindAllByKnowledgeBaseID(knowledgeBaseID)
	byID := make(map[int64]models.KnowledgeDirectory, len(items))
	for _, item := range items {
		byID[item.ID] = item
	}
	ret := make(map[int64]string, len(items))
	for _, item := range items {
		if item.ParentID > 0 {
			if parent, ok := byID[item.ParentID]; ok {
				ret[item.ID] = parent.Name + " / " + item.Name
				continue
			}
		}
		ret[item.ID] = item.Name
	}
	return ret
}

func (s *knowledgeDirectoryService) validateParent(knowledgeBaseID int64, parentID int64, selfID int64) error {
	if parentID <= 0 {
		return nil
	}
	if parentID == selfID {
		return errorsx.InvalidParam("不能将目录设为自己的子目录")
	}
	parent := s.Get(parentID)
	if parent == nil {
		return errorsx.InvalidParam("父目录不存在")
	}
	if parent.KnowledgeBaseID != knowledgeBaseID {
		return errorsx.InvalidParam("父目录不属于当前知识库")
	}
	if parent.ParentID > 0 {
		return errorsx.InvalidParam("知识库目录最多支持二级")
	}
	return nil
}

func (s *knowledgeDirectoryService) findByName(knowledgeBaseID int64, parentID int64, name string) *models.KnowledgeDirectory {
	return repositories.KnowledgeDirectoryRepository.FindOne(sqls.DB(), sqls.NewCnd().Eq("knowledge_base_id", knowledgeBaseID).Eq("parent_id", parentID).Eq("name", name))
}
