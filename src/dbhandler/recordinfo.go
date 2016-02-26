package dbhandler

import (
	"encoding/json"
)

type RecordingInfo struct {
	Callid        string `json:"uuid"`
	RecordingPath string `json:"record_path"`
	RecordingTime string `json:"record_time"`
	CallerNumber  string `json:"caller_number"`
	CalledNumber  string `json:"called_number"`
}

type Response struct {
	Recordinginfo *RecordingInfo `json:recordinfo`
	Error         string         `json:err,omitempty`
	Code          string         `json:code,omitempty`
}

func GenerateRspJson(recordinginfo *RecordingInfo, errmsg string, code string) []byte {
	response := Response{recordinginfo, errmsg, code}

	rsp, _ := json.Marshal(response)
	return rsp
}
