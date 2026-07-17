package handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/damage-claim-service/internal/utils"
)

func (s *Server) exportClaimHandler(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.requireAuthority(w, r, authorityRoles)
	if !ok {
		return
	}

	format := utils.NormalizeString(r.URL.Query().Get("format"))
	if format == "" {
		format = "csv"
	}
	if format != "csv" && format != "pdf" {
		// #nosec G706 -- path value and format are sanitized with utils.SafeLogValue.
		log.Printf("WARN damage-claim-service claim_export invalid_format id=%s actor=%s format=%s", utils.SafeLogValue(r.PathValue("id")), ctx.ActorUserID, utils.SafeLogValue(format))
		utils.WriteError(w, http.StatusBadRequest, "invalid_format", "format must be csv or pdf")
		return
	}

	claim, ok := s.store.Get(r.PathValue("id"))
	if !ok {
		// #nosec G706 -- path value is sanitized with utils.SafeLogValue.
		log.Printf("WARN damage-claim-service claim_export not_found id=%s actor=%s", utils.SafeLogValue(r.PathValue("id")), ctx.ActorUserID)
		utils.WriteError(w, http.StatusNotFound, "not_found", "claim was not found")
		return
	}

	switch format {
	case "csv":
		writeClaimCSV(w, claim)
	case "pdf":
		writeClaimPDF(w, claim)
	}
	// #nosec G706 -- format is sanitized with utils.SafeLogValue.
	log.Printf("INFO damage-claim-service claim_export completed id=%s reference=%s format=%s actor=%s", claim.ID, claim.Reference, utils.SafeLogValue(format), ctx.ActorUserID)
}

func writeClaimCSV(w http.ResponseWriter, claim models.DamageClaimRecord) {
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"claim_%s.csv\"", claim.Reference))
	w.WriteHeader(http.StatusOK)

	writer := csv.NewWriter(w)
	_ = writer.Write([]string{"Field", "Value"})
	_ = writer.Write([]string{"Reference", csvCell(claim.Reference)})
	_ = writer.Write([]string{"IncidentReference", csvCell(claim.IncidentReference)})
	_ = writer.Write([]string{"ReporterName", csvCell(claim.Reporter.Name)})
	_ = writer.Write([]string{"ReporterPhone", csvCell(claim.Reporter.Phone)})
	_ = writer.Write([]string{"ReporterEmail", csvCell(claim.Reporter.Email)})
	_ = writer.Write([]string{"DamageType", csvCell(claim.DamageType)})
	_ = writer.Write([]string{"DamageDescription", csvCell(claim.DamageDescription)})
	_ = writer.Write([]string{"EstimatedLossAmount", csvCell(claim.EstimatedLossAmount)})
	_ = writer.Write([]string{"VerificationStatus", csvCell(claim.VerificationStatus)})
	_ = writer.Write([]string{"VerifiedBy", csvCell(claim.VerifiedBy)})
	_ = writer.Write([]string{"VerificationNotes", csvCell(claim.VerificationNotes)})
	_ = writer.Write([]string{"Status", csvCell(claim.Status)})
	_ = writer.Write([]string{"LocationAddress", csvCell(claim.Location.Address)})
	_ = writer.Write([]string{"LocationLat", fmt.Sprintf("%f", claim.Location.Lat)})
	_ = writer.Write([]string{"LocationLng", fmt.Sprintf("%f", claim.Location.Lng)})
	_ = writer.Write([]string{"CreatedAt", claim.CreatedAt.Format(time.RFC3339)})
	_ = writer.Write([]string{"UpdatedAt", claim.UpdatedAt.Format(time.RFC3339)})
	writer.Flush()
}

// csvCell neutralizes CSV formula injection: spreadsheet apps execute cells
// whose first character is =, +, -, or @, so a leading single quote forces the
// value to be treated as text.
func csvCell(value string) string {
	if value == "" {
		return value
	}
	switch value[0] {
	case '=', '+', '-', '@':
		return "'" + value
	default:
		return value
	}
}

func writeClaimPDF(w http.ResponseWriter, claim models.DamageClaimRecord) {
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"claim_%s.pdf\"", claim.Reference))
	w.WriteHeader(http.StatusOK)

	pdf := buildClaimPDF(claim)
	_, _ = w.Write(pdf)
}

func buildClaimPDF(claim models.DamageClaimRecord) []byte {
	var buf bytes.Buffer

	type line struct {
		label string
		value string
	}
	lines := []line{
		{label: "Damage Claim Report", value: ""},
		{label: "", value: ""},
		{label: "Reference:", value: claim.Reference},
		{label: "Incident Reference:", value: claim.IncidentReference},
		{label: "Reporter Name:", value: claim.Reporter.Name},
		{label: "Damage Type:", value: claim.DamageType},
		{label: "Description:", value: claim.DamageDescription},
		{label: "Estimated Loss Amount:", value: claim.EstimatedLossAmount},
		{label: "Verification Status:", value: claim.VerificationStatus},
		{label: "Location:", value: claim.Location.Address},
		{label: "Notes:", value: claim.VerificationNotes},
	}

	streamLines := []string{
		"BT",
		"/F1 12 Tf",
	}
	y := 720
	for _, l := range lines {
		if l.label == "" {
			y -= 14
			continue
		}
		if l.value == "" {
			streamLines = append(streamLines, fmt.Sprintf("72 %d Td (%s) Tj", y, pdfEscape(l.label)))
		} else {
			streamLines = append(streamLines, fmt.Sprintf("72 %d Td (%s %s) Tj", y, pdfEscape(l.label), pdfEscape(l.value)))
		}
		streamLines = append(streamLines, "0 -14 Td")
		y -= 14
	}
	streamLines = append(streamLines, "ET")
	stream := strings.Join(streamLines, "\n")

	buf.WriteString("%PDF-1.4\n")
	objOffsets := []int64{}

	writeObj := func(content string) {
		objOffsets = append(objOffsets, int64(buf.Len()))
		buf.WriteString(content)
	}

	writeObj("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")
	writeObj("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")
	writeObj("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>\nendobj\n")
	writeObj(fmt.Sprintf("4 0 obj\n<< /Length %d >>\nstream\n%s\nendstream\nendobj\n", len(stream), stream))
	writeObj("5 0 obj\n<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>\nendobj\n")

	xrefOffset := buf.Len()
	buf.WriteString("xref\n")
	_, _ = fmt.Fprintf(&buf, "0 %d\n", len(objOffsets)+1)
	buf.WriteString("0000000000 65535 f \n")
	for _, offset := range objOffsets {
		_, _ = fmt.Fprintf(&buf, "%010d 00000 n \n", offset)
	}

	buf.WriteString("trailer\n")
	_, _ = fmt.Fprintf(&buf, "<< /Size %d /Root 1 0 R >>\n", len(objOffsets)+1)
	buf.WriteString("startxref\n")
	_, _ = fmt.Fprintf(&buf, "%d\n", xrefOffset)
	buf.WriteString("%%EOF\n")

	return buf.Bytes()
}

func pdfEscape(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "(", "\\(")
	value = strings.ReplaceAll(value, ")", "\\)")
	return value
}
