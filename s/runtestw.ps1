#!/usr/local/bin/pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

go build -o testw.exe github.com/kjk/winapigen/testw
exitIfFailed
.\testw.exe
Remove-Item .\testw.exe
