package main

import (
	"bufio"
	"fmt"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kagurazakayashi/NyarukoMQTTChat/mqttmgr"
)

func main() {
	var confServer string = "tcp://127.0.0.1:1883"
	var confMQTTvar uint = 4
	var confUsername string = ""
	var confPassword string = ""
	var confQos byte = 0
	var confRetained bool = false
	var confTopic string = "test/topic"
	var confLWT string = "异常退出！"

	var client mqtt.Client = mqttmgr.MQTTConnect(confServer, confUsername, confPassword, confQos, confRetained, confTopic, confLWT, confMQTTvar)
	mqttmgr.MQTTSubscribe(client, confTopic, confQos)
	var pubmsg string = ""
	// var publishOK bool =
	fmt.Println("输入消息内容（按回车换行，双击回车提交，/exit 退出）：")
	// osc := make(chan os.Signal, 1)
	// signal.Notify(osc, os.Interrupt, os.Kill)
	// oscs := <-osc
	// logprint("I", "收到指令："+oscs.String())
	// mqttmgr.MQTTDisconnect(client, 300)
	// os.Exit(0)
	counts := make(map[string]int)
	input := bufio.NewScanner(os.Stdin)

	for {
		for input.Scan() {
			if input.Text() == "" {
				break
			} else if input.Text() == "/exit" {
				mqttmgr.MQTTUnsubscribe(client)
				mqttmgr.MQTTDisconnect(client, 300)
				os.Exit(0)
			} else {
				counts[input.Text()]++
				if len(pubmsg) > 0 {
					pubmsg += "\n" + input.Text()
				} else {
					pubmsg += input.Text()
				}
			}
		}
		mqttmgr.MQTTPublish(client, confTopic, pubmsg, confQos, confRetained)
		pubmsg = ""
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
