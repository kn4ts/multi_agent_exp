package multi_agent_exp

import (
	"fmt"
	"log"
	"net"
	"strings"
)

/* ////////////////////////////////////////////////////////////
	ネットワークオペレータの構造体
//////////////////////////////////////////////////////////// */
type NetworkOperator struct {
//	Upstr_exist_ch chan int   // 上流プログラムの存在判定用チャネル
//	agent_exist_ch []chan int // エージェントの存在判定用チャネル配列
	num_connect_ch chan int   // TCP接続数の制限用チャネル

	host_ip  *HostIP  // ホストIPとポート保存用構造体
	agent_ip *AgentIP // エージェントIPの保存用構造体

	Supv *Supervisor // 監視者用の構造体
}

// サーバのIPアドレスとポートの格納用構造体
type HostIP struct {
	addr string
	port string
}

// エージェントのIPアドレスを格納する構造体
type AgentIP struct {
	addr []string
}

// オペレータのコンストラクタ（生成・初期化関数）
func NewNetworkOperator(haddr, hport string, aaddr []string) *NetworkOperator {
	nh := NewHostIP(haddr, hport) // サーバIPアドレス，ポートを格納
	na := NewAgentIP(aaddr) // エージェントのIPアドレスを格納

	num_of_agent := len(aaddr) // エージェントの数をアドレス数から取得

	// 監視者構造体の生成・初期化
	ns := NewSupervisor(num_of_agent)

	ach := make([]chan int, num_of_agent) // エージェント数と同要素数のチャネル配列を宣言
	for i := range ach {
		ach[i] = make(chan int, 1) // 各配列要素をint型のチャネルとして初期化
	}
	return &NetworkOperator{
		//Upstr_exist_ch: make(chan int, 1),
		//agent_exist_ch: ach,
		num_connect_ch: make(chan int, num_of_agent+1), // エージェント数+1のバッファを持つチャネルを宣言

		host_ip:  nh,
		agent_ip: na,

		Supv: ns,
	}
}

// ホストIP・ポート格納構造体の生成・初期化
func NewHostIP(addr, port string) *HostIP {
	return &HostIP{
		addr: addr,
		port: port,
	}
}

// エージェントIP格納構造体の生成・初期化
func NewAgentIP(addr []string) *AgentIP {
	n := len(addr)
	ad := make([]string, n)
	for i := 0; i < n; i++ {
		ad[i] = addr[i]
	}
	return &AgentIP{
		addr: ad,
	}
}

/* =========================================================
  TCP接続の待ち受けメソッド
========================================================= */
func (opr *NetworkOperator) WaitConnection() {
	// Hostアドレス+ポート文字列の生成
	addr_port := opr.host_ip.addr + opr.host_ip.port

	// TCP接続リッスン開始
	listen, err := net.Listen("tcp", addr_port)
	if err != nil {
		log.Fatalf("TCP接続のリッスンに失敗しました\n")
	}
	fmt.Printf("%sで受付開始しました\n", addr_port)

	for {
		// TCP接続数上限になっていればブロック
		opr.num_connect_ch <- 1
		// TCP接続を受理
		conn, err := listen.Accept()
		if err != nil {
			log.Fatal("TCP接続を確立できませんでした\n")
		}
		// 受理した接続を扱うメソッドを並列で実行
		go opr.DealConnection(conn)
	}
}

/* =========================================================
  受理した接続を扱うメソッド
========================================================= */
func (opr *NetworkOperator) DealConnection(conn net.Conn) {
	var connected_address = conn.RemoteAddr().String()
	//fmt.Println("Accept connection ", connected_address) // デバッグ用表示関数

	// 接続IPアドレスによる役割の振り分け
	if strings.HasPrefix(connected_address, opr.host_ip.addr) {
		// fmt.Println(opr.host_ip.addr) // デバッグ用表示関数
		opr.Supv.upstr_exist_ch <- 1 // 上流プログラムが既に存在すればブロック
		fmt.Printf("Hello, upstream\n")

		// 上流プログラム用クライアントメソッドの実行
		opr.UpstreamClient(conn)

		<-opr.Supv.upstr_exist_ch // 上流プログラムの存在判定チャネルの値を解放
		fmt.Printf("Bye, upstream\n")
	} else {
		// どのエージェントかの判定
		for i := 0; i < len(opr.agent_ip.addr); i++ {
			if strings.HasPrefix(connected_address, opr.agent_ip.addr[i]) {
				opr.Supv.agent_exist_ch[i] <- 1 // エージェントiが既に存在すればブロック
				fmt.Println("Hello, agent ", i+1)

				// エージェント用クライアントメソッドの実行
				opr.AgentClient(conn, i)

				<-opr.Supv.agent_exist_ch[i] // エージェントの存在判定チャネルから値を解放
				fmt.Println("Bye, agent ", i+1)
				break
			} else { // 接続先がどれでもなければ弾く
				fmt.Println("未知のクライアントです．ADDRESS:", connected_address)
			}
		}
	}
	conn.Close() // 接続を閉じる
	<-opr.num_connect_ch // TCP接続の枠を1つ解放
}

/* ////////////////////////////////////////////////////////////
	上流プログラム側のTCPクライアント処理メソッド
	（最新の計測値が充分速い速度で送られてくることを想定）
//////////////////////////////////////////////////////////// */
func (opr *NetworkOperator) UpstreamClient(conn net.Conn) {
	for {
		// 一時変数を定義
		buf := make([]byte, 1024)
		// Readerを作成して，送られてきたメッセージを格納する
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("TCP read error in Upstream: ", conn.RemoteAddr().String())
			break
		}
		// 受信メッセージをスーパバイザに送る
		opr.Supv.upstr_Rx_ch <- string(buf[:n])
		//opr.supv.upstr_Rx_ch <- string(buf[:n])
	}
}

/* ////////////////////////////////////////////////////////////
	エージェント側のTCPクライアント処理メソッド
	（制御周期ごとに計算値が送られてくることを想定）
//////////////////////////////////////////////////////////// */
func (opr *NetworkOperator) AgentClient(conn net.Conn, i int) {
	//fmt.Println("agent ", i)
	for {
		// TCP受信
		buf := make([]byte, 1024)
		// Readerを作成して、送られてきたメッセージを出力する
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Println("TCP read error in agent ", i, " address:", conn.RemoteAddr().String())
			break
		}
		// 受信メッセージを監視者に送る
		opr.Supv.agent_Rx_ch[i] <- string(buf[:n])
		//opr.supv.agent_Rx_ch <- string(buf[:n])

		// 送信メッセージを監視者から受け取る
		tcp_send_str := <-opr.Supv.agent_Tx_ch[i]
		// TCP送信
		_, err = conn.Write([]byte(tcp_send_str))
		if err != nil {
			fmt.Println("TCP send error in agent ", i, "address", conn.RemoteAddr().String())
			break
		}
	}
}
