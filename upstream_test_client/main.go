package main

import (
	"fmt"
//	"log"
	"net"
	"encoding/json"
	//"strings"
	"time"
	"os"
)

const NA = 4

type Message struct {
	Cmd	string	`json:"cmd"`
	Time	string	`json:"time"`
	Agents	[]Agent	`json:"agents"`
}

type Agent struct {
	Pos_x	float64	`json:"x"`
	Pos_y	float64	`json:"y"`
	Angle	float64	`json:"angle"`
}

func NewAgent(x,y,th float64) Agent {
	a := Agent{ x, y, th }
	return a
}

type Timer struct {
	interval time.Duration // sampling time [ms]

	tick_ch chan int
	tick_stop chan int
}

func NewTimer(intvl int) Timer{
	tm := Timer{ time.Duration(intvl), make(chan int), make(chan int) }
	return tm
}

func ( tm *Timer) TickFunc() {
	ticker := time.NewTicker( tm.interval * time.Millisecond )

	LOOP:
		for {
			select {
			case <-ticker.C:
				//fmt.Printf("now -> %v\n", time.Now())
				tm.tick_ch <- 1
			case <-tm.tick_stop:
				//fmt.Println("Timer stop.")
				ticker.Stop()
				break LOOP
			}
		}
}

func main(){
	const layout = "15-04-05.000"

	// メッセージの初期化
	msg := Message{}
	for i:=0 ; i<NA ; i++ {
		msg.Agents = append(msg.Agents, NewAgent(float64(i),0,1))
	}

	// TCP接続
	//conn, _ := net.Dial("tcp", "localhost:8001")
	//conn, _ := net.Dial("tcp", "192.168.11.6:8001")
	//conn, _ := net.Dial("tcp", "172.24.137.244:8001")
	raddr, err := net.ResolveTCPAddr("tcp", "172.24.137.244:8001" )
	 if err != nil {
		fmt.Println("net resolve TCP Addr error ")
		os.Exit(1)
	}
	laddr, err := net.ResolveTCPAddr("tcp", "172.24.137.244:8002" )
	if err != nil {
		fmt.Println("net resolve TCP Addr error ")
		os.Exit(1)
	}
	//conn, _ := net.Dial("tcp", "172.24.137.244:8001")
	conn, _ := net.DialTCP("tcp", laddr, raddr)

	//conn, _ := net.Dial("tcp", "192.168.179.10:8001")
	//conn, _ := net.Dial("tcp", "127.0.0.1:8001")
	ch := make(chan string)

	// タイマーの初期化と実行
	//tim := NewTimer(100)
	tim := NewTimer(2000)
	go tim.TickFunc()

	// キーボード入力待ち関数のスタート
	go func() {
		//var res string
		var keyinput string
		for {
			fmt.Scan(&keyinput)
			if keyinput == "e" {
				ch <- "e"
				//res = fmt.Sprintf("toggle\r")
				//conn.Write([]byte(res))
			} else if keyinput == "c" {
				ch <- "c"
				//res = fmt.Sprintf("start\r")
				//conn.Write([]byte(res))
			} else if keyinput == "q"{
				ch <- "q"
			}else {
				ch <- "0"
			}
		}
	}()

	// 受信データ表示関数のスタート
	//go func() {
	//	for{
	//		//buf := make([]byte, 1024)
	//		//// Readerを作成して、送られてきたメッセージを出力する
	//		//n, err := conn.Read(buf)
	//		//if err != nil {
	//		//	break
	//		//}
	//		//fmt.Printf("Recieve message: %s\n",string(buf[:n]))
	//	}
	//}()

	var cmd string
	// 繰り返し処理部分
	for {

		select {
		case cmd = <- ch:
		default:
			//cmd = "none"
		}

		// タイマー毎に実行
		select {
		case <- tim.tick_ch:
			msg.Cmd = cmd
			msg.Time = time.Now().Format(layout)
			msgj, _ := json.Marshal(&msg)
			fmt.Println(string(msgj))

			conn.Write(msgj)
			//conn.Write([]byte(msgj))

			cmd = "none"
			msg.Cmd = cmd
		default:
		}

	}
	conn.Close()
}
