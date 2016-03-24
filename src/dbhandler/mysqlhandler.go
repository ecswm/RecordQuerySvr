package dbhandler

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"time"
)

type DBObj struct {
	db *sql.DB
}

var DbObj *DBObj

func NewDB(connectstring string) error {
	DbObj = new(DBObj)
	db, err := sql.Open("mysql", connectstring)
	if err != nil {
		fmt.Println("database initialize %s, connect_string is error : ", connectstring, err.Error())
		db.Close()
		return err
	}
	err = db.Ping()
	if err != nil {
		fmt.Println("database ping is error : ", err.Error())
		db.Close()
		log.Fatal(err)
		return err
	}
	fmt.Println("connect database success")
	DbObj.db = db
	return nil
}

func (this *DBObj) CloseDB() error {
	err := this.db.Close()
	return err
}

func (this *DBObj) QueryRecordingPath(callid string) (string, error) {
	var recordingpath string = ""
	fmt.Println("query recordinginfo, call_id is ", callid)
	err := this.db.QueryRow(qryrecscript, callid).Scan(&recordingpath)
	if err != nil {
		fmt.Println("query recordinginfo error : ", err.Error())
	}
	return recordingpath, err
}

func (this *DBObj) InsertRecordingInfo(recordinginfo RecordingInfo) (bool, error) {
	ret := false
	stmtIns, err := this.db.Prepare(createrecscript)
	defer stmtIns.Close()

	if err == nil {
		_, err = stmtIns.Exec(recordinginfo.Callid,
			recordinginfo.RecordingPath,
			recordinginfo.RecordingTime,
			recordinginfo.CallerNumber,
			recordinginfo.CalledNumber)
		if err != nil {
			fmt.Println("inert recordinginfo error : ", err.Error())
			ret = false
		} else {
			ret = true
		}
	}
	return ret, err
}

//创建一个job信息
func (this *DBObj) CreateJobInfo(jobid string,customer string,appname string)(bool,error){
	ret:=true
	stmtIns,err:= this.db.Prepare(createjobscript)
	defer stmtIns.Close()

	if err == nil{
		_,err = stmtIns.Exec(jobid,
				customer,
				appname,
				time.Now().Format("2006-01-02 15:04:05"),
		)
		if err != nil{
			fmt.Println("create jobinfo error : ",err.Error())
			ret = false
		}
	}
	return ret,err
}

//更新一个job信息
func (this *DBObj) UpdateJobInfo (jobid string,callid string,result string)(bool,error){
	ret:= true
	stmtIns,err:= this.db.Prepare(updatejobscript)
	defer stmtIns.Close()

	if err == nil{
		_,err = stmtIns.Exec(callid,result,jobid)
		if err != nil{
			fmt.Println("update jobinfo error :",err.Error())
			ret = false
		}
	}
	return ret,err

}

func (this *DBObj) DelRecordingInfo(callid string) {
	stmt, err := this.db.Prepare(delrecscript)
	defer stmt.Close()

	if err != nil {
		_, err = stmt.Exec(callid)
		if err != nil {
			fmt.Println("delete reordinginfo error : ", err.Error())
		}
	}
}
