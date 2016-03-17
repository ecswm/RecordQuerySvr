package main

import (
	"dbhandler"
	"fmt"
	"io"
	"net/http"
	"os"
	_ "strings"
	"time"
)

var fsclient *FsClient = nil
var config Configuration

func insertrecording(rw http.ResponseWriter, req *http.Request) {
	fmt.Println("insert recordinginfo request, the request info is ", req.URL.Path)
	err := req.ParseForm()
	if err != nil {
		fmt.Println(fmt.Sprintf("parseform error is : %s", err.Error()))
		return
	}
	callid := req.FormValue("callid")
	callernumber := req.FormValue("callernumber")
	callednumber := req.FormValue("callednumber")
	recordingpath := req.FormValue("recordingpath")

	fmt.Println(fmt.Sprintf("parseform insertrecordinginfo request, data is %s_%s_%s_%s", callid, callernumber, callednumber, recordingpath))
	recordinginfo := dbhandler.RecordingInfo{callid,
		recordingpath,
		time.Now().Format("2006-01-02 15:04:05"),
		callernumber,
		callednumber}

	ret, err := dbhandler.DbObj.InsertRecordingInfo(recordinginfo)
	rsp := dbhandler.GenerateRspJson(&recordinginfo, "", "0")
	if ret != true {
		fmt.Println("insert new record error is ", err.Error())
		rsp = dbhandler.GenerateRspJson(&recordinginfo, err.Error(), "0")
	}
	rw.Header().Set("content-type", "application/json")
	rw.Write(rsp)
}

func queryrecording(rw http.ResponseWriter, req *http.Request) {
	fmt.Println("queryrecording request, the request info is ", req.URL.Path)
	err := req.ParseForm()
	if err != nil {
		fmt.Println("parseform error is : ", err.Error())
		return
	}
	callid := req.FormValue("call_id")
	fmt.Println(fmt.Sprintf("parseform queryrecord data is %s", callid))

	recordingpath, err := dbhandler.DbObj.QueryRecordingPath(callid)
	rw.Header().Set("content-type", "application/json")
	if err != nil {
		fmt.Println("queryrecording error is ", err.Error())
		rsp := dbhandler.GenerateRspJson(nil, err.Error(), "")
		rw.Write(rsp)
		return
	}
	f, err := os.Open(recordingpath)
	if err != nil {
		rsp := dbhandler.GenerateRspJson(nil, err.Error(), "")
		rw.Write(rsp)
		return
	}
	defer f.Close()
	rw.Header().Set("content-type", "audio/wav")
	bytes, err := io.Copy(rw, f)
	if err != nil {
		fmt.Println("Error: copy file to httpclient responsefailed:", err)
	} else {
		fmt.Printf("send %v to %v success with %v bytes\n", req.URL.Path, req.RemoteAddr, bytes)
	}
}

func doublecall(rw http.ResponseWriter, req *http.Request) {
	fmt.Println("doublecall request, the request info is ", req.URL.Path)
	err := req.ParseForm()

	if err != nil {
		fmt.Println("parseform error is : ", err.Error())
		return
	}

	var (
		caller_number string
		called_number string
		sigparams     string
		authorization string
	)

	if req.Method == "POST" {
		caller_number, called_number = req.PostFormValue("Caller_number"), req.PostFormValue("Called_number")
		sigparams = req.FormValue("SigParameter")
		authorization = req.Header.Get("Authorization")
	}

	fmt.Println(fmt.Sprintf("parseform doublecall data is {caller_number: %s, called_number: %s, sigs: %s}", caller_number, called_number, sigparams))

	var (
		name string
		ret  bool
	)
	if name, ret = config.DecodeSigParams(sigparams, authorization); ret == false {
		//检验未通过
		rw.Header().Set("content-type", "text/plain")
		rw.WriteHeader(http.StatusForbidden)
		rw.Write([]byte("invalid user"))
		return
	}

	app := DoubleCallApp{
		Request:  &DoubleCallAppReq{Ani: config.GetUserAni(name), CallerNumber: caller_number, CalledNumber: called_number},
		Response: make(chan *DoubleCallAppRsp),
		Job_Uuid: make(chan string),
		Err:      make(chan error)}

	fsclient.PushAppRequest(&app)

	rw.Header().Set("content-type", "application/json")
	var (
		doublecallrsp *DoubleCallAppRsp
		job_uuid      string
	)
	select {
	case err = <-app.Err:
		rw.Write([]byte(err.Error()))
	case job_uuid = <-app.Job_Uuid:
		rw.Write([]byte(job_uuid))
	case <-time.After(time.Second * 20):
		rw.Write([]byte("RequestTimeout"))
	case doublecallrsp = <-app.Response:
		if retjson, err := doublecallrsp.GenerateRspJson(); err == nil {
			rw.Write([]byte(retjson))
		}
	}
}

func initserver(url string) {
	fmt.Println("begin queryrecordsvr,listen url is", url)
	err := dbhandler.NewDB("10.0.0.33", 3306, "root", "lvcheng2015~", "fs")
	if err != nil {
		fmt.Println("connect to fs database error: ", err.Error())
	}
	http.HandleFunc("/insertrecording", insertrecording)
	http.HandleFunc("/queryrecording", queryrecording)
	http.HandleFunc("/doublecall/", doublecall)
	fmt.Println(http.ListenAndServe(url, nil))
}

func main() {
	var err error
	config = NewConfigEx("./config.json")
	fsclient, err = ConnectFs2()
	if err == nil {
		fsclient.Connection.Send("events json ALL")
		go fsclient.PopAppRequest()
		go fsclient.ReadMessage()
	}
	initserver(config.GetListenUrl())
}
