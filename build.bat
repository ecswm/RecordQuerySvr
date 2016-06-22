
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
::set %GOPATH% = "G:\work\GOLang\src\RecordQuerySvr"
::set GOPATH = %cd%
go build main
main.exe
@pause
echo 'finished'