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
	"github.com/satori/go.uuid"
	"github.com/cihub/seelog"
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
	CallId string
	Bridge_CallId string
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
	Prefix        string
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
	CallIdMapped map[string]string
}

/*
创建一个FS连接
*/
func ConnectFs() (*FsClient, error) {
	c, err := eventsocket.Dial(config.GetEslUrl(), config.GetEslPwd())
	if err != nil {
		seelog.Errorf("connect freeswitch occur err,err is:[]",err.Error())
		log.Fatal(err)
	}
	seelog.Infof("connect freeswitch success,addr is %s",config.GetEslUrl())
	fsclient := FsClient{
		Apps: make(chan interface{}, buffsize),
		Msg:  make(map[string]interface{}),
		CallIdMapped:make(map[string]string)}
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
		var (
			app interface{}
			err error
		)
		select {
		case app = <-c.Apps:
			jobid := fmt.Sprintf("%s",uuid.NewV4())
			c.Msg[jobid] = app
			switch app.(type) {
			case *DoubleCallApp:
				err = c.SendDoubleCall(app,jobid)
				dbhandler.DbObj.CreateJobInfo(jobid,app.(*DoubleCallApp).CustomerName, app.(*DoubleCallApp).AppName)
				if err != nil {
					seelog.Errorf("senddoublecall occurs err,err is: []",err.Error())
					app.(*DoubleCallApp).Err <- err
				}
			case *VoiceIdentCallApp:
				err = c.SendVoiceIdentCall(app,jobid)
				dbhandler.DbObj.CreateJobInfo(jobid, app.(*VoiceIdentCallApp).CustomerName, app.(*VoiceIdentCallApp).AppName)
				if err != nil {
					seelog.Errorf("sendvoiceidentcall occurs err,err is: []",err.Error())
					app.(*VoiceIdentCallApp).Err <- err
				}

			}
		}
	}
}

func DoubleCallCmd(jobid string,user string,caller_number string,called_number string,ani string) string{
	 return fmt.Sprintf("bgapi originate {ignore_early_media=true,origination_caller_id_number='%s'}sofia/external/'%s'@%s '&lua('/usr/local/freeswitch/scripts/%s/main.lua' %s %s)'\nJob-UUID:%s\n\n",ani, caller_number, config.GetUserGateWay(user), user,caller_number,called_number,jobid)
}

func VoiceIdentCallCmd(jobid string,user string,called_number string,ident_code string,ani string) string{
	return fmt.Sprintf("bgapi originate {ignore_early_media=true,origination_caller_id_number='%s'}sofia/external/'%s'@%s '&lua('/usr/local/freeswitch/scripts/%s/VoiceIdentCall.lua' %s %s)'\nJob-UUID:%s\n\n",ani,called_number,config.GetUserGateWay(user),user,called_number,ident_code,jobid)
}

func DoubleCallTestCmd(jobid string,user string,caller_number string,called_number string,ani string) string{
	return fmt.Sprintf("bgapi originate {ignore_early_media=true,origination_caller_id_number='%s'}user/'%s'@192.168.1.224 '&lua('/usr/local/freeswitch/scripts/%s/DoubleCall.lua' %s %s)'\nJob-UUID:%s\n\n",ani, caller_number,user, caller_number, called_number,jobid)
}

func VoiceIdentCallTestCmd(jobid string,user string,called_number string,ident_code string,ani string) string  {
	return fmt.Sprintf("bgapi originate {ignore_early_media=true,origination_caller_id_number='%s'}user/'%s'@192.168.1.224 '&lua('/usr/local/freeswitch/scripts/%s/VoiceIdentCall.lua' %s %s)'\nJob-UUID:%s\n\n",ani,called_number,user,called_number,ident_code,jobid)
}

func (c *FsClient) SendDoubleCall(app interface{},jobid string) (err error) {
	var cmd string
	if(app.(*DoubleCallApp).CustomerName == "TEST"){
		cmd=DoubleCallCmd(
			jobid,
			app.(*DoubleCallApp).CustomerName,
			app.(*DoubleCallApp).Prefix + app.(*DoubleCallApp).CallerNumber,
			app.(*DoubleCallApp).Prefix + app.(*DoubleCallApp).CalledNumber,
			app.(*DoubleCallApp).Ani)
	}else{
		cmd=DoubleCallCmd(
			jobid,
			app.(*DoubleCallApp).CustomerName,
			app.(*DoubleCallApp).Prefix + app.(*DoubleCallApp).CallerNumber,
			app.(*DoubleCallApp).Prefix + app.(*DoubleCallApp).CalledNumber,
			app.(*DoubleCallApp).Ani)
	}
	event, err := c.Connection.Send(cmd)
	if err!=nil{
		return err
	}
	event.PrettyPrint()
	return nil
}

func (c *FsClient) SendVoiceIdentCall(app interface{},jobid string)(err error){
	var cmd string
	if(app.(*VoiceIdentCallApp).CustomerName == "TEST") {
		cmd = VoiceIdentCallCmd(
			jobid,
			app.(*VoiceIdentCallApp).CustomerName,
			app.(*VoiceIdentCallApp).Prefix + app.(*VoiceIdentCallApp).CalledNumber,
			app.(*VoiceIdentCallApp).IdentCode,
			app.(*VoiceIdentCallApp).Ani)
	}else{
		cmd = VoiceIdentCallCmd(
			jobid,
			app.(*VoiceIdentCallApp).CustomerName,
			app.(*VoiceIdentCallApp).Prefix + app.(*VoiceIdentCallApp).CalledNumber,
			app.(*VoiceIdentCallApp).IdentCode,
			app.(*VoiceIdentCallApp).Ani)
	}
	event,err:= c.Connection.Send(cmd)
	if err!=nil{
		return err
	}
	event.PrettyPrint()
	return nil
}

func (c *FsClient) GetUuid() (jobid string,err error)  {
	event,err := c.Connection.Send("create_uuid")
	event.PrettyPrint()
	return event.Get("Job-Uuid"),err
}

func (c *FsClient) ReadMessage() {
	for {
		ev, err := c.Connection.ReadEvent()
		if err != nil {
			log.Fatal(err)
		}
		evtname := ev.Get("Event-Name")
		switch evtname {
		case "BACKGROUND_JOB":
			jobid := ev.Get("Job-Uuid")
			if _,ok :=c.Msg[jobid];!ok{
				seelog.Errorf("can not find jobid mapped app,jobid is: [%s]",jobid)
				ev.PrettyPrint()
				continue
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
			delete(c.Msg,jobid)
			c.CallIdMapped[rsp.Call_Uuid] = ""
			ev.PrettyPrint()
			break
		case "CHANNEL_BRIDGE":
			ev.PrettyPrint()
			callid := ev.Get("Bridge-A-Unique-Id")
			if _,ok :=c.CallIdMapped[callid];!ok{
				seelog.Errorf("the bridge-a-unique-id can not found in callidmap, the call is not current fsappsvr invoke, callid is: [%s]",callid)
				ev.PrettyPrint()
				continue
			}
			c.CallIdMapped[callid] = ev.Get("Bridge-B-Unique-Id")
			dbhandler.DbObj.UpdateCallInfo(callid,c.CallIdMapped[callid],"BRIDGEING")
			break
		case "CHANNEL_HANGUP_COMPLETE":
			ev.PrettyPrint()
			bridge_callid := ev.Get("Variable_bridge_uuid")
			if bridge_callid == ""{
				seelog.Infof("the hangup_complete event not contains bridge_callid,the call is not bridge call")
				ev.PrettyPrint()
				continue
			}
			if _,ok:= c.CallIdMapped[bridge_callid];!ok{
				seelog.Errorf("the bridge_callid can not found in callidmap,the callid is [%s]",bridge_callid)
				ev.PrettyPrint()
				continue
			}
			dbhandler.DbObj.UpdateCallInfo(bridge_callid,c.CallIdMapped[bridge_callid],ev.Get("Hangup-Cause"))
			delete(c.CallIdMapped,bridge_callid)
			break
		}
	}
}

