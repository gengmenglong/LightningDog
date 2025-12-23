package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/rpc"
)

// 优化点 1：定义工作池大小，限制并发数，保护 API 不被封
const workerCount = 5

func main() {
	// wssUrl := "wss://eth-mainnet.g.alchemy.com/v2/你的APIKey"
	wssUrl := "wss://eth-mainnet.g.alchemy.com/v2/SCz3YIdYkR5bXVwawzgbo"
	rpcClient, _ := rpc.Dial(wssUrl)
	client := ethclient.NewClient(rpcClient)

	txHashes := make(chan common.Hash, 1000) // 带缓冲的通道
	var wg sync.WaitGroup

	// 优化点 2：启动固定数量的协程 (Workers)
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for hash := range txHashes {
				// 在这里执行查询动作
				tx, isPending, err := client.TransactionByHash(context.Background(), hash)
				if err != nil || !isPending {
					continue
				}
				// 逻辑过滤：比如只关心转账金额 > 1 ETH 的交易
				oneEth := new(big.Int)
				oneEth.SetString("1000000000000000000", 10) // 1 ETH = 10^18 wei
				// if tx.Value().Cmp(oneEth) > 0 {
				// 	fmt.Printf("[Worker %d] 捕获大额交易: %s\n", workerID, hash.Hex())
				// }
				// 改进后的打印逻辑
				if tx.Value().Cmp(oneEth) > 0 {
					// 1. 获取发送者地址 (需要计算，因为 tx 里存的是签名)
					// 这里的 chainID 建议在程序初始化时获取，主网通常是 1
					signer := types.LatestSignerForChainID(big.NewInt(1))
					from, _ := types.Sender(signer, tx)

					// 2. 转换金额单位 (从 Wei 转为 ETH)
					fAmount := new(big.Float).SetInt(tx.Value())
					ethValue := new(big.Float).Quo(fAmount, big.NewFloat(1e18))

					fmt.Printf("\n--- [Worker %d] 发现大鱼！ ---\n", workerID)
					fmt.Printf("交易哈希: %s\n", tx.Hash().Hex())
					fmt.Printf("发送方: %s\n", from.Hex())
					if tx.To() != nil {
						fmt.Printf("接收方: %s\n", tx.To().Hex())
					}
					fmt.Printf("金额: %.4f ETH\n", ethValue)
					fmt.Printf("Gas 价格: %v Gwei\n", tx.GasPrice().Uint64()/1e9)
					fmt.Println("---------------------------")
				}
			}
		}(i)
	}

	// 3. 订阅哈希流
	subHashes := make(chan common.Hash)
	sub, _ := rpcClient.EthSubscribe(context.Background(), subHashes, "newPendingTransactions")

	fmt.Println("🚀 优化后的观察者已启动...")

	for {
		select {
		case err := <-sub.Err():
			log.Fatal(err)
		case hash := <-subHashes:
			// 优化点 3：非阻塞地将哈希扔进任务队列
			select {
			case txHashes <- hash:
			default:
				// 如果队列满了，丢弃该哈希，防止程序卡死
			}
		}
	}
}
