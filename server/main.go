package main

import "github.com/kn4ts/multi_agent_exp"

const HOST_ADDR = "127.0.0.1"
//const HOST_ADDR = "192.168.179.10"
const HOST_PORT = ":8001"
//const HOST_PORT = ":8001"

var AGENT_ADDR = []string{
	"192.168.179.11",
	"192.168.179.12",
	"192.168.179.13",
	"192.168.179.14"}

const TIME_SAMPLE = 2000 // sampling interval [ms]
var ADJ_MAT = [][]int{{0, 1, 0, 1},  // Adjacency matrix
		      {1, 0, 1, 0},
		      {0, 1, 0, 1},
		      {1, 0, 1, 0}}

func main() {

	// 上流側センサの数と下流側エージェントの数を定義
	//num_of_sense := 1
	//num_of_agent := len(AGENT)

	//fmt.Printf("Num of agent is %d \n",num_of_agent)

	// ネットワークオペレータを生成
	nw_operator := multi_agent_exp.NewNetworkOperator( HOST_ADDR, HOST_PORT, AGENT_ADDR )
	//nw_operator := network_operator.NewNetWorkOperator(HOST_ADDR, HOST_PORT, AGENT_ADDR)
	// 制御器を起動
	go nw_operator.Supv.Start( TIME_SAMPLE, ADJ_MAT )
	// TCP待ち受け配置
	nw_operator.WaitConnection()

}
