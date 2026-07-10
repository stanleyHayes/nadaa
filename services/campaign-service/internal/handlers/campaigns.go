package handlers

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/campaign-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/campaign-service/internal/store"
	"github.com/stanleyHayes/nadaa/services/campaign-service/internal/utils"
)

func (s *Server) listCampaignsHandler(w http.ResponseWriter, r *http.Request) {
	authority, authorityOK := parseAuthority(r)
	includeAll := authorityOK && authority.MFACompleted && allowedCampaignUpdateRoles[authority.ActorRole]

	filters, code, message := parseCampaignFilters(r, includeAll)
	if code != "" {
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}
	filters.IncludeAll = includeAll

	campaigns := s.store.ListCampaigns(r.Context(), filters, s.now().UTC())
	log.Printf("INFO campaign-service campaign_list count=%d includeAll=%t hazard=%s region=%s language=%s", len(campaigns), includeAll, filters.HazardType, filters.Region, filters.Language)
	utils.WriteJSON(w, http.StatusOK, models.CampaignListResponse{Campaigns: campaigns, GeneratedAt: s.now().UTC()})
}

func (s *Server) getCampaignHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	campaign, ok := s.store.GetCampaign(r.Context(), id)
	if !ok || !s.campaignVisible(r, campaign) {
		utils.WriteError(w, http.StatusNotFound, "not_found", "campaign was not found")
		return
	}
	utils.WriteJSON(w, http.StatusOK, models.CampaignResponse{Campaign: campaign})
}

// campaignVisible reports whether the caller may see this campaign: an authorized,
// MFA-completed authority sees any campaign; everyone else sees only published,
// in-window campaigns (matching the list endpoint), so drafts and archived
// campaigns are not leaked by enumerating sequential ids.
func (s *Server) campaignVisible(r *http.Request, campaign models.Campaign) bool {
	authority, ok := parseAuthority(r)
	if ok && authority.MFACompleted && allowedCampaignUpdateRoles[authority.ActorRole] {
		return true
	}
	return store.CampaignPubliclyVisible(campaign, s.now().UTC())
}

func (s *Server) createCampaignHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r)
	if !ok {
		return
	}

	var request models.CreateCampaignRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN campaign-service create_campaign invalid_json actor=%s error=%v", ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeCreate(request, s.now().UTC())
	if code != "" {
		log.Printf("WARN campaign-service create_campaign validation_failed actor=%s code=%s", ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	campaign := s.store.CreateCampaign(r.Context(), normalized, ctx, s.now().UTC())
	log.Printf("INFO campaign-service create_campaign completed id=%s actor=%s status=%s", campaign.ID, ctx.ActorUserID, campaign.Status)
	utils.WriteJSON(w, http.StatusCreated, models.CampaignResponse{Campaign: campaign})
}

func (s *Server) updateCampaignHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r)
	if !ok {
		return
	}

	var request models.UpdateCampaignRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN campaign-service update_campaign invalid_json id=%s actor=%s error=%v", r.PathValue("id"), ctx.ActorUserID, err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	normalized, code, message := normalizeUpdate(request, s.now().UTC())
	if code != "" {
		log.Printf("WARN campaign-service update_campaign validation_failed id=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	campaign, code, message := s.store.UpdateCampaign(r.Context(), r.PathValue("id"), normalized, ctx, s.now().UTC())
	if code != "" {
		log.Printf("WARN campaign-service update_campaign failed id=%s actor=%s code=%s", r.PathValue("id"), ctx.ActorUserID, code)
		utils.WriteError(w, utils.StatusForCode(code), code, message)
		return
	}
	log.Printf("INFO campaign-service update_campaign completed id=%s actor=%s status=%s", campaign.ID, ctx.ActorUserID, campaign.Status)
	utils.WriteJSON(w, http.StatusOK, models.CampaignResponse{Campaign: campaign})
}

func (s *Server) getCampaignMetricsHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	campaign, ok := s.store.GetCampaign(r.Context(), id)
	if !ok || !s.campaignVisible(r, campaign) {
		utils.WriteError(w, http.StatusNotFound, "not_found", "campaign was not found")
		return
	}
	metrics := s.store.ListMetrics(r.Context(), id, s.now().UTC())
	utils.WriteJSON(w, http.StatusOK, models.CampaignMetricListResponse{Metrics: metrics, CampaignID: id, GeneratedAt: s.now().UTC()})
}

func (s *Server) listCampaignTemplatesHandler(w http.ResponseWriter, r *http.Request) {
	templates := s.store.ListTemplates(r.Context())
	utils.WriteJSON(w, http.StatusOK, models.CampaignTemplateListResponse{Templates: templates, GeneratedAt: s.now().UTC()})
}

func parseCampaignFilters(r *http.Request, includeAll bool) (models.CampaignFilters, string, string) {
	query := r.URL.Query()
	filters := models.CampaignFilters{
		Region:     utils.NormalizeQueryValue(query.Get("region")),
		Language:   utils.NormalizeQueryValue(query.Get("language")),
		HazardType: utils.NormalizeQueryValue(query.Get("hazard")),
		Status:     utils.NormalizeQueryValue(query.Get("status")),
	}

	if filters.HazardType != "" && !allowedHazards[filters.HazardType] {
		return models.CampaignFilters{}, "invalid_hazard", "hazard must be a supported NADAA hazard type"
	}
	if filters.Status != "" && !allowedCampaignStatuses[filters.Status] {
		return models.CampaignFilters{}, "invalid_status", "status must be draft, published, or archived"
	}
	if filters.Status != "" && filters.Status != "published" && !includeAll {
		return models.CampaignFilters{}, "forbidden", "only published campaigns are visible to the public"
	}
	return filters, "", ""
}

func normalizeCreate(request models.CreateCampaignRequest, now time.Time) (models.CreateCampaignRequest, string, string) {
	request.Title = strings.TrimSpace(request.Title)
	request.HazardType = utils.NormalizeQueryValue(request.HazardType)
	request.Status = utils.NormalizeQueryValue(request.Status)

	if code, message := validateTitle(request.Title, true); code != "" {
		return request, code, message
	}
	if code, message := validateHazardType(request.HazardType); code != "" {
		return request, code, message
	}
	regions, code, message := validateTargetRegions(request.TargetRegions)
	if code != "" {
		return request, code, message
	}
	request.TargetRegions = regions

	languages, code, message := validateLanguages(request.Languages)
	if code != "" {
		return request, code, message
	}
	request.Languages = languages

	if code, message := validateContentBlocks(request.ContentBlocks); code != "" {
		return request, code, message
	}
	if code, message := validatePublishWindow(request.PublishWindow, request.Status, now); code != "" {
		return request, code, message
	}
	if code, message := validateStatus(request.Status); code != "" {
		return request, code, message
	}
	if code, message := validateLinkedIDs(request.LinkedGuideIDs, request.LinkedAlertIDs); code != "" {
		return request, code, message
	}
	return request, "", ""
}

func normalizeUpdate(request models.UpdateCampaignRequest, now time.Time) (models.UpdateCampaignRequest, string, string) {
	request.Title = strings.TrimSpace(request.Title)
	request.HazardType = utils.NormalizeQueryValue(request.HazardType)
	request.Status = utils.NormalizeQueryValue(request.Status)

	if code, message := validateTitle(request.Title, false); code != "" {
		return request, code, message
	}
	if code, message := validateHazardType(request.HazardType); code != "" {
		return request, code, message
	}
	if request.TargetRegions != nil {
		regions, code, message := validateTargetRegions(request.TargetRegions)
		if code != "" {
			return request, code, message
		}
		request.TargetRegions = regions
	}
	if request.Languages != nil {
		languages, code, message := validateLanguages(request.Languages)
		if code != "" {
			return request, code, message
		}
		request.Languages = languages
	}
	if request.ContentBlocks != nil {
		if code, message := validateContentBlocks(request.ContentBlocks); code != "" {
			return request, code, message
		}
	}
	if request.PublishWindow != nil {
		// Validate window shape against the requested status only; the store applies
		// the authoritative not-premature/not-stale checks against the merged
		// effective status so a status-only publish and a future-dated draft window
		// are both handled correctly.
		if code, message := validatePublishWindow(*request.PublishWindow, request.Status, now); code != "" {
			return request, code, message
		}
	}
	if code, message := validateStatus(request.Status); code != "" {
		return request, code, message
	}
	if code, message := validateLinkedIDs(request.LinkedGuideIDs, request.LinkedAlertIDs); code != "" {
		return request, code, message
	}
	if request.Title == "" && request.HazardType == "" && request.TargetRegions == nil && request.Languages == nil &&
		request.ContentBlocks == nil && request.PublishWindow == nil && request.Status == "" &&
		request.LinkedGuideIDs == nil && request.LinkedAlertIDs == nil {
		return request, "no_changes", "at least one campaign field must be supplied"
	}
	return request, "", ""
}

func validateTitle(title string, required bool) (string, string) {
	if required && (title == "" || len(title) > 200 || utils.UnsafeText(title)) {
		return "invalid_title", "title is required and must be 200 safe characters or fewer"
	}
	if !required && title != "" && (len(title) > 200 || utils.UnsafeText(title)) {
		return "invalid_title", "title must be 200 safe characters or fewer"
	}
	return "", ""
}

func validateHazardType(hazard string) (string, string) {
	if hazard == "" {
		return "", ""
	}
	if !allowedHazards[hazard] {
		return "invalid_hazard", "hazardType must be a supported NADAA hazard type"
	}
	return "", ""
}

func validateStatus(status string) (string, string) {
	if status == "" {
		return "", ""
	}
	if !allowedCampaignStatuses[status] {
		return "invalid_status", "status must be draft, published, or archived"
	}
	return "", ""
}

func validateTargetRegions(regions []string) ([]string, string, string) {
	if len(regions) == 0 {
		return nil, "invalid_target_regions", "at least one target region is required"
	}
	normalized := make([]string, len(regions))
	for i, region := range regions {
		value := strings.TrimSpace(region)
		if value == "" || len(value) > 100 || utils.UnsafeText(value) {
			return nil, "invalid_target_region", "each target region must be 100 safe characters or fewer"
		}
		normalized[i] = value
	}
	return normalized, "", ""
}

func validateLanguages(languages []string) ([]string, string, string) {
	if len(languages) == 0 {
		return nil, "invalid_languages", "at least one language is required"
	}
	normalized := make([]string, len(languages))
	for i, language := range languages {
		value := utils.NormalizeQueryValue(language)
		if len(value) < 2 || len(value) > 5 {
			return nil, "invalid_language", "each language must be a 2-5 character code"
		}
		normalized[i] = value
	}
	return normalized, "", ""
}

func validateContentBlocks(blocks []models.CampaignContentBlock) (string, string) {
	if len(blocks) == 0 {
		return "invalid_content_blocks", "at least one content block is required"
	}
	if len(blocks) > 50 {
		return "invalid_content_blocks", "a campaign may contain up to 50 content blocks"
	}
	for _, block := range blocks {
		blockType := utils.NormalizeQueryValue(block.Type)
		if !allowedContentBlockTypes[blockType] {
			return "invalid_content_block_type", "content block type must be article, checklist, or media"
		}
		if strings.TrimSpace(block.Title) == "" || len(block.Title) > 200 || utils.UnsafeText(block.Title) {
			return "invalid_content_block_title", "each content block title must be 200 safe characters or fewer"
		}
		if len(block.Body) > 5000 || utils.UnsafeText(block.Body) {
			return "invalid_content_block_body", "content block body must be 5000 safe characters or fewer"
		}
		if len(block.Items) > 50 {
			return "invalid_content_block_items", "checklists may contain up to 50 items"
		}
		for _, item := range block.Items {
			if len(item) > 500 || utils.UnsafeText(item) {
				return "invalid_content_block_item", "each checklist item must be 500 safe characters or fewer"
			}
		}
		if len(block.MediaURL) > 500 || utils.UnsafeText(block.MediaURL) {
			return "invalid_content_block_media_url", "mediaUrl must be 500 safe characters or fewer"
		}
	}
	return "", ""
}

func validatePublishWindow(window models.CampaignPublishWindow, status string, now time.Time) (string, string) {
	if window.StartsAt.IsZero() || window.EndsAt.IsZero() {
		return "invalid_publish_window", "publishWindow.startsAt and publishWindow.endsAt are required"
	}
	if window.EndsAt.Before(window.StartsAt) {
		return "invalid_publish_window", "publishWindow.endsAt must be after startsAt"
	}
	if status == "published" && now.After(window.EndsAt) {
		return "stale_campaign", "published campaigns cannot have an ended publish window"
	}
	if status == "published" && now.Before(window.StartsAt) {
		return "premature_campaign", "published campaigns cannot start in the future"
	}
	return "", ""
}

func validateLinkedIDs(guideIDs, alertIDs []string) (string, string) {
	if len(guideIDs) > 50 || len(alertIDs) > 50 {
		return "invalid_linked_ids", "linked guide and alert lists may contain up to 50 ids each"
	}
	for _, id := range guideIDs {
		if len(id) > 100 || utils.UnsafeText(id) {
			return "invalid_linked_guide_id", "each linked guide id must be 100 safe characters or fewer"
		}
	}
	for _, id := range alertIDs {
		if len(id) > 100 || utils.UnsafeText(id) {
			return "invalid_linked_alert_id", "each linked alert id must be 100 safe characters or fewer"
		}
	}
	return "", ""
}
