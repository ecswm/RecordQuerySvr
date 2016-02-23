package main

import (
	"database/sql"
	"dbhandler"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

var db *sql.DB = nil

func handler(rw http.ResponseWriter, req *http.Request) {

}

func newrecord(rw http.ResponseWriter, req *http.Request) {
	fmt.Println("insert record request, the request info is ", req.URL.Path)
	err := req.ParseForm()
	if err != nil {
		fmt.Println(fmt.Sprintf("parseform error is : %s", err.Error()))
		return
	}
	var call_id, caller_number, called_number, record_path string
	call_id = req.FormValue("call_id")
	caller_number = req.FormValue("caller_number")
	called_number = req.FormValue("called_number")
	record_path = req.FormValue("record_path")

	fmt.Println(fmt.Sprintf("parseform insertrecord data is %s_%s_%s_%s", call_id, called_number, caller_number, record_path))
	record_info := dbhandler.RecordInfo{call_id,
		record_path,
		time.Now().Format("2006-01-02 15:04:05"),
		caller_number,
		called_number}

	if db != nil {
		var rsp []byte
		ret, err := dbhandler.Insert_record(db, record_info)
		rsp = dbhandler.GenerateRspJson(&record_info, "", "0")
		if ret != true {
			fmt.Println("insert new record error is ", err.Error())
			rsp = dbhandler.GenerateRspJson(&record_info, err.Error(), "0")
		}
		rw.Header().Set("content-type", "application/json")
		rw.Write(rsp)
	}

}

func queryrecord(rw http.ResponseWriter, req *http.Request) {
	fmt.Println("query record request, the request info is ", req.URL.Path)
	err := req.ParseForm()
	if err != nil {
		fmt.Println("parseform error is : ", err.Error())
		return
	}

	var call_id, record_path string
	var f *os.File

	call_id = req.FormValue("call_id")
	fmt.Println(fmt.Sprintf("parseform queryrecord data is %s", call_id))
	if db != nil {
		record_path, err = dbhandler.Query_record_path(db, call_id)
		rw.Header().Set("content-type", "application/json")
		if err != nil {
			fmt.Println("query record error is ", err.Error())
			rsp := dbhandler.GenerateRspJson(nil, err.Error(), "")
			rw.Write(rsp)
			return
		}
		f, err = os.Open(record_path)
		if err != nil {
			rsp := dbhandler.GenerateRspJson(nil, err.Error(), "")
			rw.Write(rsp)
			return
		}
		defer f.Close()
		rw.Header().Set("content-type", "audio/wav")
		bytes, err := io.Copy(rw, f)
		if err != nil {
			fmt.Println("Error: copy file to http responsefailed:", err)
		} else {
			fmt.Printf("send %v to %v success with %v bytes\n", req.URL.Path, req.RemoteAddr, bytes)
		}
	}

}

func initserver(ip string, port string) {
	fmt.Println("begin queryrecordsvr,ip is", ip, "port is", port)
	var err error = nil
	db, err = dbhandler.Open("root:lvcheng2015~@tcp(10.0.0.33:3306)/fs")
	if err != nil {
		fmt.Println("connect to fs database error: ", err.Error())
	}
	http.HandleFunc("/new_record", newrecord)
	http.HandleFunc("/query_record", queryrecord)
	fmt.Println(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func main() {
	initserver("10.0.0.33", "8083")

}
