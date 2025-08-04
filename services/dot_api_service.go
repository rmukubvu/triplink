package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// DOT API Service Implementation
// Uses various state DOT APIs (Caltrans, FDOT, etc.) and 511.org data
type DOTAPIService struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

func NewDOTAPIService() *DOTAPIService {
	return &DOTAPIService{
		APIKey:  os.Getenv("DOT_API_KEY"), // 511.org or state DOT API key
		BaseURL: "https://api.511.org/traffic",
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Implement ConstructionAPIService interface
func (d *DOTAPIService) GetConstructionAlerts(bounds BoundingBox) ([]ConstructionAlert, error) {
	// 511.org construction alerts API
	params := url.Values{}
	params.Set("api_key", d.APIKey)
	params.Set("format", "json")
	params.Set("bbox", fmt.Sprintf("%f,%f,%f,%f",
		bounds.SouthWest.Longitude, bounds.SouthWest.Latitude,
		bounds.NorthEast.Longitude, bounds.NorthEast.Latitude))
	params.Set("event_subtypes", "construction")

	apiURL := fmt.Sprintf("%s/events?%s", d.BaseURL, params.Encode())

	resp, err := d.HTTPClient.Get(apiURL)
	if err != nil {
		// If API fails, return mock data for development
		return d.getMockConstructionAlerts(bounds), nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var constructionResp struct {
		Events []struct {
			ID          string `json:"id"`
			EventType   string `json:"event_type"`
			EventSubtype string `json:"event_subtype"`
			Headline    string `json:"headline"`
			Description string `json:"description"`
			Severity    string `json:"severity"`
			Status      string `json:"status"`
			RoadName    string `json:"road_name"`
			Direction   string `json:"direction"`
			Location    struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"location"`
			StartTime   string `json:"start_time"`
			EndTime     string `json:"end_time"`
			LastUpdated string `json:"last_updated"`
			Agency      struct {
				Name    string `json:"name"`
				Contact string `json:"contact"`
			} `json:"agency"`
			DetourInfo struct {
				Available   bool    `json:"available"`
				Description string  `json:"description"`
				ExtraTime   float64 `json:"extra_time_minutes"`
				ExtraDistance float64 `json:"extra_distance_km"`
			} `json:"detour_info"`
		} `json:"events"`
	}

	if err := json.Unmarshal(body, &constructionResp); err != nil {
		return d.getMockConstructionAlerts(bounds), nil
	}

	// Convert API response to our format
	var alerts []ConstructionAlert
	for _, event := range constructionResp.Events {
		startTime, _ := time.Parse(time.RFC3339, event.StartTime)
		var endTime *time.Time
		if event.EndTime != "" {
			et, _ := time.Parse(time.RFC3339, event.EndTime)
			endTime = &et
		}

		var detour *DetourInfo
		if event.DetourInfo.Available {
			detour = &DetourInfo{
				Route:         "Alternative route available",
				Description:   event.DetourInfo.Description,
				ExtraTime:     event.DetourInfo.ExtraTime,
				ExtraDistance: event.DetourInfo.ExtraDistance,
			}
		}

		alert := ConstructionAlert{
			AlertID:     event.ID,
			Title:       event.Headline,
			Description: event.Description,
			Location:    Coordinate{Latitude: event.Location.Latitude, Longitude: event.Location.Longitude},
			RoadName:    event.RoadName,
			StartDate:   startTime,
			EndDate:     endTime,
			Severity:    d.mapSeverity(event.Severity),
			Impact:      d.determineImpact(event.Description),
			Detour:      detour,
			Agency:      event.Agency.Name,
			Contact:     event.Agency.Contact,
		}

		alerts = append(alerts, alert)
	}

	return alerts, nil
}

func (d *DOTAPIService) GetRoadClosures(bounds BoundingBox) ([]RoadClosure, error) {
	// Similar to construction alerts but filter for road closures
	params := url.Values{}
	params.Set("api_key", d.APIKey)
	params.Set("format", "json")
	params.Set("bbox", fmt.Sprintf("%f,%f,%f,%f",
		bounds.SouthWest.Longitude, bounds.SouthWest.Latitude,
		bounds.NorthEast.Longitude, bounds.NorthEast.Latitude))
	params.Set("event_subtypes", "road_closure,lane_closure")

	apiURL := fmt.Sprintf("%s/events?%s", d.BaseURL, params.Encode())

	resp, err := d.HTTPClient.Get(apiURL)
	if err != nil {
		return d.getMockRoadClosures(bounds), nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var closureResp struct {
		Events []struct {
			ID          string `json:"id"`
			EventType   string `json:"event_type"`
			EventSubtype string `json:"event_subtype"`
			Headline    string `json:"headline"`
			Description string `json:"description"`
			RoadName    string `json:"road_name"`
			Direction   string `json:"direction"`
			Location    struct {
				Latitude  float64 `json:"latitude"`
				Longitude float64 `json:"longitude"`
			} `json:"location"`
			StartTime   string `json:"start_time"`
			EndTime     string `json:"end_time"`
			LanesAffected struct {
				Total  int `json:"total_lanes"`
				Closed int `json:"closed_lanes"`
			} `json:"lanes_affected"`
			DetourInfo struct {
				Available     bool    `json:"available"`
				Description   string  `json:"description"`
				ExtraTime     float64 `json:"extra_time_minutes"`
				ExtraDistance float64 `json:"extra_distance_km"`
			} `json:"detour_info"`
		} `json:"events"`
	}

	if err := json.Unmarshal(body, &closureResp); err != nil {
		return d.getMockRoadClosures(bounds), nil
	}

	// Convert to our format
	var closures []RoadClosure
	for _, event := range closureResp.Events {
		startTime, _ := time.Parse(time.RFC3339, event.StartTime)
		var endTime *time.Time
		if event.EndTime != "" {
			et, _ := time.Parse(time.RFC3339, event.EndTime)
			endTime = &et
		}

		var detour *DetourInfo
		if event.DetourInfo.Available {
			detour = &DetourInfo{
				Route:         "Detour route available",
				Description:   event.DetourInfo.Description,
				ExtraTime:     event.DetourInfo.ExtraTime,
				ExtraDistance: event.DetourInfo.ExtraDistance,
			}
		}

		isPartial := event.LanesAffected.Closed < event.LanesAffected.Total

		closure := RoadClosure{
			ClosureID:   event.ID,
			RoadName:    event.RoadName,
			Location:    Coordinate{Latitude: event.Location.Latitude, Longitude: event.Location.Longitude},
			Direction:   event.Direction,
			Reason:      d.extractReason(event.Description),
			StartTime:   startTime,
			EndTime:     endTime,
			IsPartial:   isPartial,
			LanesClosed: event.LanesAffected.Closed,
			TotalLanes:  event.LanesAffected.Total,
			Detour:      detour,
		}

		closures = append(closures, closure)
	}

	return closures, nil
}

func (d *DOTAPIService) GetConstructionImpact(route string) (*ConstructionImpact, error) {
	// For a given route (polyline or waypoints), determine construction impact
	
	// Mock implementation - in real scenario would analyze route against construction data
	affectedSegments := []AffectedSegment{
		{
			SegmentID:      "SEG_001",
			RoadName:       "I-5 North",
			StartPoint:     "Mile 234",
			EndPoint:       "Mile 237",
			DelayTime:      15.0,
			SpeedReduction: 50.0, // 50% speed reduction
			Description:    "Lane closure for bridge repair work",
		},
		{
			SegmentID:      "SEG_002",
			RoadName:       "US-101 South",
			StartPoint:     "Exit 45",
			EndPoint:       "Exit 42",
			DelayTime:      8.0,
			SpeedReduction: 25.0, // 25% speed reduction
			Description:    "Shoulder work, intermittent delays",
		},
	}

	alternativeRoutes := []AlternativeRoute{
		{
			RouteID:          "ALT_001",
			Description:      "Via I-405 to avoid I-5 construction",
			ExtraTime:        12.0,
			ExtraDistance:    8.5,
			TrafficCondition: "moderate",
		},
	}

	recommendations := []string{
		"Consider departing 20 minutes earlier to account for delays",
		"Use alternative route via I-405 during peak hours",
		"Monitor traffic conditions before departure",
		"Allow extra time for potential stop-and-go traffic",
	}

	totalDelayTime := 0.0
	totalExtraDistance := 0.0
	for _, segment := range affectedSegments {
		totalDelayTime += segment.DelayTime
	}

	return &ConstructionImpact{
		Route:              route,
		TotalDelayTime:     totalDelayTime,
		TotalExtraDistance: totalExtraDistance,
		AffectedSegments:   affectedSegments,
		Recommendations:    recommendations,
		AlternativeRoutes:  alternativeRoutes,
	}, nil
}

// Helper functions
func (d *DOTAPIService) getMockConstructionAlerts(bounds BoundingBox) []ConstructionAlert {
	// Mock construction data for development/testing
	alerts := []ConstructionAlert{
		{
			AlertID:     "CONST_001",
			Title:       "I-5 Bridge Repair Project",
			Description: "Lane closures on I-5 northbound for bridge deck repairs. Expect delays during peak hours.",
			Location:    Coordinate{Latitude: 34.0522, Longitude: -118.2437},
			RoadName:    "Interstate 5",
			StartDate:   time.Now().AddDate(0, 0, -10),
			EndDate:     func() *time.Time { t := time.Now().AddDate(0, 1, 15); return &t }(),
			Severity:    "moderate",
			Impact:      "Lane restrictions, expect 15-20 minute delays",
			Detour: &DetourInfo{
				Route:         "Use I-405 alternative route",
				Description:   "Exit at Sunset Blvd, take I-405 north, rejoin I-5 at Route 170",
				ExtraTime:     18.0,
				ExtraDistance: 12.3,
			},
			Agency:  "Caltrans District 7",
			Contact: "1-800-CALTRANS",
		},
		{
			AlertID:     "CONST_002",
			Title:       "US-101 Resurfacing Project",
			Description: "Overnight resurfacing work on US-101 southbound. Lane closures from 10 PM to 5 AM.",
			Location:    Coordinate{Latitude: 34.1478, Longitude: -118.1445},
			RoadName:    "US Highway 101",
			StartDate:   time.Now().AddDate(0, 0, -3),
			EndDate:     func() *time.Time { t := time.Now().AddDate(0, 0, 7); return &t }(),
			Severity:    "low",
			Impact:      "Overnight lane closures, minimal daytime impact",
			Agency:      "Caltrans District 7",
			Contact:     "1-800-CALTRANS",
		},
	}

	return alerts
}

func (d *DOTAPIService) getMockRoadClosures(bounds BoundingBox) []RoadClosure {
	// Mock road closure data
	closures := []RoadClosure{
		{
			ClosureID:   "CLOSURE_001",
			RoadName:    "CA-2 (Angeles Crest Highway)",
			Location:    Coordinate{Latitude: 34.2367, Longitude: -118.0648},
			Direction:   "Both directions",
			Reason:      "Mudslide cleanup and road repairs",
			StartTime:   time.Now().AddDate(0, 0, -5),
			EndTime:     func() *time.Time { t := time.Now().AddDate(0, 0, 14); return &t }(),
			IsPartial:   false,
			LanesClosed: 2,
			TotalLanes:  2,
			Detour: &DetourInfo{
				Route:         "Via I-210 and CA-14",
				Description:   "Take I-210 west to CA-14 north, significant detour required",
				ExtraTime:     45.0,
				ExtraDistance: 28.7,
			},
		},
		{
			ClosureID:   "CLOSURE_002",
			RoadName:    "I-10 East",
			Location:    Coordinate{Latitude: 34.0224, Longitude: -118.2851},
			Direction:   "Eastbound",
			Reason:      "Emergency bridge inspection",
			StartTime:   time.Now().Add(-2 * time.Hour),
			EndTime:     func() *time.Time { t := time.Now().Add(4 * time.Hour); return &t }(),
			IsPartial:   true,
			LanesClosed: 2,
			TotalLanes:  4,
			Detour: &DetourInfo{
				Route:         "Use right two lanes only",
				Description:   "Traffic reduced to two lanes, expect delays",
				ExtraTime:     25.0,
				ExtraDistance: 0.0,
			},
		},
	}

	return closures
}

func (d *DOTAPIService) mapSeverity(apiSeverity string) string {
	switch strings.ToLower(apiSeverity) {
	case "minor", "low":
		return "low"
	case "moderate", "medium":
		return "moderate"
	case "major", "high", "severe":
		return "high"
	default:
		return "moderate"
	}
}

func (d *DOTAPIService) determineImpact(description string) string {
	desc := strings.ToLower(description)
	
	if strings.Contains(desc, "closure") || strings.Contains(desc, "closed") {
		return "Road closure - use alternative route"
	} else if strings.Contains(desc, "lane") && strings.Contains(desc, "restriction") {
		return "Lane restrictions - expect delays"
	} else if strings.Contains(desc, "overnight") || strings.Contains(desc, "night") {
		return "Overnight work - minimal daytime impact"
	} else if strings.Contains(desc, "delay") {
		return "Expect significant delays"
	}
	
	return "Monitor conditions and allow extra time"
}

func (d *DOTAPIService) extractReason(description string) string {
	desc := strings.ToLower(description)
	
	if strings.Contains(desc, "construction") {
		return "construction"
	} else if strings.Contains(desc, "accident") {
		return "accident"
	} else if strings.Contains(desc, "weather") || strings.Contains(desc, "storm") {
		return "weather"
	} else if strings.Contains(desc, "maintenance") || strings.Contains(desc, "repair") {
		return "maintenance"
	} else if strings.Contains(desc, "emergency") {
		return "emergency"
	}
	
	return "unknown"
}