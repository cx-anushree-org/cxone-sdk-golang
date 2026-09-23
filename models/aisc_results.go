package models

// AISCResultsCollection is the response of
// GET /api/ai-sc/reader/scans/{scanId}/results — a paginated list of AI
// supply chain findings for a single scan.
type AISCResultsCollection struct {
	Data        []*AISCResult `json:"data"`
	Total       int           `json:"total"`
	CurrentPage int           `json:"currentPage"`
	LastPage    int           `json:"lastPage"`
}

// AISCResult is one AI supply chain finding — a detected AI asset (model,
// SDK, or framework) at a specific location in source.
type AISCResult struct {
	ID                      string `json:"id"`
	EvidenceKey             string `json:"evidenceKey,omitempty"`
	AssetType               string `json:"assetType"`
	AssetTypeID             string `json:"assetTypeId,omitempty"`
	AssetID                 string `json:"assetId,omitempty"`
	AssetName               string `json:"assetName"`
	Provider                string `json:"provider,omitempty"`
	AssetFirstDetectionDate string `json:"assetFirstDetectionDate,omitempty"`
	State                   string `json:"state,omitempty"`
	Path                    string `json:"path,omitempty"`
	StartLine               int    `json:"startLine,omitempty"`
	StartColumn             int    `json:"startColumn,omitempty"`
	EndLine                 int    `json:"endLine,omitempty"`
	EndColumn               int    `json:"endColumn,omitempty"`
}

// AISCAggregateResponse is the response of
// GET /api/ai-sc/reader/scans/{scanId}/results/aggregate - grouped counts
// of AI supply chain findings for a single scan.
type AISCAggregateResponse struct {
	ScanGroupsCounter []*AISCGroupCounter `json:"scanGroupsCounter"`
}

// AISCGroupCounter is one grouped count entry. Only the fields relevant to
// the requested groupBy are populated
type AISCGroupCounter struct {
	AssetTypeID string `json:"assetTypeId,omitempty"`
	AssetType   string `json:"assetType,omitempty"`
	AssetName   string `json:"assetName,omitempty"`
	Provider    string `json:"provider,omitempty"`
	Version     string `json:"version,omitempty"`
	State       string `json:"state,omitempty"`
	Count       int    `json:"count"`
}