package main

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
	"sort"
	"strconv"
	"time"
)

const (
	voteNodeNum      = 20
	superNodeNum     = 10
	mineSuperNodeNum = 3
	coinInTxMax      = 10
	coinMax          = 10000
)

type block struct {
	//上一个块的hash
	prehash string
	//本块hash
	hash string
	//时间戳
	timestamp string
	//区块内容
	data string
	//区块高度
	height int
	//挖出本块的节点地址
	address string
}

//用于存储区块链
var blockchain []block

// 交易存储队列
var txQueue string

//普通节点
type node struct {
	//代币数量
	votes int
	//节点地址
	address string
}

//竞选节点
type superNode struct {
	node
}

//投票节点池
var voteNodesPool []node

//竞选节点池
var starNodesPool []superNode

//存放可以挖矿的超级节点池
var superStarNodesPool []superNode

//生成新的区块
func generateNewBlock(oldBlock block, data string, address string) block {
	newBlock := block{}
	newBlock.prehash = oldBlock.hash
	newBlock.data = data
	newBlock.timestamp = time.Now().Format("2006-01-02 15:04:05")
	newBlock.height = oldBlock.height + 1
	newBlock.address = address
	newBlock.getHash()
	return newBlock
}

//对自身进行散列
func (b *block) getHash() {
	sumString := b.prehash + b.timestamp + b.data + b.address + strconv.Itoa(b.height)
	hash := sha256.Sum256([]byte(sumString))
	b.hash = hex.EncodeToString(hash[:])
}

//投票
func voting() {
	for _, v := range voteNodesPool {
		rInt, err := rand.Int(rand.Reader, big.NewInt(superNodeNum+1))
		if err != nil {
			log.Panic(err)
		}
		starNodesPool[int(rInt.Int64())].votes += v.votes
	}
}

//对挖矿节点进行排序
func sortMineNodes() {
	sort.Slice(starNodesPool, func(i, j int) bool {
		return starNodesPool[i].votes > starNodesPool[j].votes
	})
	superStarNodesPool = starNodesPool[:mineSuperNodeNum]
}

//初始化
func init() {
	//初始化投票节点
	for i := 0; i <= voteNodeNum; i++ {
		rInt, err := rand.Int(rand.Reader, big.NewInt(coinMax))
		if err != nil {
			log.Panic(err)
		}
		voteNodesPool = append(voteNodesPool, node{int(rInt.Int64()), "投票节点" + strconv.Itoa(i)})
	}
	//初始化竞选节点
	for i := 0; i <= superNodeNum; i++ {
		starNodesPool = append(starNodesPool, superNode{node{0, "超级节点" + strconv.Itoa(i)}})
	}
}

// 在节点间产生随机交易
func genTransactions(tps int) {
	period := 1000.0 / tps //ms
	go func() {
		p := period
		for {
			time.Sleep(time.Duration(p) * time.Millisecond)
			coins, _ := rand.Int(rand.Reader, big.NewInt(coinInTxMax))
			sen, _ := rand.Int(rand.Reader, big.NewInt(voteNodeNum))
			rec, _ := rand.Int(rand.Reader, big.NewInt(voteNodeNum))
			for { // 随机获取两个节点
				if rec != sen && voteNodesPool[sen.Int64()].votes > int(coins.Int64()) { // 账户余额检查
					break
				}
				rec, _ = rand.Int(rand.Reader, big.NewInt(voteNodeNum))
			}
			voteNodesPool[sen.Int64()].votes -= int(coins.Int64())
			voteNodesPool[rec.Int64()].votes += int(coins.Int64())
			txQueue += voteNodesPool[sen.Int64()].address + "->" + voteNodesPool[rec.Int64()].address + ":" + coins.String() + ";"
		}
	}()
}

func main() {
	fmt.Println("初始化", voteNodeNum, "个投票节点...")
	fmt.Println(voteNodesPool)
	fmt.Println("当前存在的", superNodeNum, "个竞选节点")
	fmt.Println(starNodesPool)
	fmt.Println("投票节点们开始进行投票...")
	voting()
	fmt.Println("结束投票，查看竞选节点们获得票数...")
	fmt.Println(starNodesPool)
	fmt.Println("对竞选节点按获得票数排序，前", mineSuperNodeNum, "名，当选超级节点")
	sortMineNodes()
	fmt.Println(superStarNodesPool)
	fmt.Println("开始挖矿...")
	genesisBlock := block{"0000000000000000000000000000000000000000000000000000000000000000", "", time.Now().Format("2006-01-02 15:04:05"), "我是创世区块", 1, "000000000"}
	genesisBlock.getHash()
	blockchain = append(blockchain, genesisBlock)
	fmt.Println(blockchain[0])
	i, j := 0, 0
	genTransactions(1)
	for {
		time.Sleep(time.Second)
		newBlock := generateNewBlock(blockchain[i], txQueue, superStarNodesPool[j].address) //  生成新区块
		txQueue = ""
		blockchain = append(blockchain, newBlock) // 主链添加
		fmt.Println(blockchain[i+1])
		i++
		j++
		j = j % len(superStarNodesPool) // 超级节点轮循获得出块权
		if j == 0 {                     // 一轮出块完成
			voting()        // 重新投票
			sortMineNodes() // 选取超级节点
		}
	}
}
