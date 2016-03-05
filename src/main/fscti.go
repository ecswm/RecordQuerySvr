package main

import (
	"fmt"
	"goesl"
	"net/http"
	"strings"
	"time"
)

var (
	fshost        = "10.0.0.35"
	fsport   uint = 8021
	password      = "ClueCon"
	timeout       = 10
)

/*
FS APP请求结构体
AppName:对应的功能名
JobId:对应的调用标识ID
Writer:调用者对应的http responsewriter
*/
type FsApp struct {
	AppName string
	JobId   string
	Writer  http.ResponseWriter
}

/*
FS CTI结构体
包含一个ESL Client和一个FsApp Channel
*/
type FsCti struct {
	goesl.Client
	Apps chan *FsApp
}

/*
创建一个fscti
*/
func ConnectFs() (*FsCti, error) {
	client, err := goesl.NewClient(fshost, fsport, password, timeout)
	if err != nil {
		fmt.Errorf("Error while creating new client: %s", err)
		return nil, err
	}
	fscti := FsCti{Apps: make(chan *FsApp)}
	fscti.Client = client
	return &fscti, nil
}

/*
从ESL Client中读取消息
*/
func (fscti *FsCti) ReadMessage2() {
	for {
		msg, err := fscti.ReadMessage()

		if err != nil {

			// If it contains EOF, we really dont care...
			if !strings.Contains(err.Error(), "EOF") && err.Error() != "unexpected end of JSON input" {
				fmt.Println("Error while reading Freeswitch message: %s", err)
			}
			fmt.Println("%s", msg)
			break
		}
		//解析消息，根据JobID匹配发起者
		fmt.Println("%s", msg)
	}
}

/*
发送一个DoubleCall请求
*/
func (fscti *FsCti) SendDoubleCall(caller_number string, called_number string) error {
	req := fmt.Sprintf("originate %s %s \r\n Job-UUID:%s", "{ignore_early_media=true,origination_caller_id_number='950598'}user/'80001'@10.0.0.35",
		"&bridge([origination_caller_id_number='950598']sofia/external/'910000'@10.0.0.61)", "my-job-id"+time.Now().Format("2006-01-02 15:04:05"))
	fscti.BgApi(req)
	return nil
}
