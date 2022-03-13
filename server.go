package main

import (
	"encoding/json"
	"fmt"
	"github.com/AlhimicMan/eip1559_gas_estimator/estimator"
	"net/http"
)

type ServerHandler struct {
	ethEstimator *estimator.ETHFeeEstimator
}

func GetServerHandlers(estimator *estimator.ETHFeeEstimator) *http.ServeMux {
	handler := &ServerHandler{ethEstimator: estimator}
	r := http.NewServeMux()
	r.HandleFunc("/eth", handler.EthGasPriceHandler)
	return r
}

func (h *ServerHandler) EthGasPriceHandler(w http.ResponseWriter, r *http.Request) {
	ethGaspPrice := h.ethEstimator.GetLastEstimatedPrice()
	err := json.NewEncoder(w).Encode(ethGaspPrice)
	if err != nil {
		fmt.Printf("[EthGasPriceHandler] error encoding results: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
