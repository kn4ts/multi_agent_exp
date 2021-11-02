package multi_agent_exp

import (
	"time"
)

type Timer struct {
	interval time.Duration // sampling time [ms]

	tick_ch      chan int
	tick_stop_ch chan int
}

func NewTimer(intvl int) Timer {
	tm := Timer{time.Duration(intvl), make(chan int), make(chan int)}
	return tm
}

func (tm *Timer) TickFunc() {
	ticker := time.NewTicker(tm.interval * time.Millisecond)

LOOP:
	for {
		select {
		case <-ticker.C:
			//fmt.Printf("now -> %v\n", time.Now())
			tm.tick_ch <- 1
		case <-tm.tick_stop_ch:
			//fmt.Println("Timer stop.")
			ticker.Stop()
			break LOOP
		}
	}
}
