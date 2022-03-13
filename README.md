# EIP-1559 gas price estimator

Calculate gas price for transactions with different speed estimates. 
Integrated with Alchemy API for historical transactions.

Calculated speed from average value of last blocks (config parameter analyze_blocks). Requesting blocks history, get
priority fee levels of transactions in this blocks in percentiles from levels parameters. Then calculating average fee 
value in percentile. This average values - estimated gas fee.

## Config

Rename file config.yaml_template to config.yaml. Config fields:
- node - address of Alchemy node. Fill this parameter with token
- levels - key-value config with speed name and expected percentile of transactions
- analyze_blocks - how many blocks analyze for average gas price 
- sleep_seconds - how long sleep before checking if new block appeared
- server.host - address to listen for server API
- debug - set log level to debug. If true will be printed results for each block
