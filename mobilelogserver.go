package main

import (
	"code.google.com/p/go.net/websocket"
	"flag"
	"fmt"
	"github.com/GXTime/logging"
	"github.com/navy1125/config"
	"io/ioutil"
	"net/http"
)

var (
	monitorMap map[*websocket.Conn]*websocket.Conn
	//utf8
)

func MonitorServer(ws *websocket.Conn) {
	var message string
	//world := createWorld()

	for {
		err := websocket.Message.Receive(ws, &message)
		if err != nil {
			delete(monitorMap, ws)
			logging.Error("Receive error - stopping worker: %s", err.Error())
			break
		}
		if message == "start" {
			monitorMap[ws] = ws
		}

	}
}
func Broadcask(b []byte) {
	logging.Debug("broadcast:%s", string(b))
	for k, _ := range monitorMap {
		k.Write(b)
	}
}
func LogServer(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	text, err := ioutil.ReadAll(req.Body)
	if err != nil {
		logging.Debug("log err:%s", err.Error())
	}
	//logging.Debug("%s,%s,%s", req.RemoteAddr, req.URL.String(), text)
	Broadcask([]byte(req.RemoteAddr + req.URL.String() + string(text)))
}

func main() {
	flag.Parse()
	config.SetConfig("config", *flag.String("config", "config.xml", "config xml file for start"))
	config.SetConfig("logfilename", *flag.String("logfilename", "/log/logfilename.log", "log file name"))
	config.SetConfig("deamon", *flag.String("deamon", "false", "need run as demo"))
	config.SetConfig("port", *flag.String("port", "8000", "http port "))
	config.SetConfig("log", *flag.String("log", "debug", "logger level "))
	config.LoadFromFile(config.GetConfigStr("config"), "global")
	if err := config.LoadFromFile(config.GetConfigStr("config"), "MobileLogServer"); err != nil {
		fmt.Println(err)
		return
	}
	monitorMap = make(map[*websocket.Conn]*websocket.Conn)
	logger, err := logging.NewTimeRotationHandler(config.GetConfigStr("logfilename"), "060102-15")
	if err != nil {
		fmt.Println(err)
		return
	}
	logger.SetLevel(logging.DEBUG)
	logging.AddHandler("MLOG", logger)
	http.Handle("/ws", websocket.Handler(MonitorServer))
	http.Handle("/", http.FileServer(http.Dir(".")))
	http.HandleFunc("/log/fxsj", LogServer)
	err = http.ListenAndServe(config.GetConfigStr("ip")+":"+config.GetConfigStr("port"), nil)
	if err != nil {
		fmt.Println(err)
		logging.Error("ListenAndServe:%s", err.Error())
	}
}