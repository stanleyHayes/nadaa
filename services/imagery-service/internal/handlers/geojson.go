package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/stanleyHayes/nadaa/services/imagery-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/imagery-service/internal/utils"
)

func (s *Server) geoJSONHandler(w http.ResponseWriter, r *http.Request) {
	records := s.store.ListActive()
	features := make([]models.GeoJSONFeature, 0, len(records))
	for _, record := range records {
		features = append(features, models.GeoJSONFeature{
			Type:     "Feature",
			Geometry: record.Geometry,
			Properties: map[string]any{
				"id":               record.ID,
				"reference":        record.Reference,
				"source":           record.Source,
				"captureTime":      record.CaptureTime,
				"resolutionMeters": record.ResolutionMeters,
				"downloadUrl":      fmt.Sprintf("%s/api/v1/imagery/%s/download", s.publicBaseURL(r), record.ID),
			},
		})
	}
	log.Printf("INFO imagery-service geojson count=%d", len(features))
	utils.WriteJSON(w, http.StatusOK, models.GeoJSONFeatureCollection{Type: "FeatureCollection", Features: features})
}

// publicBaseURL returns the configured externally reachable base URL, falling
// back to the request scheme and Host header when none is configured.
func (s *Server) publicBaseURL(r *http.Request) string {
	if s.config.PublicBaseURL != "" {
		return s.config.PublicBaseURL
	}
	return fmt.Sprintf("%s://%s", utils.Scheme(r), r.Host)
}
