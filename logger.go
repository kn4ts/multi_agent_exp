package multi_agent_exp

import (
	"os"
	//"fmt"
	"time"
)

type Logger struct{
	Layout	string
	Dirname	string
	Filname	string

	Log_str_ch	chan string
	Log_stop_ch	chan int
	Log_exist_ch	chan int
}

func NewLogger() *Logger {
	const layout = "data_2006-01-02_15-04-05.csv"
	return &Logger{
		Layout: layout,
		Dirname: "exp_data",
		Filname: time.Now().Format(layout),

		Log_str_ch: make(chan string),
		Log_stop_ch: make(chan int),
		Log_exist_ch: make(chan int, 1),
	}
}

func ( lg *Logger) MakeFile() *os.File{

	lg.Filname = time.Now().Format(lg.Layout)
	
	if _, err := os.Stat(lg.Dirname); os.IsNotExist(err) {
		os.Mkdir(lg.Dirname, 0777)
	}
	file, err := os.Create(lg.Dirname+"/"+lg.Filname)
	if err != nil {
		// Openエラー処理
	}
	return file
}

func ( lg *Logger) Start(){

	if len(lg.Log_exist_ch) == 0 {

		fl := lg.MakeFile()
		defer fl.Close()

		lg.Log_exist_ch <- 1
	
		LOOP:
			for{
				select {
				case log_temp := <- lg.Log_str_ch:
					fl.Write(([]byte)(log_temp+"\n"))
				case <- lg.Log_stop_ch:
					break LOOP
				default:
				}
			}
		<- lg.Log_exist_ch
	}
}
