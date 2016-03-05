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
	caller_number, called_number := req.FormValue("caller_number"), req.FormValue("called_number")
	fmt.Println(fmt.Sprintf("parseform doublecall data is %s_%s", caller_number, called_number))

	//调用ESL接口

}

func initserver(ip string, port string) {
	fmt.Println("begin queryrecordsvr,ip is", ip, "port is", port)
	err := dbhandler.NewDB("10.0.0.33", 3306, "root", "lvcheng2015~", "fs")
	if err != nil {
		fmt.Println("connect to fs database error: ", err.Error())
	}
	http.HandleFunc("/insertrecording", insertrecording)
	http.HandleFunc("/queryrecording", queryrecording)
	http.HandleFunc("/doublecall", doublecall)
	fmt.Println(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func main() {
	//initserver("10.0.0.35", "8083")

	cti, _ := ConnectFs()
	go cti.Client.Handle()

	cti.Client.Send("events json ALL")

	cti.SendDoubleCall("", "")

	cti.ReadMessage2()

	//fmt.Println("test")
}
