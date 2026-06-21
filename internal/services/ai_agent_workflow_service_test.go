package services

import (
	"testing"

	"agent-desk/internal/models"
	"agent-desk/internal/pkg/dto"
	"agent-desk/internal/pkg/dto/request"
	"agent-desk/internal/pkg/enums"

	"github.com/glebarez/sqlite"
	"github.com/mlogclub/simple/sqls"
	"gorm.io/gorm"
)

func TestAIAgentServiceSavesWorkflowBinding(t *testing.T) {
	setupAIAgentWorkflowTestDB(t)
	operator := aiAgentWorkflowTestOperator()
	aiConfigID := createAIAgentWorkflowTestConfig(t)
	knowledgeID := createAIAgentWorkflowTestKnowledgeBase(t)
	versionID := createAIAgentWorkflowVersion(t)

	item, err := AIAgentService.CreateAIAgent(request.CreateAIAgentRequest{
		Name:              "workflow agent",
		AIConfigID:        aiConfigID,
		ServiceMode:       enums.IMConversationServiceModeAIOnly,
		HandoffMode:       enums.AIAgentHandoffModeWaitPool,
		FallbackMode:      enums.AIAgentFallbackModeNoAnswer,
		KnowledgeIDs:      []int64{knowledgeID},
		RuntimeMode:       enums.AIAgentRuntimeModeWorkflow,
		WorkflowVersionID: versionID,
	}, operator)
	if err != nil {
		t.Fatalf("CreateAIAgent() error = %v", err)
	}

	if item.RuntimeMode != enums.AIAgentRuntimeModeWorkflow {
		t.Fatalf("expected workflow runtime mode, got %d", item.RuntimeMode)
	}
	if item.WorkflowVersionID != versionID {
		t.Fatalf("expected workflow version %d, got %d", versionID, item.WorkflowVersionID)
	}
}

func TestAIAgentServiceRejectsWorkflowModeWithoutVersion(t *testing.T) {
	setupAIAgentWorkflowTestDB(t)
	operator := aiAgentWorkflowTestOperator()
	aiConfigID := createAIAgentWorkflowTestConfig(t)
	knowledgeID := createAIAgentWorkflowTestKnowledgeBase(t)

	_, err := AIAgentService.CreateAIAgent(request.CreateAIAgentRequest{
		Name:         "workflow agent without version",
		AIConfigID:   aiConfigID,
		ServiceMode:  enums.IMConversationServiceModeAIOnly,
		HandoffMode:  enums.AIAgentHandoffModeWaitPool,
		FallbackMode: enums.AIAgentFallbackModeNoAnswer,
		KnowledgeIDs: []int64{knowledgeID},
		RuntimeMode:  enums.AIAgentRuntimeModeWorkflow,
	}, operator)
	if err == nil {
		t.Fatalf("expected workflow runtime without version to fail")
	}
}

func setupAIAgentWorkflowTestDB(t *testing.T) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:"+t.Name()+"?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite db: %v", err)
	}
	if err := db.AutoMigrate(&models.AIAgent{}, &models.AIConfig{}, &models.KnowledgeBase{}, &models.AIWorkflow{}, &models.AIWorkflowVersion{}); err != nil {
		t.Fatalf("auto migrate: %v", err)
	}
	sqls.SetDB(db)
}

func createAIAgentWorkflowTestConfig(t *testing.T) int64 {
	t.Helper()
	item := &models.AIConfig{
		Name:      "workflow-test-config",
		Provider:  enums.AIProviderOpenAI,
		ModelType: enums.AIModelTypeLLM,
		ModelName: "gpt-test",
		Status:    enums.StatusOk,
	}
	if err := sqls.DB().Create(item).Error; err != nil {
		t.Fatalf("create ai config: %v", err)
	}
	return item.ID
}

func createAIAgentWorkflowTestKnowledgeBase(t *testing.T) int64 {
	t.Helper()
	item := &models.KnowledgeBase{
		Name:          "workflow-test-kb",
		KnowledgeType: string(enums.KnowledgeBaseTypeFAQ),
		Status:        enums.StatusOk,
	}
	if err := sqls.DB().Create(item).Error; err != nil {
		t.Fatalf("create knowledge base: %v", err)
	}
	return item.ID
}

func createAIAgentWorkflowVersion(t *testing.T) int64 {
	t.Helper()
	workflow := &models.AIWorkflow{
		Name:      "workflow-test",
		OwnerType: "ai_agent",
		OwnerID:   1,
		Status:    enums.StatusOk,
	}
	if err := sqls.DB().Create(workflow).Error; err != nil {
		t.Fatalf("create workflow: %v", err)
	}
	version := &models.AIWorkflowVersion{
		WorkflowID: workflow.ID,
		Version:    1,
		Status:     enums.StatusOk,
	}
	if err := sqls.DB().Create(version).Error; err != nil {
		t.Fatalf("create workflow version: %v", err)
	}
	return version.ID
}

func aiAgentWorkflowTestOperator() *dto.AuthPrincipal {
	return &dto.AuthPrincipal{
		UserID:   1,
		Username: "agent-workflow-tester",
		Nickname: "agent-workflow-tester",
	}
}
