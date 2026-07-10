package store

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/stanleyHayes/nadaa/services/ml-service/internal/models"
	"github.com/stanleyHayes/nadaa/services/ml-service/internal/utils"
)

const (
	// ForecastModelVersion identifies the deterministic resource-demand model.
	ForecastModelVersion = "resource-forecast-rules-0.1.0"

	forecastHazardType     = "flood"
	defaultForecastWindowH = 24
	stagingRadiusMeters    = 5000
)

// severityRank orders the flood-risk label severities so a district can inherit
// its worst constituent cell severity.
var severityRank = map[string]int{
	"low": 1, "moderate": 2, "high": 3, "severe": 4, "emergency": 5,
}

var severityByRank = map[int]string{
	1: "low", 2: "moderate", 3: "high", 4: "severe", 5: "emergency",
}

// stagingBase is a fixed candidate staging position for a responder agency.
type stagingBase struct {
	id         string
	label      string
	agencyType string
	location   models.Coordinates
}

// stagingBases are the deterministic candidate staging positions the planner can
// recommend. Real deployments would source these from an agency facility registry.
var stagingBases = []stagingBase{
	{"staging_accra_central_fire", "Accra Central Fire Station", "fire", models.Coordinates{Lat: 5.545, Lng: -0.205}},
	{"staging_osu_fire", "Osu Fire Post", "fire", models.Coordinates{Lat: 5.556, Lng: -0.182}},
	{"staging_tema_fire", "Tema Fire Station", "fire", models.Coordinates{Lat: 5.669, Lng: -0.016}},
	{"staging_ridge_ambulance", "Ridge Ambulance Base", "ambulance", models.Coordinates{Lat: 5.563, Lng: -0.190}},
	{"staging_kaneshie_ambulance", "Kaneshie Ambulance Base", "ambulance", models.Coordinates{Lat: 5.570, Lng: -0.235}},
	{"staging_tema_ambulance", "Tema Ambulance Point", "ambulance", models.Coordinates{Lat: 5.675, Lng: -0.020}},
	{"staging_nadmo_accra", "NADMO Accra Metro Depot", "nadmo", models.Coordinates{Lat: 5.560, Lng: -0.200}},
}

// unitsPerAgency sets how many predicted incidents one staged unit is sized for.
var unitsPerAgency = map[string]float64{"fire": 8, "ambulance": 6, "nadmo": 10}

// forecastOptions carries the scenario levers applied to the baseline model.
type forecastOptions struct {
	region           string
	riskLevel        string
	hazardTypes      []string
	historicalWeight float64
	timeWindowHours  int
}

func defaultForecastOptions() forecastOptions {
	return forecastOptions{historicalWeight: 1, timeWindowHours: defaultForecastWindowH}
}

// districtForecast pairs a demand forecast with its district centroid so staging
// suggestions can be matched to demand without recomputing geometry.
type districtForecast struct {
	forecast models.DemandForecast
	centroid models.Coordinates
}

// ListForecasts returns per-district demand forecasts, optionally filtered by region.
func (m *MemoryStore) ListForecasts(region string, now time.Time) []models.DemandForecast {
	opts := defaultForecastOptions()
	opts.region = region
	districts := m.computeDistrictForecasts(opts, now)
	forecasts := make([]models.DemandForecast, 0, len(districts))
	for _, d := range districts {
		forecasts = append(forecasts, d.forecast)
	}
	return forecasts
}

// StagingSuggestions returns recommended staging positions, optionally filtered by
// agency type, derived from the current demand forecasts.
func (m *MemoryStore) StagingSuggestions(agencyType string, now time.Time) []models.StagingSuggestion {
	districts := m.computeDistrictForecasts(defaultForecastOptions(), now)
	generatedAt := now.UTC().Format(time.RFC3339)
	agencyType = strings.TrimSpace(strings.ToLower(agencyType))

	suggestions := make([]models.StagingSuggestion, 0, len(stagingBases))
	for _, base := range stagingBases {
		if agencyType != "" && base.agencyType != agencyType {
			continue
		}
		nearest, ok := nearestDistrict(base.location, districts)
		if !ok {
			continue
		}
		suggestions = append(suggestions, buildStagingSuggestion(base, nearest, generatedAt))
	}

	sort.SliceStable(suggestions, func(i, j int) bool {
		if suggestions[i].RecommendedUnits != suggestions[j].RecommendedUnits {
			return suggestions[i].RecommendedUnits > suggestions[j].RecommendedUnits
		}
		return suggestions[i].ID < suggestions[j].ID
	})
	return suggestions
}

// CompareScenarios returns the baseline forecast alongside an adjusted scenario.
// The scope filters (region, riskLevel, hazardTypes) apply to BOTH scenarios so
// their totals stay comparable; only the levers (historicalWeight, capacityFactor,
// timeWindowHours) differ between the baseline and the adjusted scenario.
func (m *MemoryStore) CompareScenarios(req models.CompareScenarioRequest, now time.Time) []models.ScenarioResult {
	baseParams := models.CompareScenarioRequest{
		Region:      req.Region,
		RiskLevel:   req.RiskLevel,
		HazardTypes: req.HazardTypes,
	}
	baseOpts := defaultForecastOptions()
	baseOpts.region = req.Region
	baseOpts.riskLevel = strings.TrimSpace(strings.ToLower(req.RiskLevel))
	baseOpts.hazardTypes = req.HazardTypes
	baseline := m.scenarioResult("Current conditions", baseParams, baseOpts, now)

	adjusted := m.scenarioResult("Adjusted scenario", req, optionsFromRequest(req), now)
	return []models.ScenarioResult{baseline, adjusted}
}

func (m *MemoryStore) scenarioResult(name string, params models.CompareScenarioRequest, opts forecastOptions, now time.Time) models.ScenarioResult {
	districts := m.computeDistrictForecasts(opts, now)
	forecasts := make([]models.DemandForecast, 0, len(districts))
	for _, d := range districts {
		forecasts = append(forecasts, d.forecast)
	}
	return models.ScenarioResult{
		Name:       name,
		Parameters: params,
		Forecasts:  forecasts,
		Summary:    summarizeForecasts(forecasts),
	}
}

func optionsFromRequest(req models.CompareScenarioRequest) forecastOptions {
	opts := defaultForecastOptions()
	opts.region = req.Region
	opts.riskLevel = strings.TrimSpace(strings.ToLower(req.RiskLevel))
	opts.hazardTypes = req.HazardTypes
	if req.HistoricalWeight != 0 {
		opts.historicalWeight = req.HistoricalWeight
	}
	if req.TimeWindowHours != 0 {
		opts.timeWindowHours = req.TimeWindowHours
	}
	// capacityFactor is echoed back in the scenario parameters but is reserved for
	// future capacity-aware staging optimization; it does not change demand counts.
	return opts
}

// computeDistrictForecasts aggregates the flood-risk feature grid into per-district
// demand forecasts. The math is fully deterministic given the seeded feature rows.
func (m *MemoryStore) computeDistrictForecasts(opts forecastOptions, now time.Time) []districtForecast {
	if opts.historicalWeight == 0 {
		opts.historicalWeight = 1
	}
	if opts.timeWindowHours == 0 {
		opts.timeWindowHours = defaultForecastWindowH
	}
	if len(opts.hazardTypes) > 0 && !containsFold(opts.hazardTypes, forecastHazardType) {
		return nil
	}

	type agg struct {
		region, district                              string
		histTotal, sumComposite, sumRainfall, sumVuln float64
		sumLat, sumLng                                float64
		cellCount, worstRank                          int
	}

	groups := map[string]*agg{}
	order := make([]string, 0)
	regionFilter := strings.TrimSpace(strings.ToLower(opts.region))

	for _, row := range m.features {
		if regionFilter != "" && strings.ToLower(row.Region) != regionFilter {
			continue
		}
		key := row.Region + "|" + row.District
		g, ok := groups[key]
		if !ok {
			g = &agg{region: row.Region, district: row.District}
			groups[key] = g
			order = append(order, key)
		}
		g.histTotal += featureValue(row, "historical_flood_reports_30d")
		g.sumComposite += featureValue(row, "composite_rule_score")
		g.sumRainfall += featureValue(row, "rainfall_forecast_24h_mm")
		g.sumVuln += featureValue(row, "vulnerable_population_pct")
		g.sumLat += featureValue(row, "centroid_lat")
		g.sumLng += featureValue(row, "centroid_lng")
		g.cellCount++
		if rank := severityRank[strings.ToLower(severityLabel(row))]; rank > g.worstRank {
			g.worstRank = rank
		}
	}

	sort.Strings(order)
	generatedAt := now.UTC().Format(time.RFC3339)
	windowStart := now.UTC()
	windowEnd := windowStart.Add(time.Duration(opts.timeWindowHours) * time.Hour)

	results := make([]districtForecast, 0, len(order))
	for _, key := range order {
		g := groups[key]
		if g.cellCount == 0 {
			continue
		}
		cells := float64(g.cellCount)
		avgComposite := g.sumComposite / cells
		avgRainfall := g.sumRainfall / cells
		avgVuln := g.sumVuln / cells

		windowFactor := float64(opts.timeWindowHours) / 24.0
		exposureFactor := 1.0 + (avgVuln/100.0)*0.75
		predictedFloat := (g.histTotal*opts.historicalWeight + avgComposite*4.0) *
			(0.7 + avgRainfall/150.0) * exposureFactor * windowFactor
		predicted := max(int(math.Round(predictedFloat)), 0)

		riskLevel := riskLevelForDistrict(g.worstRank, avgComposite)
		// riskLevel is a minimum-severity threshold: keep districts at or above it.
		if opts.riskLevel != "" && severityRank[riskLevel] < severityRank[opts.riskLevel] {
			continue
		}

		confidenceScore := forecastConfidenceScore(g.histTotal, g.cellCount, avgComposite)
		forecast := models.DemandForecast{
			ID:                     "forecast_" + utils.SanitizeID(g.district),
			Region:                 g.region,
			District:               g.district,
			TimeWindowStart:        windowStart.Format(time.RFC3339),
			TimeWindowEnd:          windowEnd.Format(time.RFC3339),
			PredictedIncidentCount: predicted,
			HazardType:             forecastHazardType,
			Confidence:             confidenceBand(confidenceScore),
			ConfidenceScore:        roundTo(confidenceScore, 2),
			Factors:                forecastFactors(g.histTotal, opts.historicalWeight, avgRainfall, avgComposite, avgVuln),
			RiskLevel:              riskLevel,
			GeneratedAt:            generatedAt,
		}
		results = append(results, districtForecast{
			forecast: forecast,
			centroid: models.Coordinates{Lat: g.sumLat / cells, Lng: g.sumLng / cells},
		})
	}

	sort.SliceStable(results, func(i, j int) bool {
		if results[i].forecast.PredictedIncidentCount != results[j].forecast.PredictedIncidentCount {
			return results[i].forecast.PredictedIncidentCount > results[j].forecast.PredictedIncidentCount
		}
		return results[i].forecast.District < results[j].forecast.District
	})
	return results
}

func forecastFactors(histTotal, historicalWeight, avgRainfall, avgComposite, avgVuln float64) []models.ForecastFactor {
	return []models.ForecastFactor{
		{Name: "historical_incidents", Label: "Historical flood reports (30d)", Value: roundTo(histTotal*historicalWeight, 2), Weight: 0.35, Direction: "increases_demand"},
		{Name: "rainfall_forecast", Label: "Rainfall forecast 24h (mm)", Value: roundTo(avgRainfall, 2), Weight: 0.25, Direction: "increases_demand"},
		{Name: "risk_score", Label: "Composite flood-risk score", Value: roundTo(avgComposite, 4), Weight: 0.25, Direction: "increases_demand"},
		{Name: "population_exposure", Label: "Vulnerable population (%)", Value: roundTo(avgVuln, 2), Weight: 0.15, Direction: "increases_demand"},
	}
}

func buildStagingSuggestion(base stagingBase, nearest districtForecast, generatedAt string) models.StagingSuggestion {
	perUnit := unitsPerAgency[base.agencyType]
	if perUnit <= 0 {
		perUnit = 8
	}
	units := min(max(int(math.Ceil(float64(nearest.forecast.PredictedIncidentCount)/perUnit)), 1), 5)

	reason := fmt.Sprintf("Elevated predicted flood demand in %s (~%d incidents in 24h, %s risk)",
		nearest.forecast.District, nearest.forecast.PredictedIncidentCount, nearest.forecast.RiskLevel)

	return models.StagingSuggestion{
		ID:                     base.id,
		Location:               base.location,
		LocationLabel:          base.label,
		AgencyType:             base.agencyType,
		Reason:                 reason,
		Confidence:             nearest.forecast.Confidence,
		ConfidenceScore:        nearest.forecast.ConfidenceScore,
		OperationalConstraints: operationalConstraints(base.agencyType, nearest.forecast.RiskLevel),
		RecommendedUnits:       units,
		RadiusMeters:           stagingRadiusMeters,
		GeneratedAt:            generatedAt,
	}
}

func operationalConstraints(agencyType, riskLevel string) []string {
	constraints := []string{"Road congestion can extend response times during peak hours."}
	switch agencyType {
	case "fire":
		constraints = append(constraints, "Confirm water tanker and hydrant availability before repositioning.")
	case "ambulance":
		constraints = append(constraints, "Coordinate with hospital emergency capacity before staging (NADAA-121).")
	case "nadmo":
		constraints = append(constraints, "Coordinate multi-agency staging through the district disaster committee.")
	}
	if riskLevel == "severe" || riskLevel == "emergency" {
		constraints = append(constraints, "Flooded access routes may require alternate approaches; verify with route planning (NADAA-130).")
	}
	return constraints
}

func summarizeForecasts(forecasts []models.DemandForecast) models.ScenarioSummary {
	summary := models.ScenarioSummary{}
	if len(forecasts) == 0 {
		return summary
	}
	total := 0
	confidenceSum := 0.0
	topDemand := -1
	for _, f := range forecasts {
		total += f.PredictedIncidentCount
		confidenceSum += f.ConfidenceScore
		if f.PredictedIncidentCount > topDemand {
			topDemand = f.PredictedIncidentCount
			summary.HighestRiskRegion = f.Region
			summary.HighestRiskHazard = f.HazardType
		}
	}
	summary.TotalPredictedIncidents = total
	summary.AverageConfidenceScore = roundTo(confidenceSum/float64(len(forecasts)), 2)
	return summary
}

func nearestDistrict(location models.Coordinates, districts []districtForecast) (districtForecast, bool) {
	if len(districts) == 0 {
		return districtForecast{}, false
	}
	best := districts[0]
	bestDist := utils.HaversineMeters(location, best.centroid)
	for _, d := range districts[1:] {
		if dist := utils.HaversineMeters(location, d.centroid); dist < bestDist {
			bestDist = dist
			best = d
		}
	}
	return best, true
}

func riskLevelForDistrict(worstRank int, avgComposite float64) string {
	if label, ok := severityByRank[worstRank]; ok {
		return label
	}
	switch {
	case avgComposite >= 0.85:
		return "emergency"
	case avgComposite >= 0.7:
		return "severe"
	case avgComposite >= 0.5:
		return "high"
	case avgComposite >= 0.3:
		return "moderate"
	default:
		return "low"
	}
}

func forecastConfidenceScore(histTotal float64, cellCount int, avgComposite float64) float64 {
	score := 0.5
	if histTotal > 0 {
		score += 0.2
	}
	score += 0.15 * math.Min(1.0, float64(cellCount)/3.0)
	if avgComposite >= 0.7 || avgComposite <= 0.3 {
		score += 0.1
	}
	return clampFloat(score, 0.3, 0.95)
}

func confidenceBand(score float64) string {
	switch {
	case score >= 0.8:
		return "high"
	case score >= 0.6:
		return "medium"
	default:
		return "low"
	}
}

// severityLabel returns a feature row's label_severity string value.
func severityLabel(row FeatureRow) string {
	if raw, ok := row.Values["label_severity"]; ok {
		if s, ok := raw.(string); ok {
			return s
		}
	}
	return ""
}

func containsFold(values []string, target string) bool {
	for _, v := range values {
		if strings.EqualFold(strings.TrimSpace(v), target) {
			return true
		}
	}
	return false
}

func clampFloat(value, lower, upper float64) float64 {
	if value < lower {
		return lower
	}
	if value > upper {
		return upper
	}
	return value
}

func roundTo(value float64, decimals int) float64 {
	factor := math.Pow(10, float64(decimals))
	return math.Round(value*factor) / factor
}
