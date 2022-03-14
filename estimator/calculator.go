package estimator

import (
	"context"
	"fmt"
	"github.com/AlhimicMan/eip1559_gas_estimator/config"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ybbus/jsonrpc/v2"
	"sort"
	"sync"
	"time"
)

func WeiToGwe(wei int64) float64 {
	gWei := float64(wei) / 1e9
	return gWei
}

type ETHFeeEstimator struct {
	endpoint             string
	client               *ethclient.Client
	lastBlockEstimates   *GasPriceResult
	estimatesMu          sync.RWMutex
	lastProcessedBlock   uint64
	sleepDuration        time.Duration
	blockCountForAverage int
	rewardsPercentiles   []int
	requiredLevels       map[string]int
}

func NewETHFeeEstimator() (*ETHFeeEstimator, error) {
	endpoint := config.GetNodeEndpoint()
	client, err := initiateClient(endpoint)
	if err != nil {
		return nil, fmt.Errorf("client not initialized: %v", err)
	}
	requiredLevels, err := config.GetPercentilesToLevel()
	if err != nil {
		return nil, fmt.Errorf("error getting percents to level from config: %w", err)
	}
	var expectedRewardsPercentiles []int
	for _, level := range requiredLevels {
		expectedRewardsPercentiles = append(expectedRewardsPercentiles, level)
	}
	sort.Ints(expectedRewardsPercentiles)

	blocksCountHistory := config.GetBlocksCountForEstimate()
	timeToSleep := config.GetBlockCheckerSleepTime()
	sleepDuration := time.Duration(timeToSleep)

	return &ETHFeeEstimator{
		endpoint:             endpoint,
		client:               client,
		lastBlockEstimates:   nil,
		estimatesMu:          sync.RWMutex{},
		lastProcessedBlock:   0,
		sleepDuration:        sleepDuration,
		blockCountForAverage: blocksCountHistory,
		rewardsPercentiles:   expectedRewardsPercentiles,
		requiredLevels:       requiredLevels,
	}, nil

}

func initiateClient(endpoint string) (*ethclient.Client, error) {
	client, err := ethclient.Dial(endpoint)
	if err != nil {
		return nil, err
	}
	fmt.Println("Successfully made connection with endpoint")
	return client, nil
}

func (e *ETHFeeEstimator) getBaseFeeValue(ctx context.Context) (int64, error) {
	lastBlockRec, err := e.client.BlockByNumber(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("error getting last block data: %w", err)
	}
	lastBlockBaseFeeWei := lastBlockRec.BaseFee().Int64()
	return lastBlockBaseFeeWei, nil
}

func (e *ETHFeeEstimator) calculateBasicFees(ctx context.Context) (*LastBlockStat, error) {
	nodeLastBlock, err := e.client.BlockNumber(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting BlockNumber: %v", err)
	}

	baseFeeValue, err := e.getBaseFeeValue(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting base fee value: %v", err)
	}

	result := &LastBlockStat{
		BlockNumber: nodeLastBlock,
		BaseFee:     baseFeeValue,
	}
	return result, nil
}

func (e *ETHFeeEstimator) calculateFeeAverageHistory(historyItems []HistoryBlockFees) (*HistoricalFeesAverages, error) {
	perPercentilesLevels := make(map[int][]int64)
	for _, percentile := range e.rewardsPercentiles {
		perPercentilesLevels[percentile] = make([]int64, 0, len(historyItems))
	}
	for _, blockRec := range historyItems {
		for percent, feeLevel := range blockRec.PriorityFeesPerGas {
			perPercentilesLevels[percent] = append(perPercentilesLevels[percent], feeLevel)
		}
	}
	feesAverages := make(map[int]int64)
	for percentile, feesValues := range perPercentilesLevels {
		var total int64
		for _, number := range feesValues {
			total = total + number
		}
		average := total / int64(len(feesValues))
		feesAverages[percentile] = average
	}

	result := &HistoricalFeesAverages{AverageFees: feesAverages}
	return result, nil
}

func (e *ETHFeeEstimator) convertRewardsPercentiles(rewards []string) (map[int]int64, error) {
	if len(rewards) != len(e.rewardsPercentiles) {
		return nil, fmt.Errorf("lengths of rewards and percentiles are different")
	}
	results := make(map[int]int64)
	for i := 0; i < len(e.rewardsPercentiles); i++ {
		rewardVal, err := hexToInt(rewards[i])
		if err != nil {
			return nil, fmt.Errorf("error converting reward to int64: %w", err)
		}
		percentile := e.rewardsPercentiles[i]
		results[percentile] = rewardVal
	}
	return results, nil
}

func (e *ETHFeeEstimator) getBlocksHistoryFromNode(blocksCount int) (*HistoryRewards, error) {
	callParams := []interface{}{blocksCount, "latest", e.rewardsPercentiles}
	rewards := &HistoryRewards{}
	rpcClient := jsonrpc.NewClient(e.endpoint)
	err := rpcClient.CallFor(rewards, "eth_feeHistory", callParams)
	if err != nil {
		return nil, fmt.Errorf("error making call: %w", err)
	}
	return rewards, nil
}

func (e *ETHFeeEstimator) getHistoryBlocks(rewards *HistoryRewards) ([]HistoryBlockFees, error) {
	blocksCount := len(rewards.Reward)
	blocksRewards := make([]HistoryBlockFees, 0, blocksCount)
	var i int64
	for i = 0; i < int64(blocksCount); i++ {
		blockBaseFeeGas := rewards.BaseFeePerGas[i]
		blockBaseFeeGasValue, err := hexToInt(blockBaseFeeGas)
		if err != nil {
			fmt.Printf("error converting %s to int64\n", blockBaseFeeGas)
		}
		priorityFeesPerGas, err := e.convertRewardsPercentiles(rewards.Reward[i])
		if err != nil {
			fmt.Printf("error converting priorite fees per gas for block: %e\n", err)
		}
		resultRec := HistoryBlockFees{
			BaseFee:            blockBaseFeeGasValue,
			PriorityFeesPerGas: priorityFeesPerGas,
			UsedRatio:          rewards.GasUsedRatio[i],
		}
		blocksRewards = append(blocksRewards, resultRec)
	}
	return blocksRewards, nil
}

func (e *ETHFeeEstimator) printHistoricalFeesAverage(averagesStat *HistoricalFeesAverages) {
	for _, percent := range e.rewardsPercentiles {
		average := averagesStat.AverageFees[percent]
		fmt.Printf("for %d average value is %f\n", percent, WeiToGwe(average))
	}
}

func (e *ETHFeeEstimator) printBasicFeesLog(blockStat *LastBlockStat) {
	fmt.Printf("For block %d calculated:\n\tbase fee: %f\n",
		blockStat.BlockNumber, WeiToGwe(blockStat.BaseFee))
	fmt.Println("----")
}

func (e *ETHFeeEstimator) RunEstimatesProcessor(ctx context.Context) {

	for true {
		nodeLastBlock, err := e.client.BlockNumber(ctx)
		if err != nil {
			fmt.Printf("Error getting BlockNumber: %v", err)
			continue
		}
		if nodeLastBlock == e.lastProcessedBlock {
			time.Sleep(e.sleepDuration * time.Second)
			continue
		}
		e.lastProcessedBlock = nodeLastBlock

		blockStat, err := e.calculateBasicFees(ctx)
		if err != nil {
			fmt.Printf("Error calculating basic fees from last block: %v", err)
			continue
		}

		rewards, err := e.getBlocksHistoryFromNode(e.blockCountForAverage)
		if err != nil {
			fmt.Printf("error getting blocks history: %v", err)
			continue
		}

		historyItems, err := e.getHistoryBlocks(rewards)
		if err != nil {
			fmt.Printf("error getting history items: %v", err)
			continue
		}

		averageFees, err := e.calculateFeeAverageHistory(historyItems)
		if err != nil {
			fmt.Printf("error calculating average fees for percentiles: %v", err)
			continue
		}
		if config.GetDebugEnabled() {
			fmt.Printf("\nLast block is %d\n", nodeLastBlock)
			e.printBasicFeesLog(blockStat)
			e.printHistoricalFeesAverage(averageFees)
		}
		priceLevels := make([]GasPriceForSpeed, 0, len(e.requiredLevels))
		for levelName, levelPercentile := range e.requiredLevels {
			feeForLevel := averageFees.AverageFees[levelPercentile]
			maxFeePerGas := feeForLevel + 2*blockStat.BaseFee
			speedResult := GasPriceForSpeed{
				SpeedName:            levelName,
				MaxFeePerGas:         maxFeePerGas,
				MaxPriorityFeePerGas: feeForLevel,
			}
			priceLevels = append(priceLevels, speedResult)
		}
		lastCalculatedStat := &GasPriceResult{
			LastBlock:     nodeLastBlock,
			BaseFeePerGas: blockStat.BaseFee,
			PriceLevels:   priceLevels,
		}
		e.estimatesMu.Lock()
		e.lastBlockEstimates = lastCalculatedStat
		e.estimatesMu.Unlock()
		time.Sleep(e.sleepDuration * time.Second)
	}
}

func (e *ETHFeeEstimator) GetLastEstimatedPrice() GasPriceResult {
	e.estimatesMu.RLock()
	result := *e.lastBlockEstimates
	e.estimatesMu.RUnlock()
	return result
}
