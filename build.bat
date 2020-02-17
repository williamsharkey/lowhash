:: file: build_all.bat

@echo off
setlocal ENABLEEXTENSIONS ENABLEDELAYEDEXPANSION


echo.
echo ---^> building Go for linux 386
echo.
set GOOS=linux
set GOARCH=386
go build -o build/linux32/lowhash

echo.
echo ---^> building Go for linux amd64
echo.
set GOOS=linux
set GOARCH=amd64
go build -o build/linux64/lowhash


echo.
echo ---^> building Go for macOS 386
echo.
set GOOS=darwin
set GOARCH=386
go build -o build/osx32/lowhash

echo.
echo ---^> building Go for macOS amd64
echo.
set GOOS=darwin
set GOARCH=amd64
go build -o build/osx64/lowhash


echo.
echo ---^> building Go for windows 386
echo.
set GOOS=windows
set GOARCH=386
go build -o build/win32/lowhash.exe

echo.
echo ---^> building Go for windows amd64
echo.
set GOOS=windows
set GOARCH=amd64
go build -o build/win64/lowhash.exe

