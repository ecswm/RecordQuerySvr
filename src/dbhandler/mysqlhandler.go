package dbhandler

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

type DBObj struct {
	db *sql.DB
}

var DbObj *DBObj

func NewDB(ip string, port uint, user string, password string, dbname string) error {
	connectstring := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s", user, password, ip, port, dbname)
	dbobj := new(DBObj)
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
	}
	dbobj.db = db
	return nil
}

func (this *DBObj) CloseDB() error {
	err := this.db.Close()
	return err
}

func (this *DBObj) QueryRecordingPath(callid string) (string, error) {
	var recordingpath string = ""
	fmt.Println("query recordinginfo, call_id is ", callid)
	err := this.db.QueryRow("SELECT record_path FROM call_record WHERE call_id = ?", callid).Scan(&recordingpath)
	if err != nil {
		fmt.Println("query recordinginfo error : ", err.Error())
	}
	return recordingpath, err
}

func (this *DBObj) InsertRecordingInfo(recordinginfo RecordingInfo) (bool, error) {
	ret := false
	stmtIns, err := this.db.Prepare("INSERT INTO call_record(call_id,record_path,record_time,caller_number,called_number) VALUES(?,?,?,?,?)")
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

func (this *DBObj) DelRecordingInfo(callid string) {
	stmt, err := this.db.Prepare("DELETE FROM call_record WHERE call_id =?")
	defer stmt.Close()

	if err != nil {
		_, err = stmt.Exec(callid)
		if err != nil {
			fmt.Println("delete reordinginfo error : ", err.Error())
		}
	}
}
