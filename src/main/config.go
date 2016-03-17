package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	_ "strconv"
	"strings"
)

type Configuration interface {
	GetUserAni(string) string
	GetListenUrl() string
	GetEslUrl() string
	GetEslPwd() string
	DecodeSigParams(sigparams string, authorization string) (string, bool)
}

/*
主叫,名字,密钥
*/
type userinfo struct {
	Ani  string
	Name string
	Key  string
}

/*

{"BindIp":"192.168.1.103",
"BindPort":9090,
"FsHost":"10.0.0.35",
"FsPort":8021,
"Password":"ClueCon",
"Timeout":10,
"Users":[
{"Ani":"950598","Name":"GreenTown","Key":"0B13F3A3-534F-4389-A258-1B32F99750F0"},
{"Ani":"950598","Name":"lanxihealthcare","Key":"9BF1758C-4FBB-4089-92BE-0AA70C503EB4"}],
"SupportApps":["DoubleCall","VoiceIndentCall"]
}
*/
/*
绑定的IP,绑定的端口,FS的IP,端口,支持的APPS种类,用户列表
*/
type configuration struct {
	BindIp      string
	BindPort    uint
	FsHost      string
	FsPort      uint
	Password    string
	Timeout     uint
	SupportApps []string
	Users       []userinfo
}

func (this configuration) GetUserAni(name string) string {
	for _, user := range this.Users {
		if user.Name == name {
			return user.Ani
		}
	}
	return "95059"
}

func (this configuration) GetListenUrl() string {
	return fmt.Sprintf("%s:%d", this.BindIp, this.BindPort)
}

func (this configuration) GetEslUrl() string {
	return fmt.Sprintf("%s:%d", this.FsHost, this.FsPort)
}

func (this configuration) GetEslPwd() string {
	return this.Password
}

/*
从key.ini中获取user/key信息
*/
func NewConfigEx(path string) Configuration {
	config := new(configuration)
	file, err := os.Open(path)
	defer file.Close()
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)

	if err != nil {
		return nil
	}
	return config
}

/*
解析sigparams参数
*/

func (this configuration) DecodeSigParams(sigparams string, authorization string) (string, bool) {
	ret := false
	decoded, err := base64.StdEncoding.DecodeString(authorization)
	if err != nil {
		fmt.Println("decode error:", err)
		return "", ret
	}
	outarray := strings.Split(string(decoded), ":")
	if len(outarray) > 0 {
		for _, user := range this.Users {
			if user.Name == outarray[0] {
				md5Ctx := md5.New()
				md5Ctx.Write([]byte(user.Name + user.Key + outarray[1]))
				md5hash := md5Ctx.Sum(nil)
				if strings.ToUpper(hex.EncodeToString(md5hash)) == sigparams {
					ret = true
				}
			}
		}
	}
	return outarray[0], ret
}
