package dbhandler

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
)

func Open(_connectstring string) (*sql.DB, error) {
	db, err := sql.Open("mysql", _connectstring)
	if err != nil {
		fmt.Println("database initialize %s, connect_string is error : ", _connectstring, err.Error())
		db.Close()
		db = nil
	}
	err = db.Ping()
	if err != nil {
		fmt.Println("database ping is error : ", err.Error())
		db.Close()
		log.Fatal(err)
	}
	return db, err
}

func Query_record_path(db *sql.DB, call_id string) (string, error) {
	var record_path string = ""
	fmt.Println("query record, call_id is ", call_id)
	err := db.QueryRow("SELECT record_path FROM call_record WHERE call_id = ?", call_id).Scan(&record_path)
	if err != nil {
		fmt.Println("query record error : ", err.Error())
	}
	return record_path, err
}

func Insert_record(db *sql.DB, recordinfo RecordInfo) (bool, error) {
	var ret bool = false
	stmtIns, err := db.Prepare("INSERT INTO call_record(call_id,record_path,record_time,caller_number,called_number) VALUES(?,?,?,?,?)")
	defer stmtIns.Close()

	if err == nil {
		_, err = stmtIns.Exec(recordinfo.Uuid,
			recordinfo.Record_path,
			recordinfo.Record_time,
			recordinfo.Caller_number,
			recordinfo.Called_number)
		if err != nil {
			fmt.Println("inert into error : ", err.Error())
			ret = false
		} else {
			ret = true
		}

	}
	return ret, err
}

func Delete_record(db *sql.DB, call_id string) {
	stmt, err := db.Prepare("DELETE FROM call_record WHERE call_id =?")
	defer stmt.Close()

	if err != nil {
		_, err = stmt.Exec(call_id)
		if err != nil {
			fmt.Println("delete record error : ", err.Error())
		}
	}
}
