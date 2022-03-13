package main

import (
	"context"
	"fmt"
	"github.com/AlhimicMan/eip1559_gas_estimator/config"
	"github.com/AlhimicMan/eip1559_gas_estimator/estimator"
	"net/http"
)

func main() {
	err := config.InitConfig()
	if err != nil {
		fmt.Printf("Error initialising config: %v", err)
		return
	}
	ethEstimator, err := estimator.NewETHFeeEstimator()
	if err != nil {
		fmt.Printf("Error initialising eth gas price estimator: %v", err)
		return
	}
	handler := GetServerHandlers(ethEstimator)
	ctx := context.Background()
	go ethEstimator.RunEstimatesProcessor(ctx)

	listenAddr := config.GetListenHost()
	fmt.Println(fmt.Sprintf("Staring application on %s", listenAddr))

	_ = http.ListenAndServe(listenAddr, handler)
}
