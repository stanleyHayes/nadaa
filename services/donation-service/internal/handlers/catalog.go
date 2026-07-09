package handlers

import (
	"log"
	"net/http"

	"github.com/stanleyHayes/nadaa/services/donation-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/donation-service/internal/utils"
)

func (s *Server) listCatalogHandler(w http.ResponseWriter, _ *http.Request) {
	items := s.store.ListCatalog()
	log.Printf("INFO donation-service aid_catalog_list count=%d", len(items))
	utils.WriteJSON(w, http.StatusOK, models.AidCatalogResponse{Items: items, GeneratedAt: s.now().UTC()})
}
