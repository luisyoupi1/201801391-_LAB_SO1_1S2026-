package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

const testCarnet = "201801391"

func TestHealthContract(t *testing.T) {
	fixedTime := time.Date(2026, time.August, 3, 14, 30, 0, 0, time.UTC)
	app := mustApp(t, Config{
		Name:    "API1",
		VM:      "VM1",
		Carnet:  testCarnet,
		Address: ":8081",
		Now:     func() time.Time { return fixedTime },
	})

	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()
	app.Handler().ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("código HTTP inesperado: %d", recorder.Code)
	}
	var response healthResponse
	decode(t, recorder, &response)
	if response.Status != "UP" || response.Message != "API1 is Ready" || response.VM != "VM1" || response.Carnet != testCarnet {
		t.Fatalf("respuesta /health incorrecta: %+v", response)
	}
	if response.Timestamp != "2026-08-03T14:30:00Z" {
		t.Fatalf("timestamp inesperado: %s", response.Timestamp)
	}
}

func TestCrossAPICallWhenTargetIsUp(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Fatalf("ruta consultada inesperada: %s", r.URL.Path)
		}
		writeJSON(w, http.StatusOK, healthResponse{Status: "UP"})
	}))
	defer target.Close()

	app := mustApp(t, Config{
		Name:    "API2",
		VM:      "VM1",
		Carnet:  testCarnet,
		Address: ":8082",
		Targets: []Target{{Name: "API1", VM: "VM1", BaseURL: target.URL}},
	})

	request := httptest.NewRequest(http.MethodGet, "/api2/201801391/call-api1", nil)
	recorder := httptest.NewRecorder()
	app.Handler().ServeHTTP(recorder, request)

	var response callResponse
	decode(t, recorder, &response)
	if !response.Connection || response.APIName != "API1" || response.Carnet != testCarnet {
		t.Fatalf("respuesta de comunicación incorrecta: %+v", response)
	}
	if response.Message != "The API1 located on the VM1 is working" {
		t.Fatalf("mensaje inesperado: %s", response.Message)
	}
}

func TestCrossAPICallHandlesBadStatus(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{Status: "DOWN"})
	}))
	defer target.Close()

	app := mustApp(t, Config{
		Name:    "API3",
		VM:      "VM2",
		Carnet:  testCarnet,
		Address: ":8083",
		Targets: []Target{{Name: "API1", VM: "VM1", BaseURL: target.URL}},
	})

	request := httptest.NewRequest(http.MethodGet, "/api3/201801391/call-api1", nil)
	recorder := httptest.NewRecorder()
	app.Handler().ServeHTTP(recorder, request)

	var response callResponse
	decode(t, recorder, &response)
	if response.Connection {
		t.Fatalf("se esperaba connection=false: %+v", response)
	}
	if response.Message != "ERROR: The API1 located on the VM1 is not working" {
		t.Fatalf("mensaje de error inesperado: %s", response.Message)
	}
}

func TestCrossAPICallHandlesUnavailableTarget(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, healthResponse{Status: "UP"})
	}))
	targetURL := target.URL
	target.Close()

	app := mustApp(t, Config{
		Name:    "API1",
		VM:      "VM1",
		Carnet:  testCarnet,
		Address: ":8081",
		Targets: []Target{{Name: "API3", VM: "VM2", BaseURL: targetURL}},
	})

	request := httptest.NewRequest(http.MethodGet, "/api1/201801391/call-api3", nil)
	recorder := httptest.NewRecorder()
	app.Handler().ServeHTTP(recorder, request)

	var response callResponse
	decode(t, recorder, &response)
	if response.Connection {
		t.Fatalf("se esperaba connection=false para una API inalcanzable: %+v", response)
	}
	if response.Message != "ERROR: The API3 located on the VM2 is not working" {
		t.Fatalf("mensaje de error inesperado: %s", response.Message)
	}
}

func TestWrongCarnetIsNotARoute(t *testing.T) {
	app := mustApp(t, Config{
		Name:    "API1",
		VM:      "VM1",
		Carnet:  testCarnet,
		Address: ":8081",
		Targets: []Target{{Name: "API2", VM: "VM1", BaseURL: "http://api2:8082"}},
	})

	request := httptest.NewRequest(http.MethodGet, "/api1/000000000/call-api2", nil)
	recorder := httptest.NewRecorder()
	app.Handler().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("se esperaba 404 para otro carnet; se obtuvo %d", recorder.Code)
	}
}

func TestOnlyGETIsAccepted(t *testing.T) {
	app := mustApp(t, Config{Name: "API1", VM: "VM1", Carnet: testCarnet, Address: ":8081"})
	request := httptest.NewRequest(http.MethodPost, "/health", nil)
	recorder := httptest.NewRecorder()
	app.Handler().ServeHTTP(recorder, request)
	if recorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("se esperaba 405; se obtuvo %d", recorder.Code)
	}
}

func mustApp(t *testing.T, config Config) *App {
	t.Helper()
	app, err := New(config)
	if err != nil {
		t.Fatalf("New devolvió error: %v", err)
	}
	return app
}

func decode(t *testing.T, recorder *httptest.ResponseRecorder, destination any) {
	t.Helper()
	if err := json.NewDecoder(recorder.Body).Decode(destination); err != nil {
		t.Fatalf("JSON inválido: %v", err)
	}
}
