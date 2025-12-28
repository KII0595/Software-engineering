package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"
	"time"
)

type Metrics struct {
	Requests uint64 `json:"requests"`
	Uptime   string `json:"uptime"`
}

type WebServer struct {
	srv       *http.Server
	started   time.Time
	reqCount  uint64
}

func NewWebServer(port string) *WebServer {
	if port == "" {
		port = ":8000"
	}

	ws := &WebServer{
		started: time.Now(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /", ws.home)
	mux.HandleFunc("GET /status", ws.status)
	mux.HandleFunc("GET /metrics", ws.metrics)
	mux.HandleFunc("GET /sleep", ws.sleep)
	mux.HandleFunc("GET /compute", ws.compute)

	handler := ws.accessLog(ws.panicRecovery(mux))

	ws.srv = &http.Server{
		Addr:    port,
		Handler: handler,
	}

	return ws
}

func (ws *WebServer) accessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddUint64(&ws.reqCount, 1)
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func (ws *WebServer) panicRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rcv := recover(); rcv != nil {
				log.Printf("Обработана паника: %v", rcv)
				http.Error(w, "Внутренняя ошибка", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (ws *WebServer) home(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprintln(w, "Сервер конкурентности Go работает")
	fmt.Fprintln(w, "Эндпоинты: /status, /metrics, /sleep?sec=3, /compute")
}

func (ws *WebServer) status(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"state": "running"})
}

func (ws *WebServer) metrics(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	m := Metrics{
		Requests: atomic.LoadUint64(&ws.reqCount),
		Uptime:   time.Since(ws.started).Round(time.Second).String(),
	}
	json.NewEncoder(w).Encode(m)
}

func (ws *WebServer) sleep(w http.ResponseWriter, r *http.Request) {
	sec := 1
	if s := r.URL.Query().Get("sec"); s != "" {
		if n, err := time.ParseDuration(s + "s"); err == nil {
			sec = int(n.Seconds())
		}
	}
	time.Sleep(time.Duration(sec) * time.Second)
	fmt.Fprintf(w, "Задержка %d сек завершена\n", sec)
}

func (ws *WebServer) compute(w http.ResponseWriter, _ *http.Request) {
	sum := 0
	for i := 0; i < 2_000_000; i++ {
		sum += i
	}
	fmt.Fprintf(w, "Вычисление завершено: %d\n", sum)
}

func (ws *WebServer) Launch() error {
	log.Printf("Сервер запущен на http://localhost%s", ws.srv.Addr)
	return ws.srv.ListenAndServe()
}

func (ws *WebServer) Shutdown(ctx context.Context) error {
	log.Println("Остановка сервера...")
	return ws.srv.Shutdown(ctx)
}
