-- Analytics and ML Predictions Database Schema Extensions
-- Migration: 20240115_create_analytics_tables.sql

-- ============================================================================
-- ANALYTICS TABLES
-- ============================================================================

-- On-time delivery analytics storage
CREATE TABLE IF NOT EXISTS delivery_analytics (
    id SERIAL PRIMARY KEY,
    analysis_date DATE NOT NULL,
    time_window VARCHAR(20) NOT NULL, -- 'daily', 'weekly', 'monthly'
    total_deliveries INTEGER NOT NULL DEFAULT 0,
    on_time_deliveries INTEGER NOT NULL DEFAULT 0,
    early_deliveries INTEGER NOT NULL DEFAULT 0,
    late_deliveries INTEGER NOT NULL DEFAULT 0,
    critical_delays INTEGER NOT NULL DEFAULT 0,
    on_time_percentage DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    average_delay_minutes DECIMAL(8,2) NOT NULL DEFAULT 0.00,
    average_delivery_time_hours DECIMAL(8,2) NOT NULL DEFAULT 0.00,
    improvement_percentage DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    filters JSONB, -- Store filter criteria used
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Customer satisfaction analytics
CREATE TABLE IF NOT EXISTS satisfaction_analytics (
    id SERIAL PRIMARY KEY,
    analysis_date DATE NOT NULL,
    time_window VARCHAR(20) NOT NULL,
    overall_score DECIMAL(3,2) NOT NULL DEFAULT 0.00,
    total_responses INTEGER NOT NULL DEFAULT 0,
    response_rate DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    score_improvement DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    nps_score INTEGER NOT NULL DEFAULT 0,
    satisfied_customers INTEGER NOT NULL DEFAULT 0,
    neutral_customers INTEGER NOT NULL DEFAULT 0,
    dissatisfied_customers INTEGER NOT NULL DEFAULT 0,
    retention_rate DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    churn_rate DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    average_response_time_hours DECIMAL(8,2) NOT NULL DEFAULT 0.00,
    resolution_rate DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    filters JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Load matching efficiency analytics
CREATE TABLE IF NOT EXISTS load_matching_analytics (
    id SERIAL PRIMARY KEY,
    analysis_date DATE NOT NULL,
    time_window VARCHAR(20) NOT NULL,
    total_loads INTEGER NOT NULL DEFAULT 0,
    matched_loads INTEGER NOT NULL DEFAULT 0,
    matching_rate DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    average_match_time_hours DECIMAL(8,2) NOT NULL DEFAULT 0.00,
    optimal_matches INTEGER NOT NULL DEFAULT 0,
    suboptimal_matches INTEGER NOT NULL DEFAULT 0,
    unmatched_loads INTEGER NOT NULL DEFAULT 0,
    utilization_rate DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    revenue_efficiency DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    cost_efficiency DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    improvement_opportunity DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    filters JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Capacity utilization analytics
CREATE TABLE IF NOT EXISTS capacity_analytics (
    id SERIAL PRIMARY KEY,
    analysis_date DATE NOT NULL,
    time_window VARCHAR(20) NOT NULL,
    total_capacity_kg DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    utilized_capacity_kg DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    utilization_rate DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    available_capacity_kg DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    peak_utilization DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    average_utilization DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    capacity_efficiency DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    revenue_per_capacity_unit DECIMAL(10,4) NOT NULL DEFAULT 0.00,
    cost_per_capacity_unit DECIMAL(10,4) NOT NULL DEFAULT 0.00,
    capacity_trend VARCHAR(20) DEFAULT 'stable',
    demand_vs_capacity DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    forecasted_demand_kg DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    filters JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Delay analysis and incidents
CREATE TABLE IF NOT EXISTS delay_analytics (
    id SERIAL PRIMARY KEY,
    analysis_date DATE NOT NULL,
    time_window VARCHAR(20) NOT NULL,
    total_delays INTEGER NOT NULL DEFAULT 0,
    average_delay_duration_minutes DECIMAL(8,2) NOT NULL DEFAULT 0.00,
    delay_frequency DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    total_delay_time_minutes DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    cost_of_delays DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    customer_impact_score DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    on_time_performance DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    delay_trend VARCHAR(20) DEFAULT 'stable',
    recurrent_delays INTEGER NOT NULL DEFAULT 0,
    preventable_delays INTEGER NOT NULL DEFAULT 0,
    mitigated_delays INTEGER NOT NULL DEFAULT 0,
    improvement_opportunity DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    filters JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Detailed delay incidents
CREATE TABLE IF NOT EXISTS delay_incidents (
    id SERIAL PRIMARY KEY,
    incident_id VARCHAR(50) UNIQUE NOT NULL,
    trip_id INTEGER REFERENCES trips(id),
    vehicle_id INTEGER REFERENCES vehicles(id),
    driver_id INTEGER REFERENCES users(id),
    route_name VARCHAR(255),
    origin VARCHAR(255) NOT NULL,
    destination VARCHAR(255) NOT NULL,
    scheduled_departure TIMESTAMP NOT NULL,
    actual_departure TIMESTAMP,
    scheduled_arrival TIMESTAMP NOT NULL,
    actual_arrival TIMESTAMP,
    delay_duration_minutes DECIMAL(8,2) NOT NULL DEFAULT 0.00,
    delay_type VARCHAR(50) NOT NULL, -- 'traffic', 'weather', 'mechanical', 'loading', 'other'
    delay_category VARCHAR(50) NOT NULL, -- 'external', 'internal', 'customer'
    severity VARCHAR(20) NOT NULL, -- 'low', 'medium', 'high', 'critical'
    root_cause TEXT,
    customer_notified BOOLEAN DEFAULT FALSE,
    resolution_time_minutes DECIMAL(8,2) DEFAULT 0.00,
    preventable BOOLEAN DEFAULT FALSE,
    recurrent BOOLEAN DEFAULT FALSE,
    cost_impact DECIMAL(10,2) DEFAULT 0.00,
    customer_impact VARCHAR(20) DEFAULT 'low', -- 'low', 'medium', 'high'
    status VARCHAR(20) DEFAULT 'open', -- 'open', 'investigating', 'resolved', 'closed'
    reported_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    resolved_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Route performance analytics
CREATE TABLE IF NOT EXISTS route_performance (
    id SERIAL PRIMARY KEY,
    route_id VARCHAR(100) NOT NULL,
    route_name VARCHAR(255) NOT NULL,
    origin VARCHAR(255) NOT NULL,
    destination VARCHAR(255) NOT NULL,
    distance_km DECIMAL(8,2) NOT NULL,
    analysis_date DATE NOT NULL,
    time_window VARCHAR(20) NOT NULL,
    total_trips INTEGER NOT NULL DEFAULT 0,
    completed_trips INTEGER NOT NULL DEFAULT 0,
    on_time_trips INTEGER NOT NULL DEFAULT 0,
    delayed_trips INTEGER NOT NULL DEFAULT 0,
    cancelled_trips INTEGER NOT NULL DEFAULT 0,
    on_time_percentage DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    average_delay_minutes DECIMAL(8,2) NOT NULL DEFAULT 0.00,
    average_trip_duration_hours DECIMAL(8,2) NOT NULL DEFAULT 0.00,
    fuel_efficiency_kmpl DECIMAL(6,2) NOT NULL DEFAULT 0.00,
    cost_per_km DECIMAL(8,4) NOT NULL DEFAULT 0.00,
    revenue_per_km DECIMAL(8,4) NOT NULL DEFAULT 0.00,
    profitability DECIMAL(8,4) NOT NULL DEFAULT 0.00,
    customer_satisfaction DECIMAL(3,2) NOT NULL DEFAULT 0.00,
    risk_level VARCHAR(20) DEFAULT 'low',
    improvement_trend VARCHAR(20) DEFAULT 'stable',
    filters JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Driver performance analytics
CREATE TABLE IF NOT EXISTS driver_performance (
    id SERIAL PRIMARY KEY,
    driver_id INTEGER REFERENCES users(id),
    driver_name VARCHAR(255) NOT NULL,
    analysis_date DATE NOT NULL,
    time_window VARCHAR(20) NOT NULL,
    total_trips INTEGER NOT NULL DEFAULT 0,
    completed_trips INTEGER NOT NULL DEFAULT 0,
    on_time_trips INTEGER NOT NULL DEFAULT 0,
    early_trips INTEGER NOT NULL DEFAULT 0,
    late_trips INTEGER NOT NULL DEFAULT 0,
    cancelled_trips INTEGER NOT NULL DEFAULT 0,
    on_time_percentage DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    average_delay_minutes DECIMAL(8,2) NOT NULL DEFAULT 0.00,
    total_distance_km DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    total_driving_hours DECIMAL(8,2) NOT NULL DEFAULT 0.00,
    fuel_efficiency_kmpl DECIMAL(6,2) NOT NULL DEFAULT 0.00,
    safety_score DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    customer_rating DECIMAL(3,2) NOT NULL DEFAULT 0.00,
    performance_rating DECIMAL(3,2) NOT NULL DEFAULT 0.00,
    improvement_trend VARCHAR(20) DEFAULT 'stable',
    training_recommended BOOLEAN DEFAULT FALSE,
    performance_issues JSONB, -- Store performance issues as JSON
    achievements JSONB, -- Store achievements and milestones
    filters JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Vehicle utilization analytics
CREATE TABLE IF NOT EXISTS vehicle_utilization (
    id SERIAL PRIMARY KEY,
    vehicle_id INTEGER REFERENCES vehicles(id),
    vehicle_number VARCHAR(100) NOT NULL,
    vehicle_type VARCHAR(50) NOT NULL,
    analysis_date DATE NOT NULL,
    time_window VARCHAR(20) NOT NULL,
    max_capacity_kg DECIMAL(10,2) NOT NULL,
    max_volume_m3 DECIMAL(10,2) NOT NULL,
    average_load_kg DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    average_volume_m3 DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    weight_utilization DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    volume_utilization DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    overall_utilization DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    total_trips INTEGER NOT NULL DEFAULT 0,
    active_days INTEGER NOT NULL DEFAULT 0,
    idle_days INTEGER NOT NULL DEFAULT 0,
    maintenance_days INTEGER NOT NULL DEFAULT 0,
    total_distance_km DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    fuel_consumed_liters DECIMAL(8,2) NOT NULL DEFAULT 0.00,
    fuel_efficiency_kmpl DECIMAL(6,2) NOT NULL DEFAULT 0.00,
    revenue DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    operating_costs DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    maintenance_costs DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    profitability DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    efficiency_rating VARCHAR(20) DEFAULT 'average',
    utilization_trend VARCHAR(20) DEFAULT 'stable',
    location VARCHAR(255),
    status VARCHAR(20) DEFAULT 'active',
    filters JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- ML PREDICTIONS TABLES
-- ============================================================================

-- Sentiment analysis results
CREATE TABLE IF NOT EXISTS sentiment_predictions (
    id SERIAL PRIMARY KEY,
    prediction_id VARCHAR(100) UNIQUE NOT NULL,
    source_type VARCHAR(50) NOT NULL, -- 'review', 'feedback', 'message', 'survey'
    source_id INTEGER NOT NULL, -- ID of the source record
    text_content TEXT NOT NULL,
    text_hash VARCHAR(64) NOT NULL, -- MD5 hash for caching
    sentiment VARCHAR(20) NOT NULL, -- 'positive', 'negative', 'neutral'
    confidence DECIMAL(4,3) NOT NULL, -- 0.000 to 1.000
    positive_score DECIMAL(4,3) NOT NULL DEFAULT 0.000,
    negative_score DECIMAL(4,3) NOT NULL DEFAULT 0.000,
    neutral_score DECIMAL(4,3) NOT NULL DEFAULT 0.000,
    model_used VARCHAR(100) NOT NULL,
    model_version VARCHAR(20) DEFAULT '1.0',
    processing_time_ms INTEGER DEFAULT 0,
    language_detected VARCHAR(10) DEFAULT 'en',
    keywords JSONB, -- Extracted keywords and their weights
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Delay predictions
CREATE TABLE IF NOT EXISTS delay_predictions (
    id SERIAL PRIMARY KEY,
    prediction_id VARCHAR(100) UNIQUE NOT NULL,
    route_id VARCHAR(100),
    trip_id INTEGER REFERENCES trips(id),
    origin VARCHAR(255) NOT NULL,
    destination VARCHAR(255) NOT NULL,
    scheduled_departure TIMESTAMP NOT NULL,
    predicted_delay_minutes DECIMAL(8,2) NOT NULL DEFAULT 0.00,
    confidence DECIMAL(4,3) NOT NULL,
    risk_level VARCHAR(20) NOT NULL, -- 'low', 'medium', 'high', 'critical'
    factors_analyzed JSONB NOT NULL, -- JSON array of prediction factors
    weather_impact DECIMAL(5,2) DEFAULT 0.00,
    traffic_impact DECIMAL(5,2) DEFAULT 0.00,
    route_complexity_impact DECIMAL(5,2) DEFAULT 0.00,
    historical_pattern_impact DECIMAL(5,2) DEFAULT 0.00,
    vehicle_performance_impact DECIMAL(5,2) DEFAULT 0.00,
    driver_performance_impact DECIMAL(5,2) DEFAULT 0.00,
    model_used VARCHAR(100) NOT NULL,
    model_version VARCHAR(20) DEFAULT '1.0',
    processing_time_ms INTEGER DEFAULT 0,
    recommendations JSONB, -- Recommendations to mitigate delay
    actual_delay_minutes DECIMAL(8,2), -- Filled after trip completion
    prediction_accuracy DECIMAL(5,2), -- Calculated after trip completion
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Customer satisfaction predictions
CREATE TABLE IF NOT EXISTS satisfaction_predictions (
    id SERIAL PRIMARY KEY,
    prediction_id VARCHAR(100) UNIQUE NOT NULL,
    customer_id INTEGER REFERENCES users(id),
    trip_id INTEGER REFERENCES trips(id),
    predicted_rating DECIMAL(3,2) NOT NULL, -- 1.00 to 5.00
    predicted_nps INTEGER NOT NULL, -- -100 to +100
    confidence DECIMAL(4,3) NOT NULL,
    risk_factors JSONB NOT NULL, -- Array of risk factors
    improvement_areas JSONB, -- Suggested improvement areas
    delivery_timeliness_score DECIMAL(5,2) DEFAULT 0.00,
    communication_quality_score DECIMAL(5,2) DEFAULT 0.00,
    driver_behavior_score DECIMAL(5,2) DEFAULT 0.00,
    vehicle_condition_score DECIMAL(5,2) DEFAULT 0.00,
    pricing_satisfaction_score DECIMAL(5,2) DEFAULT 0.00,
    overall_experience_score DECIMAL(5,2) DEFAULT 0.00,
    model_used VARCHAR(100) NOT NULL,
    model_version VARCHAR(20) DEFAULT '1.0',
    processing_time_ms INTEGER DEFAULT 0,
    actual_rating DECIMAL(3,2), -- Filled after customer feedback
    actual_nps INTEGER, -- Filled after customer feedback
    prediction_accuracy DECIMAL(5,2), -- Calculated after feedback received
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Text classification results
CREATE TABLE IF NOT EXISTS text_classifications (
    id SERIAL PRIMARY KEY,
    classification_id VARCHAR(100) UNIQUE NOT NULL,
    source_type VARCHAR(50) NOT NULL,
    source_id INTEGER NOT NULL,
    text_content TEXT NOT NULL,
    text_hash VARCHAR(64) NOT NULL,
    top_category VARCHAR(100) NOT NULL,
    confidence DECIMAL(4,3) NOT NULL,
    categories JSONB NOT NULL, -- All categories with scores
    model_used VARCHAR(100) NOT NULL,
    model_version VARCHAR(20) DEFAULT '1.0',
    processing_time_ms INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Route optimization ML results
CREATE TABLE IF NOT EXISTS route_optimization_ml (
    id SERIAL PRIMARY KEY,
    optimization_id VARCHAR(100) UNIQUE NOT NULL,
    route_id VARCHAR(100) NOT NULL,
    origin VARCHAR(255) NOT NULL,
    destination VARCHAR(255) NOT NULL,
    waypoints JSONB, -- Array of waypoint coordinates
    optimal_departure TIMESTAMP NOT NULL,
    predicted_duration_minutes DECIMAL(8,2) NOT NULL,
    predicted_distance_km DECIMAL(8,2) NOT NULL,
    predicted_fuel_consumption_liters DECIMAL(8,2) NOT NULL,
    predicted_fuel_cost DECIMAL(10,2) NOT NULL,
    predicted_toll_cost DECIMAL(10,2) NOT NULL,
    predicted_total_cost DECIMAL(10,2) NOT NULL,
    traffic_prediction JSONB NOT NULL, -- Traffic analysis results
    weather_prediction JSONB NOT NULL, -- Weather impact analysis
    optimization_score DECIMAL(5,2) NOT NULL, -- 0-100 score
    alternative_routes JSONB, -- Alternative route options
    confidence DECIMAL(4,3) NOT NULL,
    model_used VARCHAR(100) NOT NULL,
    model_version VARCHAR(20) DEFAULT '1.0',
    processing_time_ms INTEGER DEFAULT 0,
    actual_duration_minutes DECIMAL(8,2), -- Filled after trip completion
    actual_distance_km DECIMAL(8,2), -- Filled after trip completion
    actual_fuel_cost DECIMAL(10,2), -- Filled after trip completion
    prediction_accuracy DECIMAL(5,2), -- Calculated after trip completion
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ML model performance tracking
CREATE TABLE IF NOT EXISTS ml_model_performance (
    id SERIAL PRIMARY KEY,
    model_name VARCHAR(100) NOT NULL,
    model_version VARCHAR(20) NOT NULL,
    model_type VARCHAR(50) NOT NULL, -- 'sentiment', 'delay', 'satisfaction', 'classification', 'optimization'
    prediction_date DATE NOT NULL,
    total_predictions INTEGER NOT NULL DEFAULT 0,
    successful_predictions INTEGER NOT NULL DEFAULT 0,
    failed_predictions INTEGER NOT NULL DEFAULT 0,
    average_confidence DECIMAL(4,3) NOT NULL DEFAULT 0.000,
    average_processing_time_ms DECIMAL(8,2) NOT NULL DEFAULT 0.00,
    accuracy_score DECIMAL(5,2), -- Only available for models with feedback
    precision_score DECIMAL(5,2),
    recall_score DECIMAL(5,2),
    f1_score DECIMAL(5,2),
    error_rate DECIMAL(5,2) NOT NULL DEFAULT 0.00,
    model_drift_score DECIMAL(5,2), -- Measure of model performance degradation
    last_retrained TIMESTAMP,
    needs_retraining BOOLEAN DEFAULT FALSE,
    performance_notes TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- ============================================================================
-- INDEXES FOR PERFORMANCE
-- ============================================================================

-- Analytics table indexes
CREATE INDEX IF NOT EXISTS idx_delivery_analytics_date ON delivery_analytics(analysis_date);
CREATE INDEX IF NOT EXISTS idx_delivery_analytics_window ON delivery_analytics(time_window);
CREATE INDEX IF NOT EXISTS idx_satisfaction_analytics_date ON satisfaction_analytics(analysis_date);
CREATE INDEX IF NOT EXISTS idx_satisfaction_analytics_window ON satisfaction_analytics(time_window);
CREATE INDEX IF NOT EXISTS idx_load_matching_analytics_date ON load_matching_analytics(analysis_date);
CREATE INDEX IF NOT EXISTS idx_capacity_analytics_date ON capacity_analytics(analysis_date);
CREATE INDEX IF NOT EXISTS idx_delay_analytics_date ON delay_analytics(analysis_date);

-- Delay incidents indexes
CREATE INDEX IF NOT EXISTS idx_delay_incidents_trip_id ON delay_incidents(trip_id);
CREATE INDEX IF NOT EXISTS idx_delay_incidents_vehicle_id ON delay_incidents(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_delay_incidents_driver_id ON delay_incidents(driver_id);
CREATE INDEX IF NOT EXISTS idx_delay_incidents_date ON delay_incidents(scheduled_departure);
CREATE INDEX IF NOT EXISTS idx_delay_incidents_type ON delay_incidents(delay_type);
CREATE INDEX IF NOT EXISTS idx_delay_incidents_severity ON delay_incidents(severity);

-- Performance analytics indexes
CREATE INDEX IF NOT EXISTS idx_route_performance_route_id ON route_performance(route_id);
CREATE INDEX IF NOT EXISTS idx_route_performance_date ON route_performance(analysis_date);
CREATE INDEX IF NOT EXISTS idx_driver_performance_driver_id ON driver_performance(driver_id);
CREATE INDEX IF NOT EXISTS idx_driver_performance_date ON driver_performance(analysis_date);
CREATE INDEX IF NOT EXISTS idx_vehicle_utilization_vehicle_id ON vehicle_utilization(vehicle_id);
CREATE INDEX IF NOT EXISTS idx_vehicle_utilization_date ON vehicle_utilization(analysis_date);

-- ML predictions indexes
CREATE INDEX IF NOT EXISTS idx_sentiment_predictions_hash ON sentiment_predictions(text_hash);
CREATE INDEX IF NOT EXISTS idx_sentiment_predictions_source ON sentiment_predictions(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_sentiment_predictions_sentiment ON sentiment_predictions(sentiment);
CREATE INDEX IF NOT EXISTS idx_delay_predictions_trip_id ON delay_predictions(trip_id);
CREATE INDEX IF NOT EXISTS idx_delay_predictions_route_id ON delay_predictions(route_id);
CREATE INDEX IF NOT EXISTS idx_delay_predictions_departure ON delay_predictions(scheduled_departure);
CREATE INDEX IF NOT EXISTS idx_satisfaction_predictions_customer_id ON satisfaction_predictions(customer_id);
CREATE INDEX IF NOT EXISTS idx_satisfaction_predictions_trip_id ON satisfaction_predictions(trip_id);
CREATE INDEX IF NOT EXISTS idx_text_classifications_hash ON text_classifications(text_hash);
CREATE INDEX IF NOT EXISTS idx_text_classifications_source ON text_classifications(source_type, source_id);
CREATE INDEX IF NOT EXISTS idx_route_optimization_ml_route_id ON route_optimization_ml(route_id);
CREATE INDEX IF NOT EXISTS idx_ml_model_performance_model ON ml_model_performance(model_name, model_version);
CREATE INDEX IF NOT EXISTS idx_ml_model_performance_date ON ml_model_performance(prediction_date);

-- ============================================================================
-- VIEWS FOR COMMON QUERIES
-- ============================================================================

-- Real-time delivery performance view
CREATE OR REPLACE VIEW delivery_performance_summary AS
SELECT 
    'today' as period,
    COUNT(*) as total_deliveries,
    COUNT(CASE WHEN status = 'COMPLETED' AND 
                    ABS(EXTRACT(EPOCH FROM (actual_arrival - estimated_arrival))/60) <= 30 
          THEN 1 END) as on_time_deliveries,
    ROUND(
        (COUNT(CASE WHEN status = 'COMPLETED' AND 
                         ABS(EXTRACT(EPOCH FROM (actual_arrival - estimated_arrival))/60) <= 30 
               THEN 1 END) * 100.0 / NULLIF(COUNT(CASE WHEN status = 'COMPLETED' THEN 1 END), 0)), 2
    ) as on_time_percentage,
    ROUND(
        AVG(CASE WHEN status = 'COMPLETED' AND actual_arrival > estimated_arrival 
            THEN EXTRACT(EPOCH FROM (actual_arrival - estimated_arrival))/60 
            ELSE 0 END), 2
    ) as average_delay_minutes
FROM trips 
WHERE departure_date >= CURRENT_DATE;

-- Customer satisfaction summary view
CREATE OR REPLACE VIEW satisfaction_summary AS
SELECT 
    'current' as period,
    COUNT(*) as total_reviews,
    ROUND(AVG(rating), 2) as average_rating,
    COUNT(CASE WHEN rating >= 4 THEN 1 END) as satisfied_customers,
    COUNT(CASE WHEN rating = 3 THEN 1 END) as neutral_customers,
    COUNT(CASE WHEN rating <= 2 THEN 1 END) as dissatisfied_customers,
    ROUND(
        ((COUNT(CASE WHEN rating >= 4 THEN 1 END) * 100.0 / NULLIF(COUNT(*), 0)) - 
         (COUNT(CASE WHEN rating <= 2 THEN 1 END) * 100.0 / NULLIF(COUNT(*), 0))), 0
    ) as nps_score
FROM reviews 
WHERE created_at >= CURRENT_DATE - INTERVAL '30 days';

-- ML predictions accuracy view
CREATE OR REPLACE VIEW ml_accuracy_summary AS
SELECT 
    model_name,
    model_type,
    COUNT(*) as total_predictions,
    ROUND(AVG(CASE WHEN prediction_accuracy IS NOT NULL THEN prediction_accuracy END), 2) as average_accuracy,
    ROUND(AVG(confidence), 3) as average_confidence,
    ROUND(AVG(processing_time_ms), 2) as average_processing_time
FROM (
    SELECT model_used as model_name, 'delay' as model_type, prediction_accuracy, confidence, processing_time_ms FROM delay_predictions
    UNION ALL
    SELECT model_used as model_name, 'satisfaction' as model_type, prediction_accuracy, confidence, processing_time_ms FROM satisfaction_predictions
) combined_predictions
GROUP BY model_name, model_type;

-- COMMIT the migration
COMMIT;