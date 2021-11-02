package main

import "github.com/kn4ts/multi_agent_exp"

//const HOST_ADDR = "127.0.0.1" // Hostの待ち受けIPアドレス
const HOST_ADDR = "172.24.137.244"
//const HOST_ADDR = "192.168.179.10"
//const HOST_ADDR = "192.168.179.6"
//const HOST_ADDR = "192.168.11.6"
const HOST_PORT = ":8001" // Hostの待ち受けポート

var AGENT_ADDR = []string{ // エージェントのIPアドレス配列
//	"192.168.11.11",
//	"192.168.11.12",
//	"192.168.11.13",
//	"192.168.11.14"}
	//"192.168.179.11",
	"172.24.137.244:8003",
	"192.168.179.12",
	"192.168.179.13",
	"192.168.179.14"}

const TIME_SAMPLE = 2000            // sampling interval [ms]
var ADJ_MAT = [][]int{{0, 1, 0, 1}, // Adjacency matrix
	{1, 0, 1, 0},
	{0, 1, 0, 1},
	{1, 0, 1, 0}}

func main() {

	// ネットワークオペレータを生成
	nw_operator := multi_agent_exp.NewNetworkOperator(HOST_ADDR, HOST_PORT, AGENT_ADDR)

	// 監視者を並列で起動
	go nw_operator.Supv.Start(TIME_SAMPLE, ADJ_MAT)

	// TCP待ち受け開始
	nw_operator.WaitConnection()

}
