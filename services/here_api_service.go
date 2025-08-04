package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"
)

// HERE API Service Implementation for traffic incidents and advanced routing
type HEREAPIService struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

func NewHEREAPIService() *HEREAPIService {
	return &HEREAPIService{
		APIKey:  os.Getenv("HERE_API_KEY"),
		BaseURL: "https://api.here.com/v1",
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// HERE-specific structures
type HERETrafficIncident struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Criticality int                    `json:"criticality"`
	StartTime   time.Time              `json:"start_time"`
	EndTime     *time.Time             `json:"end_time,omitempty"`
	Location    HEREIncidentLocation   `json:"location"`
	RoadClosed  bool                   `json:"road_closed"`
	Verified    bool                   `json:"verified"`
	Length      float64                `json:"length"` // meters
}

type HEREIncidentLocation struct {
	Polyline    string      `json:"polyline"`
	RoadName    string      `json:"road_name"`
	Direction   string      `json:"direction"`
	Coordinates []Coordinate `json:"coordinates"`
}

type HEREFlowData struct {
	Results []HEREFlowResult `json:"results"`
}

type HEREFlowResult struct {
	Location      Coordinate `json:"location"`
	CurrentFlow   HEREFlow   `json:"current_flow"`
	FreeFlow      HEREFlow   `json:"free_flow"`
	JamFactor     float64    `json:"jam_factor"`     // 0-10 scale
	Confidence    float64    `json:"confidence"`     // 0-1 scale
	TrafficState  string     `json:"traffic_state"`  // "flowing", "heavy", "blocked"
}

type HEREFlow struct {
	Speed           float64 `json:"speed"`            // km/h
	SpeedUncapped   float64 `json:"speed_uncapped"`   // km/h without speed limit cap
	SpeedLimit      float64 `json:"speed_limit"`      // km/h
	TravelTime      int     `json:"travel_time"`      // seconds
	TrafficLight    bool    `json:"traffic_light"`
}

type HERERouteResponse struct {
	Routes []HERERoute `json:"routes"`
	Notices []HERENotice `json:"notices,omitempty"`
}

type HERERoute struct {
	ID       string         `json:"id"`
	Sections []HERESection  `json:"sections"`
	Summary  HERESummary    `json:"summary"`
}

type HERESection struct {
	ID         string           `json:"id"`
	Type       string           `json:"type"`
	Departure  HEREPlace        `json:"departure"`
	Arrival    HEREPlace        `json:"arrival"`
	Summary    HERESummary      `json:"summary"`
	Polyline   string           `json:"polyline"`
	Actions    []HEREAction     `json:"actions,omitempty"`
	Incidents  []HEREIncidentRef `json:"incidents,omitempty"`
}

type HEREPlace struct {
	Type        string     `json:"type"`
	Location    Coordinate `json:"location"`
	OriginalLocation *Coordinate `json:"original_location,omitempty"`
}

type HERESummary struct {
	Duration       int             `json:"duration"`        // seconds
	Length         int             `json:"length"`          // meters
	BaseDuration   int             `json:"base_duration"`   // seconds without traffic
	TrafficDelay   int             `json:"traffic_delay"`   // seconds
	TypicalDuration int            `json:"typical_duration"` // seconds
	Text           string          `json:"text"`
	TrafficTime    *HERETrafficTime `json:"traffic_time,omitempty"`
}

type HERETrafficTime struct {
	Best     int `json:"best"`     // seconds
	Worst    int `json:"worst"`    // seconds
	Typical  int `json:"typical"`  // seconds
}

type HEREAction struct {
	Action      string     `json:"action"`
	Direction   string     `json:"direction"`
	Severity    string     `json:"severity"`
	Offset      int        `json:"offset"`
	Length      int        `json:"length"`
	Instruction string     `json:"instruction"`
	NextRoad    string     `json:"next_road,omitempty"`
}

type HEREIncidentRef struct {
	ID      string `json:"id"`
	Offset  int    `json:"offset"`
	Length  int    `json:"length"`
}

type HERENotice struct {
	Title   string `json:"title"`
	Code    string `json:"code"`
	Severity string `json:"severity"`
}

// Implement TrafficAPIService interface
func (h *HEREAPIService) GetTrafficConditions(origin, destination string) (*TrafficInfo, error) {
	// Get route with traffic information
	route, err := h.getRouteWithTraffic(origin, destination)
	if err != nil {
		return nil, err
	}

	if len(route.Routes) == 0 {
		return nil, fmt.Errorf("no routes found")
	}

	mainRoute := route.Routes[0]
	summary := mainRoute.Summary

	// Calculate average speed
	totalDistance := float64(summary.Length) / 1000 // Convert to km
	totalTime := float64(summary.Duration) / 3600   // Convert to hours
	avgSpeed := totalDistance / totalTime

	// Determine congestion level based on traffic delay
	delayMinutes := float64(summary.TrafficDelay) / 60
	congestionLevel := "light"
	if delayMinutes > 30 {
		congestionLevel = "heavy"
	} else if delayMinutes > 15 {
		congestionLevel = "moderate"
	}

	// Convert HERE incidents to our format
	var incidents []TrafficIncident
	for _, section := range mainRoute.Sections {
		for _, incidentRef := range section.Incidents {
			// Would need to fetch detailed incident data
			incident := TrafficIncident{
				ID:          incidentRef.ID,
				Type:        "unknown", // Would be fetched from incident details
				Description: "Traffic incident",
				Severity:    "moderate",
				Location:    section.Departure.Location,
				StartTime:   time.Now(),
				DelayImpact: delayMinutes / float64(len(section.Incidents)),
			}
			incidents = append(incidents, incident)
		}
	}

	return &TrafficInfo{
		AverageSpeed:    avgSpeed,
		CongestionLevel: congestionLevel,
		DelayMinutes:    delayMinutes,
		Incidents:       incidents,
		LastUpdated:     time.Now(),
	}, nil
}

func (h *HEREAPIService) GetTrafficIncidents(bounds BoundingBox) ([]TrafficIncident, error) {
	// HERE Traffic API v7 - Get incidents in bounding box
	bbox := fmt.Sprintf("%f,%f,%f,%f", 
		bounds.SouthWest.Longitude, bounds.SouthWest.Latitude,
		bounds.NorthEast.Longitude, bounds.NorthEast.Latitude)

	apiURL := fmt.Sprintf("https://data.traffic.hereapi.com/v7/incidents?bbox=%s&apikey=%s",
		bbox, h.APIKey)

	resp, err := h.HTTPClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get traffic incidents: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var incidentsResp struct {
		Results []HERETrafficIncident `json:"results"`
	}

	if err := json.Unmarshal(body, &incidentsResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert HERE incidents to our format
	var incidents []TrafficIncident
	for _, hereIncident := range incidentsResp.Results {
		incidentType := "other"
		switch hereIncident.Type {
		case "ACCIDENT":
			incidentType = "accident"
		case "CONSTRUCTION":
			incidentType = "construction"
		case "ROAD_CLOSURE":
			incidentType = "road_closure"
		}

		severity := "low"
		if hereIncident.Criticality >= 7 {
			severity = "high"
		} else if hereIncident.Criticality >= 4 {
			severity = "moderate"
		}

		// Get primary location from coordinates
		var location Coordinate
		if len(hereIncident.Location.Coordinates) > 0 {
			location = hereIncident.Location.Coordinates[0]
		}

		incident := TrafficIncident{
			ID:          hereIncident.ID,
			Type:        incidentType,
			Description: hereIncident.Description,
			Severity:    severity,
			Location:    location,
			StartTime:   hereIncident.StartTime,
			EndTime:     hereIncident.EndTime,
			DelayImpact: float64(hereIncident.Criticality) * 5, // Estimate delay impact
		}

		incidents = append(incidents, incident)
	}

	return incidents, nil
}

func (h *HEREAPIService) GetRouteMatrix(origins, destinations []string) (*RouteMatrix, error) {
	// HERE Matrix Routing API v8
	matrixURL := "https://matrix.router.hereapi.com/v8/matrix"

	// Build request body
	requestBody := map[string]interface{}{
		"origins":      h.buildMatrixLocations(origins),
		"destinations": h.buildMatrixLocations(destinations),
		"regionDefinition": map[string]interface{}{
			"type": "world",
		},
		"matrixAttributes": []string{"distances", "travelTimes", "routeAttributes"},
		"departureTime":    time.Now().Format(time.RFC3339),
	}

	_, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", matrixURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.URL.RawQuery = url.Values{"apikey": {h.APIKey}}.Encode()

	resp, err := h.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var matrixResp struct {
		Matrix struct {
			NumOrigins      int `json:"numOrigins"`
			NumDestinations int `json:"numDestinations"`
			TravelTimes     []int `json:"travelTimes"`     // seconds
			Distances       []int `json:"distances"`       // meters
			ErrorCodes      []int `json:"errorCodes"`
		} `json:"matrix"`
	}

	if err := json.Unmarshal(body, &matrixResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert HERE matrix response to our format
	matrix := &RouteMatrix{
		Origins:      origins,
		Destinations: destinations,
		Status:       "OK",
	}

	// Build rows and elements
	numOrigins := matrixResp.Matrix.NumOrigins
	numDestinations := matrixResp.Matrix.NumDestinations

	for i := 0; i < numOrigins; i++ {
		row := RouteMatrixRow{}
		for j := 0; j < numDestinations; j++ {
			index := i*numDestinations + j
			
			var element RouteMatrixElement
			if index < len(matrixResp.Matrix.TravelTimes) && index < len(matrixResp.Matrix.Distances) {
				if matrixResp.Matrix.ErrorCodes[index] == 0 {
					element = RouteMatrixElement{
						Distance: DistanceValue{
							Value: matrixResp.Matrix.Distances[index],
							Text:  fmt.Sprintf("%.1f km", float64(matrixResp.Matrix.Distances[index])/1000),
						},
						Duration: DurationValue{
							Value: matrixResp.Matrix.TravelTimes[index],
							Text:  fmt.Sprintf("%d mins", matrixResp.Matrix.TravelTimes[index]/60),
						},
						TrafficDuration: DurationValue{
							Value: matrixResp.Matrix.TravelTimes[index], // HERE includes traffic by default
							Text:  fmt.Sprintf("%d mins", matrixResp.Matrix.TravelTimes[index]/60),
						},
						Status: "OK",
					}
				} else {
					element = RouteMatrixElement{
						Status: "NOT_FOUND",
					}
				}
			}
			row.Elements = append(row.Elements, element)
		}
		matrix.Rows = append(matrix.Rows, row)
	}

	return matrix, nil
}

// Helper methods
func (h *HEREAPIService) getRouteWithTraffic(origin, destination string) (*HERERouteResponse, error) {
	// HERE Routing API v8
	apiURL := "https://router.hereapi.com/v8/routes"

	params := url.Values{}
	params.Set("origin", origin)
	params.Set("destination", destination)
	params.Set("transportMode", "car")
	params.Set("departureTime", time.Now().Format(time.RFC3339))
	params.Set("return", "summary,polyline,actions,incidents,tollSummary")
	params.Set("apikey", h.APIKey)

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := h.HTTPClient.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get route: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var routeResp HERERouteResponse
	if err := json.Unmarshal(body, &routeResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &routeResp, nil
}

func (h *HEREAPIService) buildMatrixLocations(addresses []string) []map[string]interface{} {
	var locations []map[string]interface{}

	for _, address := range addresses {
		location := map[string]interface{}{
			"address": map[string]string{
				"text": address,
			},
		}
		locations = append(locations, location)
	}

	return locations
}

// Enhanced route optimization with HERE advanced features
func (h *HEREAPIService) GetAdvancedRouteOptimization(request RouteOptimizationRequest) (*HERERouteResponse, error) {
	apiURL := "https://router.hereapi.com/v8/routes"

	params := url.Values{}
	params.Set("origin", request.Origin)
	params.Set("destination", request.Destination)
	params.Set("transportMode", "car")
	params.Set("routingMode", "fast") // or "short" for shortest distance
	
	// Add waypoints if present
	if len(request.Waypoints) > 0 {
		for i, waypoint := range request.Waypoints {
			params.Set(fmt.Sprintf("via%d", i), waypoint)
		}
	}

	// Add optimization preferences
	if request.Options.AvoidTolls {
		params.Add("avoid[features]", "tollRoad")
	}
	if request.Options.AvoidHighways {
		params.Add("avoid[features]", "controlledAccessHighway")
	}
	if request.Options.AvoidFerries {
		params.Add("avoid[features]", "ferry")
	}

	// Traffic-aware routing
	if request.Options.DepartureTime != nil {
		params.Set("departureTime", request.Options.DepartureTime.Format(time.RFC3339))
	} else {
		params.Set("departureTime", time.Now().Format(time.RFC3339))
	}

	// Return detailed information
	params.Set("return", "summary,polyline,actions,incidents,tollSummary,typicalDuration,instructions")
	params.Set("apikey", h.APIKey)

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	resp, err := h.HTTPClient.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get advanced route: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var routeResp HERERouteResponse
	if err := json.Unmarshal(body, &routeResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &routeResp, nil
}

// Real-time traffic flow data
func (h *HEREAPIService) GetTrafficFlow(bounds BoundingBox) (*HEREFlowData, error) {
	bbox := fmt.Sprintf("%f,%f,%f,%f", 
		bounds.SouthWest.Longitude, bounds.SouthWest.Latitude,
		bounds.NorthEast.Longitude, bounds.NorthEast.Latitude)

	apiURL := fmt.Sprintf("https://data.traffic.hereapi.com/v7/flow?bbox=%s&apikey=%s",
		bbox, h.APIKey)

	resp, err := h.HTTPClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get traffic flow: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var flowData HEREFlowData
	if err := json.Unmarshal(body, &flowData); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &flowData, nil
}