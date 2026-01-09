package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/davidl71/exarp-go/internal/tools"
)

type StatisticsRequest struct {
	Function string    `json:"function"`
	Data     []float64 `json:"data"`
	P        *float64  `json:"p,omitempty"`        // For quantile
	Decimals *int      `json:"decimals,omitempty"` // For round
	Value    *float64  `json:"value,omitempty"`    // For round
}

type StatisticsResponse struct {
	Result  *float64 `json:"result,omitempty"`
	Error   *string  `json:"error,omitempty"`
	Success bool     `json:"success"`
}

func main() {
	if len(os.Args) < 2 {
		respondError("JSON request required as argument")
		os.Exit(1)
	}

	var req StatisticsRequest
	if err := json.Unmarshal([]byte(os.Args[1]), &req); err != nil {
		respondError(fmt.Sprintf("Failed to parse JSON: %v", err))
		os.Exit(1)
	}

	var result float64

	switch req.Function {
	case "mean":
		result = tools.Mean(req.Data)
	case "median":
		result = tools.Median(req.Data)
	case "stdev":
		result = tools.StdDev(req.Data)
	case "quantile":
		if req.P == nil {
			respondError("quantile requires 'p' parameter")
			os.Exit(1)
		}
		result = tools.Quantile(req.Data, *req.P)
	case "round":
		if req.Value == nil {
			respondError("round requires 'value' parameter")
			os.Exit(1)
		}
		decimals := 2
		if req.Decimals != nil {
			decimals = *req.Decimals
		}
		result = tools.Round(*req.Value, decimals)
	default:
		respondError(fmt.Sprintf("Unknown function: %s", req.Function))
		os.Exit(1)
	}

	respondSuccess(result)
}

func respondSuccess(result float64) {
	response := StatisticsResponse{
		Result:  &result,
		Success: true,
	}
	jsonBytes, _ := json.Marshal(response)
	fmt.Println(string(jsonBytes))
}

func respondError(message string) {
	response := StatisticsResponse{
		Error:   &message,
		Success: false,
	}
	jsonBytes, _ := json.Marshal(response)
	fmt.Println(string(jsonBytes))
}
