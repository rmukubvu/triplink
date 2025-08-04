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

// External API service interfaces and implementations

// TrafficAPIService interface for traffic-related API calls
type TrafficAPIService interface {
	GetTrafficConditions(origin, destination string) (*TrafficInfo, error)
	GetRouteMatrix(origins, destinations []string) (*RouteMatrix, error)
	GetTrafficIncidents(bounds BoundingBox) ([]TrafficIncident, error)
}

// WeatherAPIService interface for weather-related API calls
type WeatherAPIService interface {
	GetCurrentWeather(lat, lng float64) (*WeatherCondition, error)
	GetWeatherForecast(lat, lng float64, hours int) ([]WeatherCondition, error)
	GetWeatherAlerts(bounds BoundingBox) ([]WeatherAlert, error)
	GetRouteWeather(waypoints []Coordinate) ([]WeatherCondition, error)
}

// MappingAPIService interface for mapping and routing
type MappingAPIService interface {
	GetOptimizedRoute(request RouteRequest) (*RouteResponse, error)
	GetDirections(origin, destination string, options DirectionOptions) (*DirectionsResponse, error)
	GeocodeAddress(address string) (*GeocodeResult, error)
	ReverseGeocode(lat, lng float64) (*GeocodeResult, error)
}

// Common structures
type Coordinate struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type BoundingBox struct {
	NorthEast Coordinate `json:"north_east"`
	SouthWest Coordinate `json:"south_west"`
}

// Traffic-related structures
type TrafficInfo struct {
	AverageSpeed     float64           `json:"average_speed"`
	CongestionLevel  string            `json:"congestion_level"`
	DelayMinutes     float64           `json:"delay_minutes"`
	Incidents        []TrafficIncident `json:"incidents"`
	LastUpdated      time.Time         `json:"last_updated"`
}

type TrafficIncident struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	Description string     `json:"description"`
	Severity    string     `json:"severity"`
	Location    Coordinate `json:"location"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	DelayImpact float64    `json:"delay_impact"`
}

type RouteMatrix struct {
	Origins      []string             `json:"origins"`
	Destinations []string             `json:"destinations"`
	Rows         []RouteMatrixRow     `json:"rows"`
	Status       string               `json:"status"`
}

type RouteMatrixRow struct {
	Elements []RouteMatrixElement `json:"elements"`
}

type RouteMatrixElement struct {
	Distance    DistanceValue `json:"distance"`
	Duration    DurationValue `json:"duration"`
	TrafficDuration DurationValue `json:"duration_in_traffic"`
	Status      string        `json:"status"`
}

type DistanceValue struct {
	Text  string `json:"text"`
	Value int    `json:"value"` // meters
}

type DurationValue struct {
	Text  string `json:"text"`
	Value int    `json:"value"` // seconds
}

// Weather-related structures
type WeatherCondition struct {
	Location      Coordinate `json:"location"`
	Temperature   float64    `json:"temperature"`   // Celsius
	Humidity      float64    `json:"humidity"`      // percentage
	WindSpeed     float64    `json:"wind_speed"`    // km/h
	WindDirection float64    `json:"wind_direction"` // degrees
	Precipitation float64    `json:"precipitation"` // mm
	Visibility    float64    `json:"visibility"`    // km
	Condition     string     `json:"condition"`
	Icon          string     `json:"icon"`
	Timestamp     time.Time  `json:"timestamp"`
}

type WeatherAlert struct {
	ID          string     `json:"id"`
	Type        string     `json:"type"`
	Severity    string     `json:"severity"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     time.Time  `json:"end_time"`
	Areas       []string   `json:"areas"`
}

// Mapping-related structures
type RouteRequest struct {
	Origin      string           `json:"origin"`
	Destination string           `json:"destination"`
	Waypoints   []string         `json:"waypoints,omitempty"`
	Options     DirectionOptions `json:"options"`
}

type DirectionOptions struct {
	Mode           string    `json:"mode"`            // driving, walking, cycling
	AvoidTolls     bool      `json:"avoid_tolls"`
	AvoidHighways  bool      `json:"avoid_highways"`
	AvoidFerries   bool      `json:"avoid_ferries"`
	DepartureTime  *time.Time `json:"departure_time,omitempty"`
	TrafficModel   string    `json:"traffic_model"`   // best_guess, pessimistic, optimistic
	Units          string    `json:"units"`           // metric, imperial
}

type RouteResponse struct {
	Routes []Route `json:"routes"`
	Status string  `json:"status"`
}

type Route struct {
	Summary      string    `json:"summary"`
	Legs         []RouteLeg `json:"legs"`
	Distance     DistanceValue `json:"distance"`
	Duration     DurationValue `json:"duration"`
	TrafficDuration DurationValue `json:"duration_in_traffic"`
	OverviewPolyline string   `json:"overview_polyline"`
	Warnings     []string  `json:"warnings"`
	Copyrights   string    `json:"copyrights"`
}

type RouteLeg struct {
	StartAddress string        `json:"start_address"`
	EndAddress   string        `json:"end_address"`
	StartLocation Coordinate   `json:"start_location"`
	EndLocation  Coordinate    `json:"end_location"`
	Distance     DistanceValue `json:"distance"`
	Duration     DurationValue `json:"duration"`
	Steps        []RouteStep   `json:"steps"`
}

type RouteStep struct {
	HTMLInstructions string        `json:"html_instructions"`
	Distance         DistanceValue `json:"distance"`
	Duration         DurationValue `json:"duration"`
	StartLocation    Coordinate    `json:"start_location"`
	EndLocation      Coordinate    `json:"end_location"`
	Polyline         string        `json:"polyline"`
	TravelMode       string        `json:"travel_mode"`
	Maneuver         string        `json:"maneuver,omitempty"`
}

type DirectionsResponse struct {
	Routes           []Route `json:"routes"`
	GeocodedWaypoints []GeocodeResult `json:"geocoded_waypoints"`
	Status           string  `json:"status"`
	ErrorMessage     string  `json:"error_message,omitempty"`
}

type GeocodeResult struct {
	Address          string     `json:"formatted_address"`
	Geometry         Geometry   `json:"geometry"`
	PlaceID          string     `json:"place_id"`
	Types            []string   `json:"types"`
	AddressComponents []AddressComponent `json:"address_components"`
}

type Geometry struct {
	Location     Coordinate   `json:"location"`
	LocationType string       `json:"location_type"`
	Viewport     BoundingBox  `json:"viewport"`
	Bounds       *BoundingBox `json:"bounds,omitempty"`
}

type AddressComponent struct {
	LongName  string   `json:"long_name"`
	ShortName string   `json:"short_name"`
	Types     []string `json:"types"`
}

// Google Maps API Service Implementation
type GoogleMapsService struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

func NewGoogleMapsService() *GoogleMapsService {
	return &GoogleMapsService{
		APIKey:  os.Getenv("GOOGLE_MAPS_API_KEY"),
		BaseURL: "https://maps.googleapis.com/maps/api",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (g *GoogleMapsService) GetTrafficConditions(origin, destination string) (*TrafficInfo, error) {
	// Use Distance Matrix API with traffic data
	url := fmt.Sprintf("%s/distancematrix/json?origins=%s&destinations=%s&departure_time=now&traffic_model=best_guess&key=%s",
		g.BaseURL, url.QueryEscape(origin), url.QueryEscape(destination), g.APIKey)

	resp, err := g.HTTPClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to get traffic conditions: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var matrixResp RouteMatrix
	if err := json.Unmarshal(body, &matrixResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Convert matrix response to traffic info
	if len(matrixResp.Rows) > 0 && len(matrixResp.Rows[0].Elements) > 0 {
		element := matrixResp.Rows[0].Elements[0]
		
		// Calculate congestion level based on traffic delay
		normalDuration := float64(element.Duration.Value)
		trafficDuration := float64(element.TrafficDuration.Value)
		delayMinutes := (trafficDuration - normalDuration) / 60

		congestionLevel := "light"
		if delayMinutes > 30 {
			congestionLevel = "heavy"
		} else if delayMinutes > 15 {
			congestionLevel = "moderate"
		}

		// Estimate average speed (assuming highway speeds)
		distance := float64(element.Distance.Value) / 1000 // km
		avgSpeed := distance / (trafficDuration / 3600)    // km/h

		return &TrafficInfo{
			AverageSpeed:    avgSpeed,
			CongestionLevel: congestionLevel,
			DelayMinutes:    delayMinutes,
			Incidents:       []TrafficIncident{}, // Would need separate API call
			LastUpdated:     time.Now(),
		}, nil
	}

	return nil, fmt.Errorf("no traffic data available")
}

func (g *GoogleMapsService) GetRouteMatrix(origins, destinations []string) (*RouteMatrix, error) {
	originsStr := url.QueryEscape(fmt.Sprintf("%v", origins)[1:len(fmt.Sprintf("%v", origins))-1])
	destinationsStr := url.QueryEscape(fmt.Sprintf("%v", destinations)[1:len(fmt.Sprintf("%v", destinations))-1])
	
	apiURL := fmt.Sprintf("%s/distancematrix/json?origins=%s&destinations=%s&departure_time=now&traffic_model=best_guess&key=%s",
		g.BaseURL, originsStr, destinationsStr, g.APIKey)

	resp, err := g.HTTPClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get route matrix: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var matrix RouteMatrix
	if err := json.Unmarshal(body, &matrix); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &matrix, nil
}

func (g *GoogleMapsService) GetTrafficIncidents(bounds BoundingBox) ([]TrafficIncident, error) {
	// Google Maps doesn't provide a direct traffic incidents API
	// This would typically use HERE API or other traffic services
	// For now, return empty slice
	return []TrafficIncident{}, nil
}

func (g *GoogleMapsService) GetOptimizedRoute(request RouteRequest) (*RouteResponse, error) {
	// Build waypoints string
	waypointsStr := ""
	if len(request.Waypoints) > 0 {
		waypoints := make([]string, len(request.Waypoints))
		for i, wp := range request.Waypoints {
			waypoints[i] = "optimize:true|" + wp
		}
		waypointsStr = "&waypoints=" + url.QueryEscape(fmt.Sprintf("%v", waypoints)[1:len(fmt.Sprintf("%v", waypoints))-1])
	}

	// Build options
	avoidStr := ""
	if request.Options.AvoidTolls {
		avoidStr = "tolls"
	}
	if request.Options.AvoidHighways {
		if avoidStr != "" {
			avoidStr += "|"
		}
		avoidStr += "highways"
	}
	if request.Options.AvoidFerries {
		if avoidStr != "" {
			avoidStr += "|"
		}
		avoidStr += "ferries"
	}
	if avoidStr != "" {
		avoidStr = "&avoid=" + avoidStr
	}

	departureTime := ""
	if request.Options.DepartureTime != nil {
		departureTime = fmt.Sprintf("&departure_time=%d", request.Options.DepartureTime.Unix())
	}

	apiURL := fmt.Sprintf("%s/directions/json?origin=%s&destination=%s%s%s%s&traffic_model=%s&key=%s",
		g.BaseURL,
		url.QueryEscape(request.Origin),
		url.QueryEscape(request.Destination),
		waypointsStr,
		avoidStr,
		departureTime,
		request.Options.TrafficModel,
		g.APIKey)

	resp, err := g.HTTPClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get optimized route: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var directionsResp DirectionsResponse
	if err := json.Unmarshal(body, &directionsResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &RouteResponse{
		Routes: directionsResp.Routes,
		Status: directionsResp.Status,
	}, nil
}

func (g *GoogleMapsService) GetDirections(origin, destination string, options DirectionOptions) (*DirectionsResponse, error) {
	// Build avoid parameter
	avoidStr := ""
	if options.AvoidTolls {
		avoidStr = "tolls"
	}
	if options.AvoidHighways {
		if avoidStr != "" {
			avoidStr += "|"
		}
		avoidStr += "highways"
	}
	if options.AvoidFerries {
		if avoidStr != "" {
			avoidStr += "|"
		}
		avoidStr += "ferries"
	}

	params := url.Values{}
	params.Set("origin", origin)
	params.Set("destination", destination)
	params.Set("mode", options.Mode)
	params.Set("units", options.Units)
	params.Set("key", g.APIKey)
	
	if avoidStr != "" {
		params.Set("avoid", avoidStr)
	}
	if options.DepartureTime != nil {
		params.Set("departure_time", fmt.Sprintf("%d", options.DepartureTime.Unix()))
		params.Set("traffic_model", options.TrafficModel)
	}

	apiURL := fmt.Sprintf("%s/directions/json?%s", g.BaseURL, params.Encode())

	resp, err := g.HTTPClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get directions: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var directionsResp DirectionsResponse
	if err := json.Unmarshal(body, &directionsResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &directionsResp, nil
}

func (g *GoogleMapsService) GeocodeAddress(address string) (*GeocodeResult, error) {
	apiURL := fmt.Sprintf("%s/geocode/json?address=%s&key=%s",
		g.BaseURL, url.QueryEscape(address), g.APIKey)

	resp, err := g.HTTPClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to geocode address: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var geocodeResp struct {
		Results []GeocodeResult `json:"results"`
		Status  string          `json:"status"`
	}

	if err := json.Unmarshal(body, &geocodeResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(geocodeResp.Results) > 0 {
		return &geocodeResp.Results[0], nil
	}

	return nil, fmt.Errorf("no geocoding results found")
}

func (g *GoogleMapsService) ReverseGeocode(lat, lng float64) (*GeocodeResult, error) {
	latlng := fmt.Sprintf("%f,%f", lat, lng)
	apiURL := fmt.Sprintf("%s/geocode/json?latlng=%s&key=%s",
		g.BaseURL, latlng, g.APIKey)

	resp, err := g.HTTPClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to reverse geocode: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var geocodeResp struct {
		Results []GeocodeResult `json:"results"`
		Status  string          `json:"status"`
	}

	if err := json.Unmarshal(body, &geocodeResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if len(geocodeResp.Results) > 0 {
		return &geocodeResp.Results[0], nil
	}

	return nil, fmt.Errorf("no reverse geocoding results found")
}

// OpenWeatherMap API Service Implementation
type OpenWeatherMapService struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
}

func NewOpenWeatherMapService() *OpenWeatherMapService {
	return &OpenWeatherMapService{
		APIKey:  os.Getenv("OPENWEATHERMAP_API_KEY"),
		BaseURL: "https://api.openweathermap.org/data/2.5",
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (w *OpenWeatherMapService) GetCurrentWeather(lat, lng float64) (*WeatherCondition, error) {
	apiURL := fmt.Sprintf("%s/weather?lat=%f&lon=%f&units=metric&appid=%s",
		w.BaseURL, lat, lng, w.APIKey)

	resp, err := w.HTTPClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get current weather: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var weatherResp struct {
		Weather []struct {
			Main        string `json:"main"`
			Description string `json:"description"`
			Icon        string `json:"icon"`
		} `json:"weather"`
		Main struct {
			Temp     float64 `json:"temp"`
			Humidity float64 `json:"humidity"`
		} `json:"main"`
		Wind struct {
			Speed float64 `json:"speed"`
			Deg   float64 `json:"deg"`
		} `json:"wind"`
		Rain struct {
			OneH float64 `json:"1h"`
		} `json:"rain"`
		Visibility int       `json:"visibility"`
		Dt         int64     `json:"dt"`
		Coord      struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		} `json:"coord"`
	}

	if err := json.Unmarshal(body, &weatherResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	condition := &WeatherCondition{
		Location: Coordinate{
			Latitude:  weatherResp.Coord.Lat,
			Longitude: weatherResp.Coord.Lon,
		},
		Temperature:   weatherResp.Main.Temp,
		Humidity:      weatherResp.Main.Humidity,
		WindSpeed:     weatherResp.Wind.Speed * 3.6, // Convert m/s to km/h
		WindDirection: weatherResp.Wind.Deg,
		Precipitation: weatherResp.Rain.OneH,
		Visibility:    float64(weatherResp.Visibility) / 1000, // Convert m to km
		Timestamp:     time.Unix(weatherResp.Dt, 0),
	}

	if len(weatherResp.Weather) > 0 {
		condition.Condition = weatherResp.Weather[0].Main
		condition.Icon = weatherResp.Weather[0].Icon
	}

	return condition, nil
}

func (w *OpenWeatherMapService) GetWeatherForecast(lat, lng float64, hours int) ([]WeatherCondition, error) {
	apiURL := fmt.Sprintf("%s/forecast?lat=%f&lon=%f&units=metric&appid=%s",
		w.BaseURL, lat, lng, w.APIKey)

	resp, err := w.HTTPClient.Get(apiURL)
	if err != nil {
		return nil, fmt.Errorf("failed to get weather forecast: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var forecastResp struct {
		List []struct {
			Dt   int64 `json:"dt"`
			Main struct {
				Temp     float64 `json:"temp"`
				Humidity float64 `json:"humidity"`
			} `json:"main"`
			Weather []struct {
				Main        string `json:"main"`
				Description string `json:"description"`
				Icon        string `json:"icon"`
			} `json:"weather"`
			Wind struct {
				Speed float64 `json:"speed"`
				Deg   float64 `json:"deg"`
			} `json:"wind"`
			Rain struct {
				ThreeH float64 `json:"3h"`
			} `json:"rain"`
			Visibility int `json:"visibility"`
		} `json:"list"`
		City struct {
			Coord struct {
				Lat float64 `json:"lat"`
				Lon float64 `json:"lon"`
			} `json:"coord"`
		} `json:"city"`
	}

	if err := json.Unmarshal(body, &forecastResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	var conditions []WeatherCondition
	maxItems := hours / 3 // OpenWeatherMap forecast is in 3-hour intervals
	if maxItems > len(forecastResp.List) {
		maxItems = len(forecastResp.List)
	}

	for i := 0; i < maxItems; i++ {
		item := forecastResp.List[i]
		condition := WeatherCondition{
			Location: Coordinate{
				Latitude:  forecastResp.City.Coord.Lat,
				Longitude: forecastResp.City.Coord.Lon,
			},
			Temperature:   item.Main.Temp,
			Humidity:      item.Main.Humidity,
			WindSpeed:     item.Wind.Speed * 3.6, // Convert m/s to km/h
			WindDirection: item.Wind.Deg,
			Precipitation: item.Rain.ThreeH,
			Visibility:    float64(item.Visibility) / 1000, // Convert m to km
			Timestamp:     time.Unix(item.Dt, 0),
		}

		if len(item.Weather) > 0 {
			condition.Condition = item.Weather[0].Main
			condition.Icon = item.Weather[0].Icon
		}

		conditions = append(conditions, condition)
	}

	return conditions, nil
}

func (w *OpenWeatherMapService) GetWeatherAlerts(bounds BoundingBox) ([]WeatherAlert, error) {
	// OpenWeatherMap One Call API provides alerts, but requires a different endpoint
	// For now, return empty slice
	return []WeatherAlert{}, nil
}

func (w *OpenWeatherMapService) GetRouteWeather(waypoints []Coordinate) ([]WeatherCondition, error) {
	var conditions []WeatherCondition

	for _, waypoint := range waypoints {
		condition, err := w.GetCurrentWeather(waypoint.Latitude, waypoint.Longitude)
		if err != nil {
			continue // Skip waypoints with weather data errors
		}
		conditions = append(conditions, *condition)
	}

	return conditions, nil
}

// Additional API service interfaces for fuel, tolls, and construction
type FuelPriceAPIService interface {
	GetFuelPrices(location Coordinate, radius float64) (*FuelPriceInfo, error)
	GetFuelPricesAlongRoute(waypoints []Coordinate) ([]FuelStation, error)
	GetFuelPriceTrends(location Coordinate, days int) (*FuelPriceTrend, error)
}

type TollAPIService interface {
	GetTollRates(origin, destination string) (*TollInfo, error)
	GetTollStationsAlongRoute(waypoints []Coordinate) ([]TollStation, error)
	CalculateTollCosts(routePolyline string) (*TollCostBreakdown, error)
}

type ConstructionAPIService interface {
	GetConstructionAlerts(bounds BoundingBox) ([]ConstructionAlert, error)
	GetRoadClosures(bounds BoundingBox) ([]RoadClosure, error)
	GetConstructionImpact(route string) (*ConstructionImpact, error)
}

// Additional structures for fuel, toll, and construction data
type FuelPriceInfo struct {
	Location      Coordinate    `json:"location"`
	AveragePrice  float64       `json:"average_price"`
	CheapestPrice float64       `json:"cheapest_price"`
	MostExpensive float64       `json:"most_expensive_price"`
	Stations      []FuelStation `json:"stations"`
	LastUpdated   time.Time     `json:"last_updated"`
}

type FuelStation struct {
	StationID   string     `json:"station_id"`
	Name        string     `json:"name"`
	Brand       string     `json:"brand"`
	Location    Coordinate `json:"location"`
	Address     string     `json:"address"`
	RegularPrice float64   `json:"regular_price"`
	PremiumPrice float64   `json:"premium_price"`
	DieselPrice  float64   `json:"diesel_price"`
	Distance     float64   `json:"distance"` // km from reference point
	IsOpen       bool       `json:"is_open"`
	Amenities    []string   `json:"amenities"`
	LastUpdated  time.Time  `json:"last_updated"`
}

type FuelPriceTrend struct {
	Location     Coordinate          `json:"location"`
	Period       string              `json:"period"`
	CurrentPrice float64             `json:"current_price"`
	AveragePrice float64             `json:"average_price"`
	TrendData    []FuelPriceDataPoint `json:"trend_data"`
	Prediction   *FuelPriceForecast   `json:"prediction,omitempty"`
}

type FuelPriceDataPoint struct {
	Date  time.Time `json:"date"`
	Price float64   `json:"price"`
}

type FuelPriceForecast struct {
	NextWeekPrice   float64 `json:"next_week_price"`
	NextMonthPrice  float64 `json:"next_month_price"`
	Confidence      float64 `json:"confidence"`
	TrendDirection  string  `json:"trend_direction"`
}

type TollInfo struct {
	Route         string        `json:"route"`
	TotalCost     float64       `json:"total_cost"`
	Currency      string        `json:"currency"`
	TollStations  []TollStation `json:"toll_stations"`
	PaymentMethods []string     `json:"payment_methods"`
	EstimatedTime float64       `json:"estimated_time"` // minutes
	LastUpdated   time.Time     `json:"last_updated"`
}

type TollStation struct {
	StationID    string     `json:"station_id"`
	Name         string     `json:"name"`
	Location     Coordinate `json:"location"`
	Highway      string     `json:"highway"`
	Direction    string     `json:"direction"`
	Cost         float64    `json:"cost"`
	VehicleTypes []string   `json:"vehicle_types"`
	Operator     string     `json:"operator"`
	IsElectronic bool       `json:"is_electronic"`
}

type TollCostBreakdown struct {
	TotalCost     float64            `json:"total_cost"`
	DetailedCosts []TollSegmentCost  `json:"detailed_costs"`
	Discounts     []TollDiscount     `json:"discounts,omitempty"`
	PaymentOptions []string          `json:"payment_options"`
}

type TollSegmentCost struct {
	SegmentName   string  `json:"segment_name"`
	EntryPoint    string  `json:"entry_point"`
	ExitPoint     string  `json:"exit_point"`
	Distance      float64 `json:"distance"`
	Cost          float64 `json:"cost"`
	VehicleClass  string  `json:"vehicle_class"`
}

type TollDiscount struct {
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
	Percentage  float64 `json:"percentage"`
}

type ConstructionAlert struct {
	AlertID     string     `json:"alert_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Location    Coordinate `json:"location"`
	RoadName    string     `json:"road_name"`
	StartDate   time.Time  `json:"start_date"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Severity    string     `json:"severity"`
	Impact      string     `json:"impact"`
	Detour      *DetourInfo `json:"detour,omitempty"`
	Agency      string     `json:"agency"`
	Contact     string     `json:"contact"`
}

type RoadClosure struct {
	ClosureID   string     `json:"closure_id"`
	RoadName    string     `json:"road_name"`
	Location    Coordinate `json:"location"`
	Direction   string     `json:"direction"`
	Reason      string     `json:"reason"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     *time.Time `json:"end_time,omitempty"`
	IsPartial   bool       `json:"is_partial"`
	LanesClosed int        `json:"lanes_closed"`
	TotalLanes  int        `json:"total_lanes"`
	Detour      *DetourInfo `json:"detour,omitempty"`
}

type DetourInfo struct {
	Route       string  `json:"route"`
	Description string  `json:"description"`
	ExtraTime   float64 `json:"extra_time"`   // minutes
	ExtraDistance float64 `json:"extra_distance"` // km
}

type ConstructionImpact struct {
	Route           string              `json:"route"`
	TotalDelayTime  float64             `json:"total_delay_time"`  // minutes
	TotalExtraDistance float64          `json:"total_extra_distance"` // km
	AffectedSegments []AffectedSegment  `json:"affected_segments"`
	Recommendations []string            `json:"recommendations"`
	AlternativeRoutes []AlternativeRoute `json:"alternative_routes,omitempty"`
}

type AffectedSegment struct {
	SegmentID   string  `json:"segment_id"`
	RoadName    string  `json:"road_name"`
	StartPoint  string  `json:"start_point"`
	EndPoint    string  `json:"end_point"`
	DelayTime   float64 `json:"delay_time"`  // minutes
	SpeedReduction float64 `json:"speed_reduction"` // percentage
	Description string  `json:"description"`
}

type AlternativeRoute struct {
	RouteID     string  `json:"route_id"`
	Description string  `json:"description"`
	ExtraTime   float64 `json:"extra_time"`     // minutes
	ExtraDistance float64 `json:"extra_distance"` // km
	TrafficCondition string `json:"traffic_condition"`
}

// Route optimization request structure
type RouteOptimizationRequest struct {
	Origin      string                    `json:"origin"`
	Destination string                    `json:"destination"`
	Waypoints   []string                  `json:"waypoints,omitempty"`
	Options     RouteOptimizationOptions  `json:"options"`
}

type RouteOptimizationOptions struct {
	AvoidTolls    bool       `json:"avoid_tolls"`
	AvoidHighways bool       `json:"avoid_highways"`
	AvoidFerries  bool       `json:"avoid_ferries"`
	DepartureTime *time.Time `json:"departure_time,omitempty"`
	TrafficModel  string     `json:"traffic_model"`
}

// Service factory function
func NewExternalAPIServices() (TrafficAPIService, WeatherAPIService, MappingAPIService, FuelPriceAPIService, TollAPIService, ConstructionAPIService) {
	googleMaps := NewGoogleMapsService()
	openWeather := NewOpenWeatherMapService()
	fuelService := NewFuelAPIService()
	tollService := NewTollAPIService()
	constructionService := NewDOTAPIService()

	return googleMaps, openWeather, googleMaps, fuelService, tollService, constructionService
}