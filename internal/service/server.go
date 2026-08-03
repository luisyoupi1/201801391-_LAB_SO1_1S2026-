package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
)

const maxHealthResponseBytes = 1 << 20

// Target identifies another API that this service can check through /health.
type Target struct {
	Name    string
	VM      string
	BaseURL string
}

// Config contains everything that changes between API1, API2 and API3.
type Config struct {
	Name    string
	VM      string
	Carnet  string
	Address string
	Targets []Target
	Client  *http.Client
	Now     func() time.Time
}

type App struct {
	config Config
}

type healthResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
	VM        string `json:"VM"`
	Carnet    string `json:"carnet"`
}

type callResponse struct {
	APIName    string `json:"apiname"`
	Message    string `json:"message"`
	Connection bool   `json:"connection"`
	Carnet     string `json:"carnet"`
}

func New(config Config) (*App, error) {
	if strings.TrimSpace(config.Name) == "" {
		return nil, errors.New("el nombre de la API es obligatorio")
	}
	if strings.TrimSpace(config.VM) == "" {
		return nil, errors.New("el nombre de la VM es obligatorio")
	}
	if strings.TrimSpace(config.Carnet) == "" {
		return nil, errors.New("el carnet es obligatorio")
	}
	if config.Address == "" {
		return nil, errors.New("la dirección de escucha es obligatoria")
	}
	if config.Client == nil {
		config.Client = &http.Client{Timeout: 3 * time.Second}
	}
	if config.Now == nil {
		config.Now = time.Now
	}
	for _, target := range config.Targets {
		if target.Name == "" || target.VM == "" || target.BaseURL == "" {
			return nil, errors.New("cada API destino necesita nombre, VM y URL")
		}
	}
	return &App{config: config}, nil
}

func (app *App) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/health", getOnly(app.health))

	apiPath := strings.ToLower(app.config.Name)
	for _, configuredTarget := range app.config.Targets {
		target := configuredTarget
		path := fmt.Sprintf("/%s/%s/call-%s", apiPath, app.config.Carnet, strings.ToLower(target.Name))
		mux.HandleFunc(path, getOnly(app.call(target)))
	}

	return securityHeaders(mux)
}

func (app *App) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, healthResponse{
		Status:    "UP",
		Message:   fmt.Sprintf("%s is Ready", app.config.Name),
		Timestamp: app.config.Now().UTC().Format(time.RFC3339),
		VM:        app.config.VM,
		Carnet:    app.config.Carnet,
	})
}

func (app *App) call(target Target) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connected := app.targetIsUp(r, target)
		message := fmt.Sprintf("The %s located on the %s is working", target.Name, target.VM)
		if !connected {
			message = fmt.Sprintf("ERROR: The %s located on the %s is not working", target.Name, target.VM)
		}
		writeJSON(w, http.StatusOK, callResponse{
			APIName:    target.Name,
			Message:    message,
			Connection: connected,
			Carnet:     app.config.Carnet,
		})
	}
}

func (app *App) targetIsUp(r *http.Request, target Target) bool {
	url := strings.TrimRight(target.BaseURL, "/") + "/health"
	req, err := http.NewRequestWithContext(r.Context(), http.MethodGet, url, nil)
	if err != nil {
		return false
	}

	response, err := app.config.Client.Do(req)
	if err != nil {
		return false
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return false
	}

	var health healthResponse
	decoder := json.NewDecoder(io.LimitReader(response.Body, maxHealthResponseBytes))
	if err := decoder.Decode(&health); err != nil {
		return false
	}
	return health.Status == "UP"
}

func getOnly(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			writeJSON(w, http.StatusMethodNotAllowed, map[string]string{"error": "method not allowed"})
			return
		}
		next(w, r)
	}
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("no se pudo escribir la respuesta JSON: %v", err)
	}
}

func Env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

// Serve starts the API with conservative timeouts and a graceful shutdown.
func Serve(config Config) error {
	app, err := New(config)
	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:              config.Address,
		Handler:           app.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	stop, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	go func() {
		<-stop.Done()
		_ = server.Close()
	}()

	log.Printf("%s lista en %s (VM=%s, carnet=%s)", config.Name, config.Address, config.VM, config.Carnet)
	err = server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}
