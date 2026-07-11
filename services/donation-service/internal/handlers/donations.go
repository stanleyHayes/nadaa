package handlers

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/stanleyHayes/nadaa/services/donation-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/donation-service/internal/utils"
)

const (
	minDonationMinor = 100       // GHS 1.00
	maxDonationMinor = 100000000 // GHS 1,000,000.00
	maxWebhookBytes  = 1 << 20
)

var allowedDonationStatuses = map[string]bool{
	"pending": true,
	"paid":    true,
	"failed":  true,
}

// createDonationHandler starts a monetary donation and returns the gateway
// authorization URL the donor is redirected to. The donation is recorded as
// pending; it is only ever marked paid after a server-side verification.
func (s *Server) createDonationHandler(w http.ResponseWriter, r *http.Request) {
	var request models.CreateDonationRequest
	if err := utils.DecodeJSON(r, &request); err != nil {
		log.Printf("WARN donation-service donation_create invalid_json error=%v", err)
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "request body must be valid JSON")
		return
	}

	input, code, message := normalizeCreateDonation(request, s.payments.Name())
	if code != "" {
		log.Printf("WARN donation-service donation_create validation_failed code=%s", code)
		utils.WriteError(w, http.StatusBadRequest, code, message)
		return
	}

	donation := s.store.CreateDonation(input, s.now().UTC())

	init := s.payments.Initialize(r.Context(), models.PaymentInitRequest{
		Reference:   donation.Reference,
		AmountMinor: donation.AmountMinor,
		Currency:    donation.Currency,
		Email:       donation.Email,
		Metadata: map[string]string{
			"donationId": donation.ID,
			"campaign":   donation.Campaign,
		},
	})

	switch init.Status {
	case "initialized":
		if updated, ok := s.store.SetDonationProviderRef(donation.Reference, init.ProviderRef, s.now().UTC()); ok {
			donation = updated
		}
		log.Printf("INFO donation-service donation_create completed id=%s reference=%s provider=%s amountMinor=%d", donation.ID, donation.Reference, donation.Provider, donation.AmountMinor)
		utils.WriteJSON(w, http.StatusCreated, models.CreateDonationResponse{Donation: donation, AuthorizationURL: init.AuthorizationURL})
	case "skipped":
		s.store.MarkDonationFailed(donation.Reference, "payments_unavailable", s.now().UTC())
		log.Printf("WARN donation-service donation_create payments_unavailable reference=%s reason=%s", donation.Reference, init.Reason)
		utils.WriteError(w, http.StatusServiceUnavailable, "payments_unavailable", "donations are not available right now")
	default:
		s.store.MarkDonationFailed(donation.Reference, "initialize_failed", s.now().UTC())
		log.Printf("ERROR donation-service donation_create initialize_failed reference=%s reason=%s", donation.Reference, init.Reason)
		utils.WriteError(w, http.StatusBadGateway, "payment_error", "could not start the payment with the provider")
	}
}

// getDonationHandler returns a donation by reference, verifying a still-pending
// payment with the gateway first so the donor sees an up-to-date status.
func (s *Server) getDonationHandler(w http.ResponseWriter, r *http.Request) {
	reference := strings.TrimSpace(r.PathValue("reference"))
	donation, ok := s.store.GetDonationByReference(reference)
	if !ok {
		utils.WriteError(w, http.StatusNotFound, "not_found", "donation not found")
		return
	}

	if donation.Status == "pending" {
		donation = s.reconcileDonation(r.Context(), donation)
	}
	utils.WriteJSON(w, http.StatusOK, donation)
}

// paystackWebhookHandler receives gateway webhooks. It verifies the signature,
// then re-verifies the transaction server-side before crediting — a webhook
// payload is never treated as proof of payment on its own.
func (s *Server) paystackWebhookHandler(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(io.LimitReader(r.Body, maxWebhookBytes))
	if err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_body", "could not read webhook body")
		return
	}

	signature := r.Header.Get("x-paystack-signature")
	if !s.payments.VerifyWebhookSignature(signature, body) {
		log.Printf("WARN donation-service webhook signature_invalid")
		utils.WriteError(w, http.StatusUnauthorized, "invalid_signature", "webhook signature verification failed")
		return
	}

	var event struct {
		Event string `json:"event"`
		Data  struct {
			Reference string `json:"reference"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		utils.WriteError(w, http.StatusBadRequest, "invalid_json", "webhook body must be valid JSON")
		return
	}

	reference := strings.TrimSpace(event.Data.Reference)
	if event.Event != "charge.success" || reference == "" {
		log.Printf("INFO donation-service webhook ignored event=%s", event.Event)
		utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ignored"})
		return
	}

	donation, ok := s.store.GetDonationByReference(reference)
	if !ok {
		log.Printf("WARN donation-service webhook unknown_reference reference=%s", reference)
		utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ignored"})
		return
	}

	s.reconcileDonation(r.Context(), donation)
	utils.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// listDonationsHandler returns donations for authority reconciliation.
func (s *Server) listDonationsHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := requireAuthority(w, r)
	if !ok {
		return
	}

	filter := models.DonationFilter{
		Status:   utils.NormalizeToken(r.URL.Query().Get("status")),
		Campaign: utils.NormalizeToken(r.URL.Query().Get("campaign")),
	}
	if filter.Status != "" && !allowedDonationStatuses[filter.Status] {
		utils.WriteError(w, http.StatusBadRequest, "invalid_status", "status must be pending, paid, or failed")
		return
	}

	donations := s.store.ListDonations(filter)
	log.Printf("INFO donation-service donation_list count=%d actor=%s status=%s", len(donations), ctx.ActorUserID, filter.Status)
	utils.WriteJSON(w, http.StatusOK, models.DonationListResponse{Donations: donations, GeneratedAt: s.now().UTC()})
}

// reconcileDonation verifies a pending donation with the gateway and applies the
// authoritative result to the store. It guards against an amount mismatch so a
// tampered or mismatched transaction is never credited, and relies on the
// store's idempotent transitions so repeated calls are safe.
func (s *Server) reconcileDonation(ctx context.Context, donation models.Donation) models.Donation {
	result := s.payments.Verify(ctx, donation.Reference)
	now := s.now().UTC()

	switch result.Status {
	case "paid":
		if result.AmountMinor > 0 && result.AmountMinor != donation.AmountMinor {
			log.Printf("ERROR donation-service donation_amount_mismatch reference=%s expected=%d verified=%d", donation.Reference, donation.AmountMinor, result.AmountMinor)
			if updated, ok := s.store.MarkDonationFailed(donation.Reference, "amount_mismatch", now); ok {
				return updated
			}
			return donation
		}
		if updated, ok := s.store.MarkDonationPaid(donation.Reference, result.Channel, result.ProviderRef, now); ok {
			log.Printf("INFO donation-service donation_paid reference=%s channel=%s", donation.Reference, result.Channel)
			return updated
		}
	case "failed":
		if updated, ok := s.store.MarkDonationFailed(donation.Reference, "verify_failed", now); ok {
			return updated
		}
	}
	return donation
}

func normalizeCreateDonation(request models.CreateDonationRequest, provider string) (models.CreateDonationInput, string, string) {
	donorName := strings.TrimSpace(request.DonorName)
	if donorName == "" {
		donorName = "Anonymous donor"
	}
	email := strings.TrimSpace(strings.ToLower(request.Email))
	currency := strings.ToUpper(strings.TrimSpace(request.Currency))
	if currency == "" {
		currency = "GHS"
	}
	campaign := strings.TrimSpace(request.Campaign)
	message := strings.TrimSpace(request.Message)

	if !utils.ValidEmail(email) {
		return models.CreateDonationInput{}, "invalid_email", "a valid email is required to start a payment"
	}
	if currency != "GHS" {
		return models.CreateDonationInput{}, "invalid_currency", "currency must be GHS"
	}
	if len(donorName) > 120 || utils.UnsafeText(donorName) {
		return models.CreateDonationInput{}, "invalid_donor_name", "donorName must be 120 safe characters or fewer"
	}
	if len(campaign) > 120 || utils.UnsafeText(campaign) {
		return models.CreateDonationInput{}, "invalid_campaign", "campaign must be 120 safe characters or fewer"
	}
	if len(message) > 500 || utils.UnsafeText(message) {
		return models.CreateDonationInput{}, "invalid_message", "message must be 500 safe characters or fewer"
	}

	amountMinor := utils.MajorToMinor(request.Amount)
	if amountMinor < minDonationMinor {
		return models.CreateDonationInput{}, "invalid_amount", "amount must be at least GHS 1.00"
	}
	if amountMinor > maxDonationMinor {
		return models.CreateDonationInput{}, "invalid_amount", "amount is above the accepted limit"
	}

	return models.CreateDonationInput{
		DonorName:   donorName,
		Email:       email,
		AmountMinor: amountMinor,
		Currency:    currency,
		Campaign:    campaign,
		Message:     message,
		Provider:    provider,
	}, "", ""
}
