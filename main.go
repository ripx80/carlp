package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

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

func recordMetrics() {
	go func() {
		for {
			opsProcessed.Inc()
			time.Sleep(2 * time.Second)
		}
	}()
}

func parse(f string) ([]byte, error) {
	fp, err := os.Open(f)
	if err != nil {
		return nil, err

	}
	defer fp.Close()

	parser := parser.NewParser(fp)
	data, ln, err := parser.Parse()
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
	// return b, nil
}

// recordMetrics()
// fmt.Println("running")
// http.Handle("/metrics", promhttp.Handler())
// http.ListenAndServe(":2112", nil)
//fp, err := os.Open(filepath.Join("save", "ln.txt"))
//fp, err := os.Open(filepath.Join("test", "short.txt"))
//fp, err := os.Open(filepath.Join("save", "intel.txt"))

func respJson(w http.ResponseWriter, r *http.Request) {
	data, err := parse(filepath.Join("save", "short"))
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

	//b, err := parse(filepath.Join("save", "short"))
	//b, err := parse(filepath.Join("save", "gamestate"))
	fmt.Println("running")
	handleRequest()

	// err = os.WriteFile("gamestate.json", b, 0644)
	// if err != nil {
	// 	fmt.Println("error:", err)
	// 	return
	// }

}
