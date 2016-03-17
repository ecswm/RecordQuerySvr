package main

import (
	_ "bytes"
	"encoding/json"
	"eventsocket"
	"fmt"
	"log"
	_ "net/http"
	"strings"
	"time"
)

var (
	fshost          = "10.0.0.35"
	fsport   uint64 = 8021
	password        = "ClueCon"
	timeout         = 10
	buffsize        = 50
)

var (
	bgevtname = "BACKGROUND_JOB"
)

/*
双呼请求
*/
type DoubleCallAppReq struct {
	CallerNumber  string
	CalledNumber  string
	Ani           string
	PushResultUrl string
}

/*
双呼APP
包含呼叫双方号码,显示的主叫信息,JobId
*/
type DoubleCallApp struct {
	Request  *DoubleCallAppReq
	Job_Uuid chan string
	Err      chan error
	Response chan *DoubleCallAppRsp
}

/*
双呼响应
包含错误代码,消息，创建时间，唯一标识
*/
type DoubleCallAppRsp struct {
	ErrCode   uint   `json:"errcode"`
	Message   string `json:"msg"`
	Time      string `json:"time"`
	Call_Uuid string `json:"callid"`
}

func (doublecallrsp *DoubleCallAppRsp) GenerateRspJson() ([]byte, error) {
	rsp, err := json.Marshal(doublecallrsp)
	return rsp, err
}

/*
FS Client
包含一个ESL Client和一个FsApp Channel
*/
type FsClient struct {
	*eventsocket.Connection
	Apps chan *DoubleCallApp
	Msg  map[string]*DoubleCallApp
}

/*
创建一个FS连接
*/
func ConnectFs2() (*FsClient, error) {
	c, err := eventsocket.Dial(config.GetEslUrl(), config.GetEslPwd())
	if err != nil {
		log.Fatal(err)
	}
	fsclient := FsClient{
		Apps: make(chan *DoubleCallApp, buffsize),
		Msg:  make(map[string]*DoubleCallApp)}
	fsclient.Connection = c
	return &fsclient, err
}

/*
将一个App请求推送至
*/
func (c *FsClient) PushAppRequest(app *DoubleCallApp) {
	c.Apps <- app
}

func (c *FsClient) PopAppRequest() {
	for {
		var app *DoubleCallApp
		select {
		case app = <-c.Apps:
			jobid, err := c.SendDoubleCall(app)
			if err != nil {
				app.Err <- err
				continue
			}
			c.Msg[jobid] = app
		}
	}
}

func (c *FsClient) SendDoubleCall(app *DoubleCallApp) (jobid string, err error) {
	fmt.Println("begin doublecall....")
	req := fmt.Sprintf("bgapi originate {ignore_early_media=true,origination_caller_id_number='%s'}sofia/external/'%s'@10.0.0.61 '&lua('/usr/local/freeswitch/scripts/LXHealthCare/main.lua' %s %s)'", app.Request.Ani, app.Request.CallerNumber, app.Request.CallerNumber, app.Request.CalledNumber)
	fmt.Println(req)
	event, err := c.Connection.Send(req)
	event.PrettyPrint()
	return event.Get("Job-Uuid"), err
}

func (c *FsClient) ReadMessage() {
	for {
		ev, err := c.Connection.ReadEvent()
		if err != nil {
			log.Fatal(err)
		}
		if bgevtname == ev.Get("Event-Name") {
			jobid := ev.Get("Job-Uuid")
			//获取body信息
			ret := strings.Split(ev.Body, " ")
			rsp := DoubleCallAppRsp{ErrCode: 0, Message: "", Time: time.Now().Format("2006-01-02 15:04:05"), Call_Uuid: strings.TrimRight(ret[1], "\n")}

			if ret[0] != "+OK" {
				rsp.ErrCode = 503
				rsp.Message = ret[1]
				rsp.Call_Uuid = ""
			}
			c.Msg[jobid].Response <- &rsp
			fmt.Println("\nNew event")
			ev.PrettyPrint()
		}
	}
}

func (c *FsClient) SendDoubleCall2() error {
	req := fmt.Sprintf("bgapi originate %s %s", "{ignore_early_media=true,origination_caller_id_number='950598'}user/'80001'@10.0.0.35",
		"&bridge([origination_caller_id_number='950598']sofia/external/'910000'@10.0.0.61)")
	event, err := c.Connection.Send(req)
	event.PrettyPrint()
	return err
}
