package handlers

import "net/http"

// Routes returns the configured HTTP handler with middleware applied.
func (s *Server) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.healthHandler)
	mux.HandleFunc("GET /api/v1/notifications/alerts", s.listAlertsHandler)
	mux.HandleFunc("POST /api/v1/notifications/alerts/{id}/deliver", s.deliverAlertHandler)
	mux.HandleFunc("GET /api/v1/notifications/delivery-logs", s.listDeliveryLogsHandler)
	mux.HandleFunc("POST /api/v1/notifications/voice-alerts", s.createVoiceAlertHandler)
	mux.HandleFunc("GET /api/v1/notifications/voice-alerts", s.listVoiceAlertsHandler)
	mux.HandleFunc("POST /api/v1/notifications/voice-alerts/{id}/review", s.reviewVoiceAlertHandler)
	mux.HandleFunc("POST /api/v1/notifications/voice-alerts/{id}/deliver", s.deliverVoiceAlertHandler)
	mux.HandleFunc("POST /api/v1/notifications/ussd", s.ussdWebhookHandler)
	mux.HandleFunc("POST /api/v1/notifications/sms/inbound", s.smsInboundHandler)
	mux.HandleFunc("POST /api/v1/notifications/whatsapp/inbound", s.whatsappWebhookHandler)
	mux.HandleFunc("POST /api/v1/notifications/whatsapp/webhook", s.whatsappWebhookHandler)
	mux.HandleFunc("GET /api/v1/notifications/access-logs", s.listAccessLogsHandler)
	return s.withMiddleware(mux)
}
