package models

// CVLabel represents a single detected label from computer vision analysis.
type CVLabel struct {
	Label      string  `json:"label"`
	Confidence float64 `json:"confidence"`
	Bbox       []int   `json:"bbox,omitempty"`
}

// CVAnalysisRequest is the body of a CV analyze request.
type CVAnalysisRequest struct {
	ImageID   string `json:"imageId"`
	ImageURL  string `json:"imageUrl,omitempty"`
	ImageName string `json:"imageName,omitempty"`
}

// CVAnalysisResult is the response from a CV analyze request.
type CVAnalysisResult struct {
	ID                  string    `json:"id"`
	ImageID             string    `json:"imageId"`
	Labels              []CVLabel `json:"labels"`
	ModelVersion        string    `json:"modelVersion"`
	Limitations         string    `json:"limitations"`
	HumanReviewRequired bool      `json:"humanReviewRequired"`
	CreatedAt           string    `json:"createdAt"`
	ReviewedBy          string    `json:"reviewedBy,omitempty"`
	ReviewedAt          string    `json:"reviewedAt,omitempty"`
	ReviewStatus        string    `json:"reviewStatus,omitempty"`
	ReviewNote          string    `json:"reviewNote,omitempty"`
}

// CVReviewRequest is the body of a CV human-review decision.
type CVReviewRequest struct {
	Decision string `json:"decision"`
	Note     string `json:"note,omitempty"`
}

// CVAnalysisResponse is the top-level response envelope.
type CVAnalysisResponse struct {
	Result CVAnalysisResult `json:"result"`
	Safety SafetyPolicy     `json:"safety"`
}

// CVResultListResponse is returned when listing cached CV results.
type CVResultListResponse struct {
	Results []CVAnalysisResult `json:"results"`
	Total   int                `json:"total"`
	Limit   int                `json:"limit"`
	Offset  int                `json:"offset"`
}

// CVResultDetailResponse is returned when fetching a single CV result.
type CVResultDetailResponse struct {
	Result CVAnalysisResult `json:"result"`
}
