package logger

import (
	"github.com/cihub/seelog"
	"log"
)

func InitLogger(){
	ILogger,err := seelog.LoggerFromConfigAsFile("./seelog.xml")

	if err !=nil{
		seelog.Critical("err parse config log file",err)
		log.Fatal("err parse config log file")
		return
	}
	seelog.ReplaceLogger(ILogger)
}

func LogE(fmt string, v ...interface{}) {
	seelog.Errorf(fmt,v)
}

func LogI(fmt string,v ...interface{}){
	seelog.Infof(fmt,v)
}

func LogW(fmt string,v ...interface{}){
	seelog.Warnf(fmt,v)
}
