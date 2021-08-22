package multi_agent_exp

import (
	"fmt"
	"log"
	"net"
	"strings"
)

type NetworkOperator struct {
	Upstr_exist_ch chan int   // 上流プログラムの存在判定用チャネル
	agent_exist_ch []chan int // エージェントの存在判定用チャネル
	num_connect_ch chan int   // TCP接続数の制限用チャネル

	host_ip  *HostIP  // ホストIPとポート保存用構造体
	agent_ip *AgentIP // エージェントIPの保存用構造体

	Supv *Supervisor // 監視者用の構造体
}

type HostIP struct {
	addr string
	port string
}

type AgentIP struct {
	addr []string
}

// オペレータのコンストラクタ（生成＆初期化関数）
func NewNetworkOperator(haddr, hport string, aaddr []string) *NetworkOperator {
	nh := NewHostIP(haddr, hport)
	na := NewAgentIP(aaddr)

	num_of_agent := len(aaddr)
	// 監視者構造体の初期化
	ns := NewSupervisor(num_of_agent)

	ach := make([]chan int, num_of_agent) // エージェント数と同要素数のチャネル配列を宣言
	for i := range ach {
		ach[i] = make(chan int) // 各配列要素をint型のチャネルとして初期化
	}
	return &NetworkOperator{
		Upstr_exist_ch: make(chan int),
		agent_exist_ch: ach,
		num_connect_ch: make(chan int, num_of_agent+1), // エージェント数+1のバッファを持つチャネルを宣言

		host_ip:  nh,
		agent_ip: na,

	//	adj_mat: adjmat,

		Supv: ns,
	}
}

// ホストIP構造体の生成
func NewHostIP(addr, port string) *HostIP {
	return &HostIP{
		addr: addr,
		port: port,
	}
}

// エージェントIP構造体の生成
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
  TCP接続の待ち受け関数
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
		// 受理した接続を扱う関数に投げる
		go opr.DealConnection(conn)
	}
}

func (opr *NetworkOperator) DealConnection(conn net.Conn) {
	/* =========================================================
	  接続IPアドレスによる役割の振り分け
	========================================================= */
	var connected_address = conn.RemoteAddr().String()
	fmt.Println("Accept connection %s", connected_address)

	if strings.HasPrefix(connected_address, opr.host_ip.addr) {
		fmt.Println("%s", opr.host_ip.addr)
		//opr.Upstr_exist_ch <- 1
		fmt.Printf("Hello, upstream\n")

		// 上流プログラム用クライアント
		opr.UpstreamClient(conn)

		fmt.Printf("Bye, upstream\n")
		//<-opr.Upstr_exist_ch
	} else {
		// どのエージェントかの判定
		for i := 0; i < len(opr.agent_ip.addr); i++ {
			if strings.HasPrefix(connected_address, opr.agent_ip.addr[i]) {
				opr.agent_exist_ch[i] <- 1
				fmt.Printf("Hello, agent %d\n", i+1)

				// エージェント用クライアント
				opr.AgentClient(conn, i)

				<-opr.agent_exist_ch[i]
				fmt.Printf("Bye, agent %d\n", i+1)
				break
			} else {
				fmt.Printf("未知のクライアントです．ADDRESS:%s\n", connected_address)
			}
		}
	}
	<-opr.num_connect_ch
}

/* ////////////////////////////////////////////////////////////
	上流プログラム側のTCPクライアント処理関数
	（最新の計測値が充分速い速度で送られてくることを想定）
//////////////////////////////////////////////////////////// */
func (opr *NetworkOperator) UpstreamClient(conn net.Conn) {
	for {
		buf := make([]byte, 1024)
		// Readerを作成して、送られてきたメッセージを出力する
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("tcp read error in Upstream: %s \n", conn.RemoteAddr().String())
			break
		}
		// 受信メッセージをスーパバイザに送る
		opr.Supv.upstr_Rx_ch <- string(buf[:n])
		//opr.supv.upstr_Rx_ch <- string(buf[:n])
	}
}

/* ////////////////////////////////////////////////////////////
	エージェント側のTCPクライアント処理関数
	（制御周期ごとに計算値が送られてくることを想定）
//////////////////////////////////////////////////////////// */
func (opr *NetworkOperator) AgentClient(conn net.Conn, i int) {
	fmt.Printf("agent %d...\n", i)
	for {
		// TCP受信
		buf := make([]byte, 1024)
		// Readerを作成して、送られてきたメッセージを出力する
		n, err := conn.Read(buf)
		if err != nil {
			fmt.Printf("tcp read error in agent %d: %s \n", i, conn.RemoteAddr().String())
			break
		}
		// 受信メッセージをスーパバイザに送る
		opr.Supv.agent_Rx_ch[i] <- string(buf[:n])
		//opr.supv.agent_Rx_ch <- string(buf[:n])

		// 送信メッセージをスーパバイザから受け取る
		//tcp_send_str := <- opr.supv.agent_Tx_ch
		tcp_send_str := <-opr.Supv.agent_Tx_ch[i]
		// TCP送信
		_, err = conn.Write([]byte(tcp_send_str))
		if err != nil {
			fmt.Printf("tcp send error in agent %d: %s \n", i, conn.RemoteAddr().String())
			break
		}

	}
}
