package runtime

func newPrepareService(catalog *toolCatalog) *prepareService {
	return &prepareService{catalog: catalog}
}

type prepareService struct {
	catalog *toolCatalog
}

func (s *prepareService) prepareToolsForRun(req *Request) error {
	if req == nil || req.ToolSet != nil || s.catalog == nil {
		return nil
	}
	toolSet, err := s.catalog.resolveForRun(req)
	if err != nil {
		return err
	}
	if toolSet != nil {
		req.ToolSet = toolSet
	}
	return nil
}

func (s *prepareService) prepareToolsForResume(req *ResumeRequest) error {
	if req == nil || req.ToolSet != nil || s.catalog == nil {
		return nil
	}
	toolSet, err := s.catalog.resolveForResume(req)
	if err != nil {
		return err
	}
	if toolSet != nil {
		req.ToolSet = toolSet
	}
	return nil
}
