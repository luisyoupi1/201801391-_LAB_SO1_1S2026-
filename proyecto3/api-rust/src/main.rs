use axum::{
    extract::State,
    http::StatusCode,
    response::{IntoResponse, Response},
    routing::{get, post},
    Json, Router,
};
use chrono::{DateTime, Utc};
use serde::{Deserialize, Serialize};
use std::{env, sync::{atomic::{AtomicU64, Ordering}, Arc}};
use tower_http::trace::TraceLayer;
use tracing::{info, warn};
use uuid::Uuid;

#[derive(Clone)]
struct AppState {
    client: reqwest::Client,
    go_api_url: String,
    dapr_url: String,
    carnet: String,
    grpc_total: Arc<AtomicU64>,
    dapr_total: Arc<AtomicU64>,
    errors_total: Arc<AtomicU64>,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct WarReport {
    #[serde(default = "new_id")]
    id: String,
    country: String,
    warplanes_in_air: i32,
    warships_in_water: i32,
    timestamp: DateTime<Utc>,
}

#[derive(Serialize)]
struct Accepted {
    status: &'static str,
    id: String,
    route: &'static str,
}

#[derive(Debug)]
struct ApiError(StatusCode, String);

impl IntoResponse for ApiError {
    fn into_response(self) -> Response {
        (self.0, Json(serde_json::json!({"error": self.1}))).into_response()
    }
}

fn new_id() -> String { Uuid::new_v4().to_string() }

fn validate(report: &WarReport) -> Result<(), ApiError> {
    const COUNTRIES: [&str; 5] = ["USA", "RUS", "CHN", "ESP", "GTM"];
    if !COUNTRIES.contains(&report.country.as_str()) {
        return Err(ApiError(StatusCode::BAD_REQUEST, "country debe ser USA, RUS, CHN, ESP o GTM".into()));
    }
    if !(0..=50).contains(&report.warplanes_in_air) {
        return Err(ApiError(StatusCode::BAD_REQUEST, "warplanes_in_air debe estar entre 0 y 50".into()));
    }
    if !(0..=30).contains(&report.warships_in_water) {
        return Err(ApiError(StatusCode::BAD_REQUEST, "warships_in_water debe estar entre 0 y 30".into()));
    }
    Ok(())
}

async fn grpc_route(State(state): State<AppState>, Json(report): Json<WarReport>) -> Result<impl IntoResponse, ApiError> {
    validate(&report)?;
    let response = state.client.post(&state.go_api_url).json(&report).send().await
        .map_err(|error| ApiError(StatusCode::BAD_GATEWAY, format!("Go API no disponible: {error}")))?;
    if !response.status().is_success() {
        state.errors_total.fetch_add(1, Ordering::Relaxed);
        return Err(ApiError(StatusCode::BAD_GATEWAY, format!("Go API respondió {}", response.status())));
    }
    state.grpc_total.fetch_add(1, Ordering::Relaxed);
    info!(id = report.id, country = report.country, "reporte aceptado por flujo gRPC");
    Ok((StatusCode::ACCEPTED, Json(Accepted { status: "accepted", id: report.id, route: "grpc" })))
}

async fn dapr_route(State(state): State<AppState>, Json(report): Json<WarReport>) -> Result<impl IntoResponse, ApiError> {
    validate(&report)?;
    let url = format!("{}/v1.0/publish/rabbitmq-pubsub/war-reports", state.dapr_url);
    let response = state.client.post(url).json(&report).send().await
        .map_err(|error| ApiError(StatusCode::SERVICE_UNAVAILABLE, format!("Dapr no disponible: {error}")))?;
    if !response.status().is_success() {
        state.errors_total.fetch_add(1, Ordering::Relaxed);
        return Err(ApiError(StatusCode::BAD_GATEWAY, format!("Dapr respondió {}", response.status())));
    }
    state.dapr_total.fetch_add(1, Ordering::Relaxed);
    Ok((StatusCode::ACCEPTED, Json(Accepted { status: "accepted", id: report.id, route: "dapr" })))
}

async fn health(State(state): State<AppState>) -> Json<serde_json::Value> {
    Json(serde_json::json!({"status":"ok", "service":"rust-api", "carnet":state.carnet}))
}

async fn metrics(State(state): State<AppState>) -> String {
    format!(
        "# TYPE mum_rust_requests_total counter\n\
         mum_rust_requests_total{{route=\"grpc\"}} {}\n\
         mum_rust_requests_total{{route=\"dapr\"}} {}\n\
         # TYPE mum_rust_errors_total counter\n\
         mum_rust_errors_total {}\n",
        state.grpc_total.load(Ordering::Relaxed),
        state.dapr_total.load(Ordering::Relaxed),
        state.errors_total.load(Ordering::Relaxed),
    )
}

#[tokio::main]
async fn main() {
    tracing_subscriber::fmt().with_env_filter(
        tracing_subscriber::EnvFilter::try_from_default_env().unwrap_or_else(|_| "info,tower_http=info".into())
    ).init();
    let state = AppState {
        client: reqwest::Client::new(),
        go_api_url: env::var("GO_API_URL").unwrap_or_else(|_| "http://go-api:8081/ingest".into()),
        dapr_url: env::var("DAPR_HTTP_URL").unwrap_or_else(|_| "http://127.0.0.1:3500".into()),
        carnet: env::var("CARNET").unwrap_or_else(|_| "201801391".into()),
        grpc_total: Arc::new(AtomicU64::new(0)),
        dapr_total: Arc::new(AtomicU64::new(0)),
        errors_total: Arc::new(AtomicU64::new(0)),
    };
    let app = Router::new()
        .route("/health", get(health))
        .route("/metrics", get(metrics))
        .route("/grpc-201801391", post(grpc_route))
        .route("/dapr-201801391", post(dapr_route))
        .layer(TraceLayer::new_for_http())
        .with_state(state);
    let addr = env::var("HTTP_ADDR").unwrap_or_else(|_| "0.0.0.0:8080".into());
    let listener = tokio::net::TcpListener::bind(&addr).await.expect("bind HTTP_ADDR");
    info!(%addr, "API Rust iniciada");
    if let Err(error) = axum::serve(listener, app).with_graceful_shutdown(shutdown()).await {
        warn!(%error, "servidor finalizado con error");
    }
}

async fn shutdown() {
    let _ = tokio::signal::ctrl_c().await;
}

#[cfg(test)]
mod tests {
    use super::*;
    fn report(country: &str, planes: i32, ships: i32) -> WarReport {
        WarReport { id: new_id(), country: country.into(), warplanes_in_air: planes, warships_in_water: ships, timestamp: Utc::now() }
    }
    #[test]
    fn accepts_valid_report() { assert!(validate(&report("USA", 50, 30)).is_ok()); }
    #[test]
    fn rejects_invalid_values() {
        assert!(validate(&report("XXX", 1, 1)).is_err());
        assert!(validate(&report("USA", 51, 1)).is_err());
        assert!(validate(&report("USA", 1, 31)).is_err());
    }
}
