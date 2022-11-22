package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	endpoint := strings.ToLower(os.Getenv("ETHERSCAN_NETWORK"))
	switch endpoint {
	case "mainnet":
		endpoint = "https://api.etherscan.io"
	case "goerli":
		endpoint = "https://api-goerli.etherscan.io"
	}
	endpoint = fmt.Sprintf("%s/api?module=proxy&action=eth_blockNumber&apikey=%s", endpoint, os.Getenv("ETHERSCAN_KEY"))

	mux := http.NewServeMux()

	reg := prometheus.NewRegistry()
	reg.MustRegister(
		prometheus.NewGaugeFunc(prometheus.GaugeOpts{
			Name: "eth_block_number",
			Help: "Latest processed block",
		}, func() float64 {
			return getLatestBlock(endpoint)
		}),
	)
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))

	srv := http.Server{
		Addr:    fmt.Sprintf(":%d", 8080),
		Handler: mux,
	}

	log.Fatal(srv.ListenAndServe())
}

type response struct {
	Status string `json:"status"`
	Result string `json:"result"`
}

func getLatestBlock(endpoint string) float64 {
	resp, err := http.Get(endpoint)
	if err != nil {
		log.Printf("net error: %s\n", err)
		return 0
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("read error: %s\n", err)
		return 0
	}

	var result response
	if err := json.Unmarshal(body, &result); err != nil {
		log.Printf("json error: %s\n", err)
		return 0
	}

	if result.Status == "0" {
		log.Printf("Etherscan API error: %s\n", result.Result)
		return 0
	}

	block, err := strconv.ParseInt(result.Result, 0, 64)
	if result.Status == "0" {
		log.Printf("Failed to parse block '%s': %s\n", result.Result, err)
		return 0
	}

	return float64(block)
}
