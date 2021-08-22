package multi_agent_exp

import (
	"fmt"
	"encoding/json"
//	"time"
)

type Supervisor struct {
	num_of_agent int

	upstr_Tx_ch chan string
	upstr_Rx_ch chan string
	agent_Tx_ch []chan string
	agent_Rx_ch []chan string

	adj_mat [][]int // 隣接行列

	upstr_Rx_msg UpstreamRxMessage
	agent_Rx_msgs []AgentRxMessage

	agent_Tx_msgs []AgentTxMessage
	//cont_log *contLog
}

type UpstreamRxMessage struct {
	Cmd	string	`json:"cmd"`
	Time	string	`json:"time"`
	Agents	[]Agent_rx	`json:"agent"`
}

type Agent_rx struct {
	Pos_x	float64	`json:"x"`
	Pos_y	float64	`json:"y"`
	Angle	float64	`json:"angle"`
}

type AgentRxMessage struct {
	Time	string	`json:"time"`
	Input	[]float64	`json:"input"`
	OutputEstimate	float64	`json:"y_hat"`
	Residual	float64	`json:"residual"`
}

type AgentTxMessage struct {
	Cmd	string	`json:"cmd"`
	Time	string	`json:"time"`
	AgentSelf	Agent_rx	`json:"agent_self"`
	Agents		[]Agent_tx	`json:"agent"`
}

type Agent_tx struct {
	Input	[]float64	`json:"u"`
	//Output	float64	`json:"y"`
	OutputEstimate	float64	`json:"y_hat"`
	Residual	float64	`json:"residual"`
}

func NewUpstreamMessage(n int) UpstreamRxMessage {
	upmsg := UpstreamRxMessage{}
	for i:=0 ; i<n ;i++ {
		upmsg.Agents = append( upmsg.Agents, Agent_rx{})
	}
	return upmsg
}

func (supv *Supervisor) UpdateTxMessage() []AgentTxMessage {
	agTxmsgs := []AgentTxMessage{}
	// 各エージェントのループ
	for i:=0 ; i<supv.num_of_agent ; i++ {
		agTxs := []Agent_tx{}
		// 接続先エージェントのループ
		for j:=0 ; j<supv.num_of_agent ; j++ {
			// エージェントの通信路をチェック
			if supv.adj_mat[i][j] == 1{
				// 通信路のあるエージェントの情報をまとめる
				agTx := Agent_tx{
					Input:		supv.agent_Rx_msgs[j].Input,
					//Output:		supv.agent_Rx_msgs[j].Output,
					OutputEstimate:	supv.agent_Rx_msgs[j].OutputEstimate,
					Residual:	supv.agent_Rx_msgs[j].Residual,
				}
				agTxs = append( agTxs, agTx)
			}
		}
		agTxmsg := AgentTxMessage{
			Cmd:		supv.upstr_Rx_msg.Cmd,
			Time:		supv.upstr_Rx_msg.Time,
			AgentSelf:	supv.upstr_Rx_msg.Agents[i],
			Agents:		agTxs,
		}
		agTxmsgs = append( agTxmsgs, agTxmsg )
	}
	return agTxmsgs
}

// 監督者のコンストラクタ
func NewSupervisor(n int) *Supervisor {
	// エージェント数と同要素数のチャネル配列を宣言
	ach := make([]chan string, n)
	for i := range ach {
		// 各配列要素を[]byte型のチャネルとして初期化
		ach[i] = make(chan string)
	}

	upmsg := NewUpstreamMessage(n)
	agmsg := []AgentRxMessage{}
	for i:=0 ; i<n ; i++ {
		agmsg = append( agmsg, AgentRxMessage{})
	}

	return &Supervisor{
		num_of_agent: n,

		upstr_Tx_ch: make(chan string),
		upstr_Rx_ch: make(chan string),
		agent_Tx_ch: ach,
		agent_Rx_ch: ach,

		upstr_Rx_msg: upmsg,
		agent_Rx_msgs: agmsg,
		//upstr_msg: upmsg
		//timer: 0,
		//cont_log: cl,
	}
}

func (supv *Supervisor) Start(intvl int, adjmat [][]int) {

	// 隣接行列の定義
	supv.adj_mat = adjmat

	// Ticker の生成と開始
	tm := NewTimer(intvl)
	go tm.TickFunc()

	// Loggerの生成
	lg := NewLogger()

	// 一時変数の定義
	//var agent_RX_temp = make( []string, len(supv.agent_Rx_ch))
	//MAIN:
		for{
			// 上流プログラム側メッセージ受信
			var upstr_RX_temp string
			select {
			case upstr_RX_temp = <- supv.upstr_Rx_ch:
				err := json.Unmarshal([]byte(upstr_RX_temp), &supv.upstr_Rx_msg)
				if err != nil { fmt.Println("unmarshal error (upstrm)")}

				supv.agent_Tx_msgs = supv.UpdateTxMessage()
			default:
			}

			//エージェント側メッセージ受信
			var agent_RX_temp string
			for i := range supv.agent_Rx_ch {
				select {
				case agent_RX_temp = <- supv.agent_Rx_ch[i]:
					err := json.Unmarshal([]byte(agent_RX_temp), supv.agent_Rx_msgs[i])
					if err != nil { fmt.Println("unmarshal error agent:", i)}

					supv.agent_Tx_msgs = supv.UpdateTxMessage()
				default:
				}
			}

			select {
			// 制御周期毎の処理内容
			case <- tm.tick_ch: // Tickerが発火していたら処理に入る

				//fmt.Println(supv.upstr_Rx_msg.Cmd)
				//fmt.Println(len(lg.Log_exist_ch))
				switch supv.upstr_Rx_msg.Cmd {
				case "e":
					if len(lg.Log_exist_ch) > 0 {
						lg.Log_stop_ch <- 1
						fmt.Println("Logger stop")
					}
				case "c":
					if len(lg.Log_exist_ch) == 0 {
						go lg.Start()
						fmt.Println("Logger start")
					}
				default:
					msgj, _ := json.Marshal(&supv.agent_Tx_msgs) // json形式に変換
					fmt.Println(string(msgj))
					if len(lg.Log_exist_ch) > 0 {
						//fmt.Println("Sending message")
						lg.Log_str_ch <- string(msgj)
					}
				}
				// 計測値の更新
				//err := json.Unmarshal([]byte(upstr_RX_temp), supv.upstr_msg)
				//if err != nil {
				//	fmt.Println(err)
				//	return
				//}
				//fmt.Println(supv.upstr_Rx_msg)
			default:
				//fmt.Println("Supevisor")
			}
		}
}
