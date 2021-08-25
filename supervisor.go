package multi_agent_exp

import (
	"encoding/json"
	"fmt"
)

/* ////////////////////////////////////////////////////////////
	監視者の構造体
//////////////////////////////////////////////////////////// */
type Supervisor struct {
	num_of_agent int     // エージェントの数
	adj_mat      [][]int // 隣接行列

	upstr_exist_ch chan int   // 上流プログラムの存在判定用チャネル
	agent_exist_ch []chan int // エージェントの存在判定用チャネル配列

	upstr_Tx_ch chan string   // 上流側送信文字列のチャネル
	upstr_Rx_ch chan string   // 上流側受信文字列のチャネル
	agent_Tx_ch []chan string // エージェント側送信文字列のチャネルの配列
	agent_Rx_ch []chan string // エージェント側受信文字列のチャネルの配列

	upstr_Rx_msg  UpstreamRxMessage // 上流側受信メッセージ格納用の構造体
	agent_Rx_msgs []AgentRxMessage  // エージェント側受信メッセージ格納用の構造体

	agent_Tx_msgs []AgentTxMessage // エージェント側送信メッセージ格納用の構造体
}

// 上流側受信メッセージ格納用の構造体
type UpstreamRxMessage struct {
	Cmd    string     `json:"cmd"`    // コマンド文字列
	Time   string     `json:"time"`   // 時刻文字列
	Agents []Agent_rx `json:"agents"` // 上流側エージェント計測値格納用の構造体配列
}

// 上流側エージェント計測値格納用の構造体
type Agent_rx struct {
	Pos_x float64 `json:"x"`
	Pos_y float64 `json:"y"`
	Angle float64 `json:"angle"`
}

// エージェント側受信メッセージ格納用の構造体
type AgentRxMessage struct {
	Time           string    `json:"time"`     // 時刻文字列
	Input          []float64 `json:"input"`    // 制御入力信号
	OutputEstimate float64   `json:"y_hat"`    // 出力推定値
	Residual       float64   `json:"residual"` // 残差信号
}

// エージェント側送信メッセージ格納用の構造体
type AgentTxMessage struct {
	Cmd       string     `json:"cmd"`        // コマンド文字列
	Time      string     `json:"time"`       // 時刻文字列
	AgentSelf Agent_rx   `json:"agent_self"` // 自身(i)の計測値格納用の構造体
	Agents    []Agent_tx `json:"agents"`     // 隣接エージェント(j)の情報格納用の構造体配列
}

// 隣接エージェント(j)の情報格納用の構造体
type Agent_tx struct {
	Input []float64 `json:"u"`
	//Output	float64	`json:"y"`
	OutputEstimate float64 `json:"y_hat"`
	Residual       float64 `json:"residual"`
}

// 上流側メッセージ格納用の構造体の生成関数
func NewUpstreamMessage(n int) UpstreamRxMessage {
	upmsg := UpstreamRxMessage{}
	for i := 0; i < n; i++ { // エージェントの数だけ受信メッセージ用構造体を用意する
		upmsg.Agents = append(upmsg.Agents, Agent_rx{})
	}
	return upmsg
}

/* =========================================================
  送信用メッセージ配列を更新・作成するメソッド
========================================================= */
func (supv *Supervisor) UpdateTxMessage() []AgentTxMessage {
	agTxmsgs := []AgentTxMessage{}
	// 各エージェントのループ
	for i := 0; i < supv.num_of_agent; i++ {
		agTxs := []Agent_tx{}
		// 接続先エージェントのループ
		for j := 0; j < supv.num_of_agent; j++ {
			// エージェントの通信路をチェック
			if supv.adj_mat[i][j] == 1 {
				// 通信路のあるエージェントの情報を追加
				agTx := Agent_tx{
					Input: supv.agent_Rx_msgs[j].Input,
					//Output:		supv.agent_Rx_msgs[j].Output,
					OutputEstimate: supv.agent_Rx_msgs[j].OutputEstimate,
					Residual:       supv.agent_Rx_msgs[j].Residual,
				}
				agTxs = append(agTxs, agTx)
			}
		}
		// 各エージェントへの送信用メッセージの生成
		agTxmsg := AgentTxMessage{
			Cmd:       supv.upstr_Rx_msg.Cmd,
			Time:      supv.upstr_Rx_msg.Time,
			AgentSelf: supv.upstr_Rx_msg.Agents[i],
			Agents:    agTxs,
		}
		agTxmsgs = append(agTxmsgs, agTxmsg)
	}
	// 更新した構造体の配列を返す
	return agTxmsgs
}

/* =========================================================
  各エージェントへメッセージを送信するメソッド
========================================================= */
func (supv *Supervisor) AllocateMessageToAgent() {
	// 各エージェントのループ
	for i := 0; i < supv.num_of_agent; i++ {
		// エージェントが存在するなら
		if len(supv.agent_exist_ch[i]) > 0 {
			msgj, _ := json.Marshal(&supv.agent_Tx_msgs[i]) // （構造体からjson形式へ変換）
			if string(msgj) != "null" {
				fmt.Println("Message to agent ", i, " :", string(msgj)) // デバッグ用表示
			}
			// メッセージ送信
			supv.agent_Tx_ch[i] <- string(msgj)
		}
	}
}

/* =========================================================
  ロガーへメッセージを送信するメソッド
========================================================= */
func (supv *Supervisor) SendMessageToLogger(lg *Logger, msgj []byte) {
	// ロガーが存在するなら記録用メッセージを送る
	if len(lg.Log_exist_ch) > 0 {
		lg.Log_str_ch <- string(msgj)
		fmt.Println("Message to logger :", string(msgj)) // デバッグ用表示
	}
}

// 監視者のコンストラクタ（生成・初期化関数）
func NewSupervisor(n int) *Supervisor {
	// エージェント数と同要素数のチャネル配列を宣言
	ech := make([]chan int, n)
	sch := make([]chan string, n)

	// エージェント数と同要素数のチャネル配列を宣言
	upmsg := NewUpstreamMessage(n)
	agRmsg := []AgentRxMessage{}
	agTmsg := []AgentTxMessage{}

	for i := 0; i < n; i++ {
		ech[i] = make(chan int, 1) // 各配列要素をint型でバッファ1のチャネルとして初期化
		sch[i] = make(chan string) // 各配列要素をstring型のチャネルとして初期化

		agRmsg = append(agRmsg, AgentRxMessage{})
		agTmsg = append(agTmsg, AgentTxMessage{})

	}

	// 監視者の構造体のポインタを返す
	return &Supervisor{
		num_of_agent: n,

		upstr_exist_ch: make(chan int, 1), // 上流プログラムの存在判定用チャネル
		agent_exist_ch: ech,               // エージェントの存在判定用チャネル配列

		upstr_Tx_ch: make(chan string),
		upstr_Rx_ch: make(chan string),
		agent_Tx_ch: sch,
		agent_Rx_ch: sch,

		upstr_Rx_msg:  upmsg,
		agent_Rx_msgs: agRmsg,

		agent_Tx_msgs: agTmsg, // エージェント側送信メッセージ格納用の構造体
	}
}

/* =========================================================
  監視者の実行メソッド
========================================================= */
func (supv *Supervisor) Start(intvl int, adjmat [][]int) {

	// 隣接行列の定義
	supv.adj_mat = adjmat

	// Ticker の生成と開始
	tm := NewTimer(intvl)
	go tm.TickFunc() // タイマーを並列で実行

	// Loggerの生成
	lg := NewLogger()

MAIN:
	for {
		// 上流プログラム側メッセージ待ち受け
		var upstr_RX_temp string // 一時文字列を定義
		select {
		case upstr_RX_temp = <-supv.upstr_Rx_ch: // 受信メッセージがあれば以下を実行
			// 受信した文字列を指定の構造体に変換して格納
			err := json.Unmarshal([]byte(upstr_RX_temp), &supv.upstr_Rx_msg)
			if err != nil {
				fmt.Println("unmarshal error (upstrm)")
			}

			supv.agent_Tx_msgs = supv.UpdateTxMessage() // 送信用構造体を更新

		default: // 受信メッセージがなければ流す
		}

		//エージェント側メッセージ待ち受け
		var agent_RX_temp string // 一時文字列を定義
		for i := range supv.agent_Rx_ch {
			select {
			case agent_RX_temp = <-supv.agent_Rx_ch[i]: // 受信メッセージがあれば以下を実行

				// 受信した文字列を指定の構造体に変換して格納
				err := json.Unmarshal([]byte(agent_RX_temp), supv.agent_Rx_msgs[i])
				if err != nil {
					fmt.Println("unmarshal error agent:", i)
				}

				supv.agent_Tx_msgs = supv.UpdateTxMessage() // 送信用構造体を更新
			default: // 受信メッセージがなければ流す
				//supv.upstr_Rx_msg.Cmd = ""
			}
		}

		select {
		// 制御周期毎の処理内容
		case <-tm.tick_ch: // Tickerが発火していたら処理に入る

			// コマンドによって実行動作を切り替え
			switch supv.upstr_Rx_msg.Cmd {
			case "q": // プログラム終了コマンド
				if len(lg.Log_exist_ch) > 0 {
					lg.Log_stop_ch <- 1
					fmt.Println("Logger stop")
				}
				break MAIN
			case "e": // 制御終了コマンド
				// ロガーが存在するならばロガーを停止する
				if len(lg.Log_exist_ch) > 0 {
					lg.Log_stop_ch <- 1
					fmt.Println("Logger stop")
				}
			case "c": // 制御開始コマンド
				// ロガーが存在しないならばロガーを開始する
				if len(lg.Log_exist_ch) == 0 {
					go lg.Start()
					fmt.Println("Logger start")
				}
			default: // それ以外のコマンドの場合
			}

			// 各エージェントへメッセージを送信
			supv.AllocateMessageToAgent()

			msgj, _ := json.Marshal(&supv.upstr_Rx_msg) // （構造体からjson形式へ変換）
			// ロガーへメッセージを送信
			supv.SendMessageToLogger(lg, msgj)

			if supv.upstr_Rx_msg.Cmd == "" {
				fmt.Println("Message recieved : none") // デバッグ用表示
			} else {
				// デバッグ用表示メッセージ
				//msgj, _ := json.Marshal(&supv.agent_Tx_msgs) // （構造体からjson形式へ変換）
				if string(msgj) != "null" {
					fmt.Println("Message recieved :", string(msgj)) // デバッグ用表示
				}
			}


		default:
			//fmt.Println("Supevisor")
		}
	}
}
