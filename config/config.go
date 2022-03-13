package config

import (
	"fmt"
	"github.com/spf13/viper"
	"strconv"
)

func InitConfig() error {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
	}
	return nil
}

func GetListenHost() string {
	listenHost := viper.GetString("server.host")
	return listenHost
}

func GetNodeEndpoint() string {
	endpoint := viper.GetString("node")
	return endpoint
}

func GetBlocksCountForEstimate() int {
	blocksCount := viper.GetInt("analyze_blocks")
	return blocksCount
}

func GetBlockCheckerSleepTime() int {
	sleepSeconds := viper.GetInt("sleep_seconds")
	return sleepSeconds
}

func GetPercentilesToLevel() (map[string]int, error) {
	configLevels := viper.GetStringMapString("levels")
	resultLevels := make(map[string]int)
	for level, percentile := range configLevels {
		percentVal, err := strconv.ParseInt(percentile, 10, 64)
		if err != nil {
			return resultLevels, fmt.Errorf("error converting config levels percents: %w", err)
		}
		resultLevels[level] = int(percentVal)
	}
	return resultLevels, nil
}
