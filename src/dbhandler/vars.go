package dbhandler


const createjobscript  = "INSERT INTO job_record(job_id,customer_name,app_name,create_time) VALUES(?,?,?,?)"
const updatejobscript  = "UPDATE job_record SET call_id = ?,job_result = ? WHERE job_id = ?"
const createrecscript = "INSERT INTO call_record(call_id,record_path,record_time,caller_number,called_number) VALUES(?,?,?,?,?)"
const qryrecscript = "SELECT record_path FROM call_record WHERE call_id = ?"
const delrecscript = "DELETE FROM call_record WHERE call_id =?"

