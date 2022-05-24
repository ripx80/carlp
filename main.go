package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/ripx80/carlp/pkgs/parser"
)

var (
	opsProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "myapp_processed_ops_total",
		Help: "The total number of processed events",
	})
)

type GameState struct {
	Version                  string                 `json:"version"`
	Name                     string                 `json:"name"`
	Version_control_revision int                    `json:"version_control_revision=86097"`
	Date                     string                 `json:"date"`
	Required_dlcs            []string               `json:"required_dlcs"`
	Player                   []Player               `json:"player"`
	Market                   Market                 `json market`
	X                        map[string]interface{} `json:"-"` // Rest of the fields should go here.
}

type Player struct {
	Country int    `json:country`
	Name    string `json:name`
}

type Market struct {
	Monthly_trades []MarketMonTrades
}

type MarketMonTrades struct {
	Trade_data TradeData
	Amount     int
	Price      int
	Id         int
}

type TradeData struct {
	Trade_type string
	Resource   string
	Country    int
}

func recordMetrics() {
	go func() {
		for {
			opsProcessed.Inc()
			time.Sleep(2 * time.Second)
		}
	}()
}

func parse(f string) (map[string]interface{}, error) {
	fp, err := os.Open(f)
	if err != nil {
		return nil, err

	}
	defer fp.Close()

	parser := parser.NewParser(fp)
	data, _, err := parser.Parse()
	if err != nil {
		return nil, err
	}
	return data, nil
	// fmt.Printf("scan lines: %d\n", ln)
	// fmt.Println("Undefined keys found:", parser.UndefKey)
	// //fmt.Println(data)

	// b, err := json.MarshalIndent(data, "", "  ")
	// if err != nil {
	// 	return nil, err
	// }

	// err = os.WriteFile("gamestate.json", b, 0644)
	// if err != nil {
	// 	fmt.Println("error:", err)
	// 	return
	// }

}

//fp, err := os.Open(filepath.Join("save", "ln.txt"))
//fp, err := os.Open(filepath.Join("test", "short.txt"))
//fp, err := os.Open(filepath.Join("save", "intel.txt"))

func respJson(w http.ResponseWriter, r *http.Request) {
	data, err := parse(filepath.Join("test", "gamestate"))
	if err != nil {
		log.Println(err)
		return
	}
	json.NewEncoder(w).Encode(data)
}

func handleRequest() {
	http.HandleFunc("/json", respJson)
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":2112", nil))
}

func main() {
	// fmt.Println("running")
	handleRequest()

	data, err := parse(filepath.Join("test", "short.txt"))
	if err != nil {
		log.Println(err)
		return
	}

	game := &GameState{}
	err = mapstructure.Decode(data, &game)
	if err != nil {
		// error
	}
	fmt.Print(game)
}
