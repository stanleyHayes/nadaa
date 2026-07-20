package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/utils"
)

const (
	// defaultPageSize is the list endpoint page size when no limit is given.
	defaultPageSize = 50
	// maxPageSize bounds caller-supplied list limits.
	maxPageSize = 200
)

// parsePagination reads limit/offset query params for list endpoints.
func parsePagination(w http.ResponseWriter, r *http.Request) (int, int, bool) {
	limit := defaultPageSize
	offset := 0
	if value := strings.TrimSpace(r.URL.Query().Get("limit")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > maxPageSize {
			utils.WriteError(w, http.StatusBadRequest, "invalid_limit", "limit must be between 1 and 200")
			return 0, 0, false
		}
		limit = parsed
	}
	if value := strings.TrimSpace(r.URL.Query().Get("offset")); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			utils.WriteError(w, http.StatusBadRequest, "invalid_offset", "offset must be zero or a positive integer")
			return 0, 0, false
		}
		offset = parsed
	}
	return limit, offset, true
}
