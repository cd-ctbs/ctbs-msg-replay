package main

import (
	"log"
	"os"
	"time"
)

var (
	AppLogger *log.Logger // Important information
)

func init() {
	// 创建按日期命名的日志文件
	filename := "log_" + time.Now().Format("20060102") + ".txt"
	f, _ := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	AppLogger = log.New(f,
		"INFO: ",
		log.Ldate|log.Lmicroseconds|log.Lshortfile)
}
