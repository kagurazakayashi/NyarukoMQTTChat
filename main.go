package main

import (
	"fmt"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func main() {
	var confServer string = "tcp://127.0.0.1:1883"
	var confUsername string = ""
	var confPassword string = ""
	var confQos byte = 0
	var confRetained bool = false

	var client mqtt.Client = mqttConnect(confServer, confUsername, confPassword, confQos, confRetained)
	var pubmsg string = timestr() + " 测试消息"
	// var publishOK bool =
	mqttPublish(client, pubmsg, confQos, confRetained)
	mqttDisconnect(client, 300)
}

// MQTT: 连接到服务器
func mqttConnect(server string, username string, password string, qos byte, retained bool) mqtt.Client {
	logprint("I", "正在连接到服务器 "+server+" ...")
	opts := mqtt.NewClientOptions().AddBroker(server).SetClientID("gotest")
	opts.SetProtocolVersion(4)
	// opts.SetUsername("username")
	// opts.SetPassword("password")
	opts.SetWill("test/topic", "异常退出", qos, retained)
	var client mqtt.Client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		logprint("E", "连接失败： "+token.Error().Error()+" 。")
		os.Exit(1)
	} else {
		logprint("I", "已连接")
	}
	return client
}

// MQTT: 断开连接
func mqttDisconnect(client mqtt.Client, timeout uint) {
	client.Disconnect(timeout)
	logprint("I", "主动断开连接")
}

// MQTT: 发送消息
func mqttPublish(client mqtt.Client, pubmsg string, qos byte, retained bool) bool {
	logprint("I", "正在发送信息： "+pubmsg+" ...")
	if token := client.Publish("test/topic", qos, retained, pubmsg); token.Wait() && token.Error() != nil {
		logprint("E", "消息发送失败： "+token.Error().Error()+" 。")
		return true
	} else {
		logprint("I", "已发送消息。")
		return false
	}
}

// 在控制台输出日志
func logprint(mtype string, msg string) {
	fmt.Println(fmt.Sprintf("[%s %s] %s", timestr(), mtype, msg))
}

// 取得时间日期字符串
func timestr() string {
	var timer time.Time = time.Now()
	curHour, curMinute, curSecond := timer.Clock()
	curYear, curMonth, curDay := timer.Date()
	var dates string = fmt.Sprintf("%d-%02d-%02d %d:%02d:%02d", curYear, curMonth, curDay, curHour, curMinute, curSecond)
	return dates
}
