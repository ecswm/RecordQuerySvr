package main

import (
	"dbhandler"
	"fmt"
	"io"
	"net/http"
	"os"
	_ "strings"
	"time"
	"log"
	"github.com/cihub/seelog"
)

var fsclient *FsClient = nil
var config Configuration

func insertrecording(rw http.ResponseWriter, req *http.Request) {
	seelog.Infof("recv insertrecording request,the request info is: [%s]",req.URL.Path)
	err := req.ParseForm()
	if err != nil {
		seelog.Errorf("parse insertrecording request occur err,err is: [%s]",err.Error())
		return
	}
	callid := req.FormValue("callid")
	callernumber := req.FormValue("callernumber")
	callednumber := req.FormValue("callednumber")
	recordingpath := req.FormValue("recordingpath")

	seelog.Infof("parseform insertrecording request,data is {callid:%s,caller_number:%s,called_number:%s,recording_path:%s}",callid,callernumber,callednumber,recordingpath)
	recordinginfo := dbhandler.RecordingInfo{callid,
		recordingpath,
		time.Now().Format("2006-01-02 15:04:05"),
		callernumber,
		callednumber}

	ret, err := dbhandler.DbObj.InsertRecordingInfo(recordinginfo)
	rsp := dbhandler.GenerateRspJson(&recordinginfo, "", "0")
	if ret != true {
		seelog.Errorf("insertrecording occur err,err is: [%s]",err.Error())
		rsp = dbhandler.GenerateRspJson(&recordinginfo, err.Error(), "0")
	}
	rw.Header().Set("content-type", "application/json")
	rw.Write(rsp)
}

func queryrecording(rw http.ResponseWriter, req *http.Request) {
	seelog.Infof("recv queryrecording request,the request info is: [%s]",req.URL.Path)
	err := req.ParseForm()
	if err != nil {
		seelog.Errorf("parse queryrecording request occur err,err is: [%s]",err.Error())
		return
	}
	callid := req.FormValue("call_id")
	seelog.Infof("parseform queryrecording request,data is {callid:%s}",callid)
	recordingpath, err := dbhandler.DbObj.QueryRecordingPath(callid)
	rw.Header().Set("content-type", "application/json")
	if err != nil {
		seelog.Errorf("queryrecording occur err,err is: [%s]",err.Error())
		rsp := dbhandler.GenerateRspJson(nil, err.Error(), "")
		rw.Write(rsp)
		return
	}

	f, err := os.Open(recordingpath)
	if err != nil {
		seelog.Errorf("open recordingfile occur err,err is: [%s]",err.Error())
		rsp := dbhandler.GenerateRspJson(nil, err.Error(), "")
		rw.Write(rsp)
		return
	}
	defer f.Close()

	rw.Header().Set("content-type", "audio/wav")
	bytes, err := io.Copy(rw, f)
	if err != nil {
		seelog.Errorf("copy recordingfile to client occur err,err is: [%s]",err.Error())
	} else {
		seelog.Infof("send %v to %v succes with %v bytes\n",req.URL.Path,req.RemoteAddr,bytes)
	}
}

func doublecall(rw http.ResponseWriter, req *http.Request) {
	seelog.Infof("recv doublecall request,the request info is:[%s]",req.URL.Path)
	err := req.ParseForm()

	if err != nil {
		seelog.Errorf("parse doublecall request occur err,err is: [%s]",err.Error())
		return
	}

	var (
		caller_number string
		called_number string
		sigparams     string
		authorization string
		user 	string
		ret     bool
	)
	if req.Method == "POST" {
		caller_number, called_number = req.PostFormValue("Caller_number"), req.PostFormValue("Called_number")

	}
	if req.Method == "GET"{
		caller_number,called_number = req.FormValue("Caller_number"),req.FormValue("Called_number")
	}
	sigparams = req.FormValue("SigParameter")
	authorization = req.Header.Get("Authorization")

	if user, ret = config.DecodeSigParams(sigparams, authorization); ret == false {
		//检验未通过
		seelog.Errorf("user authentication failed,appname:doublecall,cause:user is not exist or sigparams is incorrect")
		rw.Header().Set("content-type", "text/plain")
		rw.WriteHeader(http.StatusForbidden)
		rw.Write([]byte("invalid user or sigparams"))
		return
	}
	seelog.Infof("parse doublecall data is {user:%s,caller_number:%s,called_number:%s,sigs:%s}",user,caller_number,called_number,sigparams)

	app := DoubleCallApp{CallerNumber:caller_number,CalledNumber:called_number}
	app.CallApp = CallApp{
		Prefix:config.GetUserPrefix(user),
		AppName:"doublecall",
		CustomerName:user,
		Ani:config.GetUserAni(user),
		PushResultUrl:"",
		Response: make(chan *CallAppRsp),
		Job_Uuid: make(chan string),
		Err:      make(chan error)}
	fsclient.PushAppRequest(&app)

	rw.Header().Set("content-type", "application/json")
	var (
		callrsp *CallAppRsp
		job_uuid      string
	)
	select {
	case err = <-app.Err:
		rw.Write([]byte(err.Error()))
	case job_uuid = <-app.Job_Uuid:
		rw.Write([]byte(job_uuid))
	case <-time.After(time.Second * 100):
		rw.Write([]byte("RequestTimeout"))
	case callrsp = <-app.Response:
		if retjson, err := callrsp.GenerateRspJson(); err == nil {
			rw.Write([]byte(retjson))
		}
	}
}

func voiceidentcall(rw http.ResponseWriter, req *http.Request)  {
	seelog.Infof("recv voiceidentcall request,the request info is: [%s]",req.URL.Path)

	err:=req.ParseForm()
	if err!=nil{
		seelog.Errorf("parse voiceidentcall request occur err,err is: [%s]",err.Error())
		return
	}

	var(
		called_number   string
		ident_code 	string
		sigparams	string
		authorization 	string
		user 		string
		ret  		bool
	)
	if req.Method == "POST"{
		called_number,ident_code = req.PostFormValue("called_number"), req.PostFormValue("ident_code")
	}
	if req.Method == "GET"{
		called_number,ident_code = req.FormValue("called_number"),req.FormValue("ident_code")
	}
	sigparams = req.FormValue("SigParameter")
	authorization = req.Header.Get("Authorization")

	if user, ret = config.DecodeSigParams(sigparams, authorization); ret == false {
		//检验未通过
		seelog.Errorf("user authentication failed,appname:voiceidentcall,cause:user is not exist or sigparams is incorrect")
		rw.Header().Set("content-type", "text/plain")
		rw.WriteHeader(http.StatusForbidden)
		rw.Write([]byte("invalid user"))
		return
	}
	seelog.Infof("parseform voiceident data is {user:%s, called_number: %s, ident_code: %s, sigs: %s}", user,called_number,ident_code, sigparams)

	app := VoiceIdentCallApp{IdentCode:ident_code,CalledNumber:called_number}
	app.CallApp = CallApp{AppName:"voiceidentcall",
		Prefix:config.GetUserPrefix(user),
		CustomerName:user,
		Ani:config.GetUserAni(user),
		PushResultUrl:"",
		Response: make(chan *CallAppRsp),
		Job_Uuid: make(chan string),
		Err:      make(chan error)}
	fsclient.PushAppRequest(&app)

	rw.Header().Set("content-type", "application/json")
	var (
		callrsp *CallAppRsp
		job_uuid      string
	)
	select {
	case err = <-app.Err:
		rw.Write([]byte(err.Error()))
	case job_uuid = <-app.Job_Uuid:
		rw.Write([]byte(job_uuid))
	case <-time.After(time.Second * 100):
		rw.Write([]byte("RequestTimeout"))
	case callrsp = <-app.Response:
		if retjson, err := callrsp.GenerateRspJson(); err == nil {
			rw.Write([]byte(retjson))
		}
	}

}

func outboundcall(rw http.ResponseWriter,req *http.Request){
	seelog.Infof("recv outboundcall request,the request info is: [%s]",req.URL.Path)

	err:=req.ParseForm()
	if err!=nil{
		seelog.Errorf("parse outboundcall request occur err,err is: [%s]",err.Error())
		return
	}

	var(
		called_number   string
		caller_number 	string
		sigparams	string
		authorization 	string
		user 		string
		ret  		bool
	)
	if req.Method == "POST"{
		called_number,caller_number = req.PostFormValue("called_number"), req.PostFormValue("caller_number")
	}
	if req.Method == "GET"{
		called_number,caller_number = req.FormValue("called_number"),req.FormValue("caller_number")
	}
	sigparams = req.FormValue("SigParameter")
	authorization = req.Header.Get("Authorization")

	if user, ret = config.DecodeSigParams(sigparams, authorization); ret == false {
		//检验未通过
		seelog.Errorf("user authentication failed,appname:outboundcall,cause:user is not exist or sigparams is incorrect")
		rw.Header().Set("content-type", "text/plain")
		rw.WriteHeader(http.StatusForbidden)
		rw.Write([]byte("invalid user"))
		return
	}
	seelog.Infof("parseform outboundcall data is {user:%s, caller_number: %s, called_number: %s, sigs: %s}", user,caller_number,called_number, sigparams)

}

func initserver(url string) {
	seelog.Infof("begin fsappsvr,listen url is: %s",url)
	http.HandleFunc("/insertrecording", insertrecording)
	http.HandleFunc("/queryrecording", queryrecording)
	http.HandleFunc("/doublecall/", doublecall)
	http.HandleFunc("/voiceidentcall/",voiceidentcall)
	http.HandleFunc("/outboundcall/",outboundcall)
	fmt.Println(http.ListenAndServe(url, nil))
}


func InitLogger(){
	ILogger,err := seelog.LoggerFromConfigAsFile("./seelog.xml")

	if err !=nil{
		seelog.Critical("err parse config log file",err)
		log.Fatal("err parse config log file")
		return
	}
	seelog.ReplaceLogger(ILogger)
}

func main() {
	InitLogger()
	config = NewConfigEx("./config.json")
	err := dbhandler.NewDB(config.GetDBConnectString())
	if err != nil {
		seelog.Errorf("connect to fs database error: %s",err.Error())
		os.Exit(0)
	}
	fsclient, err = ConnectFs()
	if err == nil {
		fsclient.Connection.Send("events json ALL")
		go fsclient.PopAppRequest()
		go fsclient.ReadMessage()
	}
	initserver(config.GetListenUrl())
}
