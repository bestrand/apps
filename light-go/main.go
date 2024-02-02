package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "os"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
    infoAPI = "https://ipwhois.app/json/"
    port    = ":8080"
)

var (
    light_switch = "off"
    duration  prometheus.Gauge
    metrics   = prometheus.NewRegistry()
)

type Info struct {
    Continent, Country, City, Switch string
    Latitude, Longitude    float64
}

func main() {
    initMetrics()
    light_switch = os.Getenv("SWITCH")
    go light()
    http.HandleFunc("/", infoHandler)
    http.Handle("/metrics", promhttp.HandlerFor(metrics, promhttp.HandlerOpts{}))
    fmt.Println("Server is running on", port)
    err := http.ListenAndServe(port, nil)
    handleError(err, "Failed to start the server: %v\n")
}

func initMetrics() {
    duration = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "duration",
        Help: "Duration oflight being on",
    })
    metrics.MustRegister(duration)
}

func get(url string, target interface{}) error {
    resp, err := http.Get(url)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("API request failed with status code: %d", resp.StatusCode)
    }
    return json.NewDecoder(resp.Body).Decode(target)
}

func infoHandler(w http.ResponseWriter, r *http.Request) {

    server, err := serverInfo()
    if err != nil {
        handleError(err, "Internal Server Error: %v\n")
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    server.Switch = light_switch

    response(w, server)
}

func serverInfo() (*Info, error) {
    var info Info
    err := get(infoAPI, &info)
    return &info, err
}

func light() {
    startTime := time.Now()

    go func() {
        for {
            now := time.Now()
            elapsed := float64(now.Sub(startTime).Nanoseconds())
            duration.Set(elapsed)
            time.Sleep(time.Second)
        }
    }()
}

func response(w http.ResponseWriter, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    if err := json.NewEncoder(w).Encode(data); err != nil {
        handleError(err, "Internal Server Error: %v\n")
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    }
}

func handleError(err error, format string) {
    if err != nil {
        fmt.Printf(format, err)
    }
}
