package estimator

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"reflect"
	"sync"
	"testing"
)

func TestFeeEstimator_calculateFeeAverageHistory(t *testing.T) {
	tests := []struct {
		name         string
		historyItems []HistoryBlockFees
		speedLevels  map[string]int
		want         *HistoricalFeesAverages
		wantErr      bool
	}{
		{
			name: "calculate simple history",
			historyItems: []HistoryBlockFees{
				{
					BaseFee: 101669930643,
					PriorityFeesPerGas: map[int]int64{
						5:  2425000000,
						10: 2500000000,
						55: 14103005255,
						80: 66195898652,
					},
					UsedRatio: 0.2,
				},
				{
					BaseFee: 160058947776,
					PriorityFeesPerGas: map[int]int64{
						5:  2500000000,
						10: 2500000000,
						55: 2500000000,
						80: 39941052224,
					},
					UsedRatio: 0.9,
				},
			},
			speedLevels: map[string]int{"safeLow": 5, "average": 10, "fast": 55, "fastest": 80},
			want: &HistoricalFeesAverages{
				AverageFees: map[int]int64{
					5:  2462500000,
					10: 2500000000,
					55: 8301502627,
					80: 53068475438,
				},
			},
			wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ETHFeeEstimator{
				endpoint:             "no-ep",
				client:               nil,
				lastBlockEstimates:   nil,
				estimatesMu:          sync.RWMutex{},
				lastProcessedBlock:   0,
				sleepDuration:        1,
				blockCountForAverage: len(tt.historyItems),
				rewardsPercentiles:   []int{5, 10, 55, 80},
				requiredLevels:       tt.speedLevels,
			}
			got, err := e.calculateFeeAverageHistory(tt.historyItems)
			assert.Nil(t, err)
			if (err != nil) != tt.wantErr {
				t.Errorf("calculateFeeAverageHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateFeeAverageHistory() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeeEstimator_convertRewardsPercentiles(t *testing.T) {

	tests := []struct {
		name    string
		rewards []string

		want    map[int]int64
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "test rewards calculation",
			rewards: []string{
				"0x59682f00",
				"0x59782f00",
				"0x597a2f00",
				"0x77359400",
			},
			want: map[int]int64{
				5:  1500000000,
				10: 1501048576,
				55: 1501179648,
				80: 2000000000,
			},
			wantErr: assert.NoError,
		},
		{
			name: "test length mismatch",
			rewards: []string{
				"0x59682f00",
				"0x59782f00",
				"0x597a2f00",
			},
			want:    nil,
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ETHFeeEstimator{
				endpoint:             "no-ep",
				client:               nil,
				lastBlockEstimates:   nil,
				estimatesMu:          sync.RWMutex{},
				lastProcessedBlock:   0,
				sleepDuration:        1,
				blockCountForAverage: 1,
				rewardsPercentiles:   []int{5, 10, 55, 80},
				requiredLevels:       nil,
			}
			got, err := e.convertRewardsPercentiles(tt.rewards)
			if !tt.wantErr(t, err, fmt.Sprintf("convertRewardsPercentiles(%v)", tt.rewards)) {
				return
			}
			assert.Equalf(t, tt.want, got, "convertRewardsPercentiles(%v)", tt.rewards)
		})
	}
}

func TestFeeEstimator_getHistoryBlocks(t *testing.T) {
	tests := []struct {
		name          string
		blocksRewards *HistoryRewards
		wantErr       assert.ErrorAssertionFunc
	}{
		{
			name: "test one block",
			blocksRewards: &HistoryRewards{
				OldestBlock: "0xdb7a7d",
				Reward: [][]string{
					{"0x3b9aca00", "0x448b9b80"},
					{"0x1a3608c8", "0x3b9aca00"},
				},
				BaseFeePerGas: []string{"0x4d2cda487", "0x540b21d38"},
				GasUsedRatio:  []float64{0.85598, 0.73598},
			},
			wantErr: assert.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ETHFeeEstimator{
				endpoint:             "no-ep",
				client:               nil,
				lastBlockEstimates:   nil,
				estimatesMu:          sync.RWMutex{},
				lastProcessedBlock:   0,
				sleepDuration:        1,
				blockCountForAverage: 1,
				rewardsPercentiles:   []int{5, 10},
				requiredLevels:       nil,
			}
			got, err := e.getHistoryBlocks(tt.blocksRewards)
			if !tt.wantErr(t, err, fmt.Sprintf("getHistoryBlocks(%v)", tt.blocksRewards)) {
				return
			}
			for i, reward := range got {
				assert.NotNilf(t, reward.PriorityFeesPerGas, fmt.Sprintf("in reward %d", i))
				expectedRatio := tt.blocksRewards.GasUsedRatio[i]
				assert.Equal(t, expectedRatio, reward.UsedRatio)
			}
			assert.Equal(t, len(got), len(tt.blocksRewards.Reward))
		})
	}
}
