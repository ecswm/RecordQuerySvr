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
	"dbhandler"
)

var (
	buffsize        = 50
	bgevtname = "BACKGROUND_JOB"
)

/*
语音验证码APP
包含呼叫方号码,验证码
*/
type VoiceIdentCallApp struct {
	CalledNumber string
	IdentCode string
	CallApp
}

/*
双呼APP
包含呼叫双方号码,显示的主叫信息,JobId
*/
type DoubleCallApp struct {
	CallerNumber string
	CalledNumber string
	CallApp
}

type CallApp struct {
	Job_Uuid chan string
	Err      chan error
	Response chan *CallAppRsp
	Ani           string
	CustomerName  string
	PushResultUrl string
	AppName       string
}
/*
双呼响应
包含错误代码,消息，创建时间，唯一标识
*/
type CallAppRsp struct {
	ErrCode   uint   `json:"errcode"`
	Message   string `json:"msg"`
	Time      string `json:"time"`
	Call_Uuid string `json:"callid"`
}

func (callrsp *CallAppRsp) GenerateRspJson() ([]byte, error) {
	rsp, err := json.Marshal(callrsp)
	return rsp, err
}

/*
FS Client
包含一个ESL Client和一个FsApp Channel
*/
type FsClient struct {
	*eventsocket.Connection
	Apps chan  interface{}
	Msg  map[string]interface{}
}

/*
创建一个FS连接
*/
func ConnectFs() (*FsClient, error) {
	c, err := eventsocket.Dial(config.GetEslUrl(), config.GetEslPwd())
	if err != nil {
		log.Fatal(err)
	}
	fsclient := FsClient{
		Apps: make(chan interface{}, buffsize),
		Msg:  make(map[string]interface{})}
	fsclient.Connection = c
	return &fsclient, err
}

/*
将一个App请求推送至
*/
func (c *FsClient) PushAppRequest(app interface{}) {
	c.Apps <- app
}

func (c *FsClient) PopAppRequest() {
	for {
		var app interface{}
		var(
			jobid string
			err error
		)

		select {
		case app = <-c.Apps:
			switch app.(type) {
			case *DoubleCallApp:
				jobid, err = c.SendDoubleCall(app)
				dbhandler.DbObj.CreateJobInfo(jobid, app.(*DoubleCallApp).CustomerName, app.(*DoubleCallApp).AppName)
				if err != nil {
					app.(*DoubleCallApp).Err <- err
					continue
				}
			case *VoiceIdentCallApp:
				jobid,err = c.SendVoiceIdentCall(app)
				dbhandler.DbObj.CreateJobInfo(jobid,app.(*VoiceIdentCallApp).CustomerName,app.(*VoiceIdentCallApp).AppName)
				if err != nil {
					app.(*VoiceIdentCallApp).Err <- err
					continue
				}
			}
			c.Msg[jobid] = app
		}
	}
}

func DoubleCallCmd(caller_number string,called_number string,ani string) string{
	return fmt.Sprintf("bgapi originate {ignore_early_media=true,origination_caller_id_number='%s'}sofia/external/'%s'@10.0.0.61 '&lua('/usr/local/freeswitch/scripts/LXHealthCare/main.lua' %s %s)'", ani, caller_number, caller_number, called_number)
}

func VoiceIdentCallCmd(called_number string,ident_code string) string{
	return fmt.Sprintf("bgapi originate {ignore_early_media=true,origination_caller_id_number='%s'}sofia/external/'%s@10.0.0.61 '&lua('/usr/local/freeswitch/scripts/GreenTown/VoiceIdentCall.lua' %s %s)'",called_number,called_number,ident_code)

}

func (c *FsClient) SendDoubleCall(app interface{}) (jobid string, err error) {
	cmd:=DoubleCallCmd(app.(*DoubleCallApp).CallerNumber,app.(*DoubleCallApp).CalledNumber,app.(*DoubleCallApp).Ani)
	event, err := c.Connection.Send(cmd)
	event.PrettyPrint()
	return event.Get("Job-Uuid"), err
}

func (c *FsClient) SendVoiceIdentCall(app interface{})(jobid string,err error){
	cmd:=VoiceIdentCallCmd(app.(*VoiceIdentCallApp).CalledNumber,app.(*VoiceIdentCallApp).IdentCode)
	event,err:= c.Connection.Send(cmd)
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
			if _,ok :=c.Msg[jobid];!ok{
				fmt.Println("the event is not current service scope\n")
				ev.PrettyPrint()
				return
			}
			//获取body信息
			ret := strings.Split(ev.Body, " ")
			rsp := CallAppRsp{ErrCode: 0, Message: "OK", Time: time.Now().Format("2006-01-02 15:04:05"), Call_Uuid: strings.TrimRight(ret[1], "\n")}

			if ret[0] != "+OK" {
				rsp.ErrCode = 503
				rsp.Message = strings.TrimRight(ret[1], "\n")
				rsp.Call_Uuid = ""
			}
			dbhandler.DbObj.UpdateJobInfo(jobid,rsp.Call_Uuid,rsp.Message)

			switch c.Msg[jobid].(type) {
			case *DoubleCallApp:
				c.Msg[jobid].(*DoubleCallApp).Response <- &rsp
			case *VoiceIdentCallApp:
				c.Msg[jobid].(*VoiceIdentCallApp).Response <- &rsp
			}
			fmt.Println("\nNew event")
			ev.PrettyPrint()
		}
	}
}

