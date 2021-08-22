package main

import (
	"fmt"
	"encoding/json"
//	"strings"
//	"time"
)

const NA = 4

type Message struct {
	Cmd	int	`json:"cmd"`
	Time	string	`json:"time"`
	Agents	[]Agent	`json:"agent"`
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

func main() {
	
	msg := Message{}
	msg.Agents = append( msg.Agents, NewAgent(1, 2, 5.5))
	msg.Agents = append( msg.Agents, NewAgent(2.5, 3, 7.1))
	
	fmt.Println(msg)
	msgj, _ := json.Marshal(&msg)
	fmt.Println(string(msgj))

}
