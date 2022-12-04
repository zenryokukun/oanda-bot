package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"
)

// time.FormatでYYYY-mm-ddTHH:MM:SS.000000000Z形式にするlayout
func layout() string {
	return "2006-01-02T15:04:05.000000000Z"
}

// layout -> unix
func toUnix(fmtTime string) int64 {
	t, err := time.Parse(layout(), fmtTime)
	if err != nil {
		fmt.Println(err)
	}
	return t.Unix()
}

// json形式のデータをインデント付きでprintするヘルパー関数
func prettyPrint(i interface{}) {
	b, _ := json.MarshalIndent(i, "", "  ")
	fmt.Println(string(b))
}

//Linux,Windowsによってコマンドが違うのでここで解決する
func genPyCommand() string {
	//"windows" or "linux"
	switch runtime.GOOS {
	case "windows":
		return "python"
	case "linux":
		return "python3"
	default:
		return ""
	}
}
