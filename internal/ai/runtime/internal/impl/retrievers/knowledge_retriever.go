package retrievers

import (
	"context"
	"strings"

	"cs-agent/internal/ai/rag"
	"cs-agent/internal/ai/runtime/internal/impl/callbacks"
	"cs-agent/internal/models"
	"cs-agent/internal/pkg/utils"
)

const defaultRuntimeKnowledgeContextTokens = 4000

type KnowledgeRetriever struct {
	AIAgent *models.AIAgent
}

type KnowledgeRetrieveOptions struct {
	ContextMaxTokens int
	TopK             int
	ScoreThreshold   float64
	QueryPreview     string
}

type KnowledgeRetrieveResult struct {
	KnowledgeBaseIDs []int64
	Query            string
	Options          KnowledgeRetrieveOptions
	Hits             []rag.RetrieveResult
	ContextResults   []rag.RetrieveResult
	ContextText      string
	Trace            *rag.RetrieveTrace
	TraceItems       []callbacks.RetrieverTraceItem
	TraceSummary     callbacks.RetrieverTraceSummary
}

func NewKnowledgeRetriever(aiAgent *models.AIAgent) *KnowledgeRetriever {
	return &KnowledgeRetriever{AIAgent: aiAgent}
}

func (r *KnowledgeRetriever) KnowledgeBaseIDs() []int64 {
	if r == nil || r.AIAgent == nil {
		return nil
	}
	return utils.SplitInt64s(r.AIAgent.KnowledgeIDs)
}

func (r *KnowledgeRetriever) Retrieve(ctx context.Context, query string) ([]rag.RetrieveResult, *rag.RetrieveTrace, error) {
	return r.RetrieveByOptions(ctx, KnowledgeRetrieveOptions{}, query)
}

func (r *KnowledgeRetriever) RetrieveByOptions(ctx context.Context, opts KnowledgeRetrieveOptions, query string) ([]rag.RetrieveResult, *rag.RetrieveTrace, error) {
	ids := r.KnowledgeBaseIDs()
	return rag.Retrieve.RetrieveWithTrace(ctx, rag.RetrieveRequest{
		Query:            query,
		KnowledgeBaseIDs: ids,
		TopK:             opts.TopK,
		ScoreThreshold:   opts.ScoreThreshold,
	})
}

func (r *KnowledgeRetriever) RetrieveContext(ctx context.Context, query string) (*KnowledgeRetrieveResult, error) {
	return r.RetrieveContextByOptions(ctx, KnowledgeRetrieveOptions{}, query)
}

func (r *KnowledgeRetriever) RetrieveContextByOptions(ctx context.Context, opts KnowledgeRetrieveOptions, query string) (*KnowledgeRetrieveResult, error) {
	query = strings.TrimSpace(query)
	knowledgeBaseIDs := r.KnowledgeBaseIDs()
	contextMaxTokens := opts.ContextMaxTokens
	if contextMaxTokens <= 0 {
		contextMaxTokens = defaultRuntimeKnowledgeContextTokens
	}
	queryPreview := strings.TrimSpace(opts.QueryPreview)
	if queryPreview == "" {
		queryPreview = query
	}
	ret := &KnowledgeRetrieveResult{
		KnowledgeBaseIDs: append([]int64(nil), knowledgeBaseIDs...),
		Query:            query,
		Options: KnowledgeRetrieveOptions{
			ContextMaxTokens: contextMaxTokens,
			TopK:             opts.TopK,
			ScoreThreshold:   opts.ScoreThreshold,
			QueryPreview:     queryPreview,
		},
	}
	if query == "" || len(knowledgeBaseIDs) == 0 {
		return ret, nil
	}
	results, trace, err := r.RetrieveByOptions(ctx, opts, query)
	if err != nil {
		return nil, err
	}
	ret.Hits = append([]rag.RetrieveResult(nil), results...)
	ret.Trace = trace
	ret.ContextResults = rag.Retrieve.SelectContextResults(results, contextMaxTokens)
	ret.ContextText = strings.TrimSpace(rag.Retrieve.BuildContext(ctx, results, contextMaxTokens))
	ret.TraceItems = buildRetrieverTraceItems(queryPreview, results, trace)
	ret.TraceSummary = buildRetrieverTraceSummary(ret.Options, ret.ContextResults, results, trace)
	return ret, nil
}

func buildRetrieverTraceItems(queryPreview string, results []rag.RetrieveResult, trace *rag.RetrieveTrace) []callbacks.RetrieverTraceItem {
	if len(results) == 0 {
		return nil
	}
	latencyMs := int64(0)
	if trace != nil {
		latencyMs = trace.EmbeddingMs + trace.VectorSearchMs + trace.HydrateMs
	}
	ret := make([]callbacks.RetrieverTraceItem, 0, len(results))
	for _, item := range results {
		ret = append(ret, callbacks.RetrieverTraceItem{
			Query:           queryPreview,
			KnowledgeBaseID: item.KnowledgeBaseID,
			DocumentID:      item.DocumentID,
			DocumentTitle:   item.DocumentTitle,
			Score:           float64(item.Score),
			LatencyMs:       latencyMs,
		})
	}
	return ret
}

func buildRetrieverTraceSummary(opts KnowledgeRetrieveOptions, contextResults []rag.RetrieveResult, results []rag.RetrieveResult, trace *rag.RetrieveTrace) callbacks.RetrieverTraceSummary {
	ret := callbacks.RetrieverTraceSummary{
		TopK:             opts.TopK,
		ScoreThreshold:   opts.ScoreThreshold,
		ContextMaxTokens: opts.ContextMaxTokens,
		HitCount:         len(results),
		ContextCount:     len(contextResults),
	}
	if trace != nil {
		ret.EmbeddingMs = trace.EmbeddingMs
		ret.VectorSearchMs = trace.VectorSearchMs
		ret.HydrateMs = trace.HydrateMs
	}
	return ret
}
