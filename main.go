package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/big"
	"time"
)

// 优化点 1：定义工作池大小，限制并发数，保护 API 不被封
// const workerCount = 5
type Block struct {
	Timestamp     int64    // 时间戳
	Data          []byte   // 交易数据（第一步先用简单的字符串代替）
	PrevBlockHash []byte   // 前一个区块的哈希值
	Hash          []byte   // 当前区块的哈希值
	Nonce         int      // 随机数
	Difficulty    int      // 难度
	Target        *big.Int // 目标值
}
type BlockChain struct {
	Blocks []*Block
}

const targetBits = 17 // 难度值：代表哈希值前导零的位数（这里数值越大越难）
type ProofOfWork struct {
	block  *Block
	target *big.Int // 目标值：算出来的哈希必须比这个数小
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	fmt.Println(target)
	target.Lsh(target, uint(256-b.Difficulty))
	fmt.Println(target)
	pow := &ProofOfWork{b, target}
	return pow
}
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join([][]byte{
		pow.block.PrevBlockHash,
		pow.block.Data,
		big.NewInt(pow.block.Timestamp).Bytes(),
		big.NewInt(int64(nonce)).Bytes(),
		big.NewInt(int64(targetBits)).Bytes(),
	}, []byte{})
	return data
}
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0
	fmt.Printf("开始挖掘包含数据 \"%s\" 的区块\n", pow.block.Data)
	for nonce < 100000000 { // 设置一个足够大的上限
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data) // 计算哈希
		fmt.Printf("\r%x", hash)   // 实时打印哈希值（可选）
		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) == -1 {
			break
		}
		nonce++
	}
	fmt.Print("\n\n")
	return nonce, hash[:]
}
func NewBlockChain() *BlockChain {
	return &BlockChain{
		Blocks: []*Block{
			&Block{Data: []byte("Genesis Block")},
		},
	}
}
func (bc *BlockChain) AddBlock(data string) {
	prevBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock := &Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlock.Hash,
		Nonce:         0,
		Difficulty:    targetBits,
	}
	// SetHash(newBlock)
	pow := NewProofOfWork(newBlock)
	nonce, hash := pow.Run()
	newBlock.Hash = hash[:]
	newBlock.Nonce = nonce
	bc.Blocks = append(bc.Blocks, newBlock)
}

// 编写一个 SetHash 方法。将 PrevBlockHash + Data + Timestamp 拼接后进行 SHA-256 运算。
func SetHash(block *Block) {
	hash := sha256.New()
	hash.Write(block.PrevBlockHash)
	hash.Write(block.Data)
	hash.Write(big.NewInt(block.Timestamp).Bytes())
	block.Hash = hash.Sum(nil)
}
func main() {
	block := &Block{
		Timestamp:     1234567890,
		Data:          []byte("test"),
		PrevBlockHash: []byte("prevHash"),
	}
	SetHash(block)
	fmt.Println(hex.EncodeToString(block.Hash))
	bc := NewBlockChain()
	bc.AddBlock("Send 1 BTC to Ivan")
	bc.AddBlock("Send 2 more BTC to Ivan")
	for _, block := range bc.Blocks {
		fmt.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Nonce: %d\n", block.Nonce)
		fmt.Printf("Difficulty: %d\n", block.Difficulty)
		fmt.Printf("Target: %x\n", block.Target)
		fmt.Println()
	}
}
