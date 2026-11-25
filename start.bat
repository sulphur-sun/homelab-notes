@echo off
cd /d "%~dp0\api"
start "" "api.exe"

cd /d "%~dp0\db\target\debug"
start "" "db.exe"
