package enums

type AIProvider string

const (
	AIProviderOpenAI AIProvider = "openai"
)

var aiProviderLabelMap = map[AIProvider]string{
	AIProviderOpenAI: "OpenAI",
}

func GetAIProviderLabel(provider AIProvider) string {
	return aiProviderLabelMap[provider]
}

type AIModelType string

const (
	AIModelTypeLLM       AIModelType = "llm"
	AIModelTypeEmbedding AIModelType = "embedding"
	AIModelTypeRerank    AIModelType = "rerank"
)

var aiModelTypeLabelMap = map[AIModelType]string{
	AIModelTypeLLM:       "大语言模型",
	AIModelTypeEmbedding: "向量模型",
	AIModelTypeRerank:    "重排序模型",
}

func GetAIModelTypeLabel(modelType AIModelType) string {
	return aiModelTypeLabelMap[modelType]
}

type AIAgentRuntimeMode int

const (
	AIAgentRuntimeModeBuiltinGraph AIAgentRuntimeMode = 1
	AIAgentRuntimeModeWorkflow     AIAgentRuntimeMode = 2
)

var AIAgentRuntimeModeValues = []AIAgentRuntimeMode{
	AIAgentRuntimeModeBuiltinGraph,
	AIAgentRuntimeModeWorkflow,
}

var aiAgentRuntimeModeLabelMap = map[AIAgentRuntimeMode]string{
	AIAgentRuntimeModeBuiltinGraph: "内置 Graph",
	AIAgentRuntimeModeWorkflow:     "会话流程",
}

func GetAIAgentRuntimeModeLabel(mode AIAgentRuntimeMode) string {
	return aiAgentRuntimeModeLabelMap[mode]
}
