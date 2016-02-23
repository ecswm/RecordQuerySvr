package dbhandler

import (
	"encoding/json"
)

type RecordInfo struct {
	Uuid          string `json:"uuid"`
	Record_path   string `json:"record_path"`
	Record_time   string `json:"record_time"`
	Caller_number string `json:"caller_number"`
	Called_number string `json:"called_number"`
}

type ResponseInfo struct {
	Recordinfo *RecordInfo `json:recordinfo`
	Error      string      `json:err,omitempty`
	Code       string      `json:code,omitempty`
}

/*
func GenerateRspJson(call_id string,
	record_path string,
	record_time string,
	caller_number string,
	called_number string,
	errmsg string,
	code string) []byte {
	recordinfo := RecordInfo{call_id, record_path, record_time, caller_number, called_number}
	responseinfo := ResponseInfo{recordinfo, errmsg, code}

	rsp, _ := json.Marshal(responseinfo)
	return rsp
}
*/
func GenerateRspJson(recordinfo *RecordInfo, errmsg string, code string) []byte {
	responseinfo := ResponseInfo{recordinfo, errmsg, code}

	rsp, _ := json.Marshal(responseinfo)
	return rsp
}
