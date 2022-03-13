package estimator

// LastBlockStat - fees from last block
type LastBlockStat struct {
	BlockNumber uint64
	BaseFee     int64
}

type HistoryRewards struct {
	OldestBlock   string     `json:"oldestBlock"`
	Reward        [][]string `json:"reward"`
	BaseFeePerGas []string   `json:"baseFeePerGas"`
	GasUsedRatio  []float64  `json:"gasUsedRatio"`
}

// HistoryBlockFees - history fees from block. all fees are in wei
type HistoryBlockFees struct {
	BaseFee            int64         `json:"base_fee"`
	PriorityFeesPerGas map[int]int64 `json:"priority_fees_per_gas"` // Divided in percentiles
	UsedRatio          float64       `json:"used_ratio"`
}

// HistoricalFeesAverages - average fees in selected percentiles
type HistoricalFeesAverages struct {
	AverageFees map[int]int64
}

// GasPriceForSpeed - record for one speed gas price
type GasPriceForSpeed struct {
	SpeedName            string `json:"speed"`
	MaxFeePerGas         int64  `json:"max_fee_per_gas"`
	MaxPriorityFeePerGas int64  `json:"max_priority_fee_per_gas"`
}

// GasPriceResult - result for calculated gas price for levels
type GasPriceResult struct {
	LastBlock     uint64             `json:"last_block"`
	BaseFeePerGas int64              `json:"base_fee_per_gas"`
	PriceLevels   []GasPriceForSpeed `json:"price_levels"`
}
