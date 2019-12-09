package mqttmgr

import (
	"fmt"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTConnect : 连接到服务器
func MQTTConnect(server string, clientid string, username string, password string, qos byte, retained bool, topic string, lwt string, mqttvar uint) mqtt.Client {
	logprint("I", "正在连接到服务器 "+server+" ...")
	opts := mqtt.NewClientOptions().AddBroker(server).SetClientID("gotest")
	opts.SetProtocolVersion(mqttvar)
	if len(username) > 0 {
		opts.SetUsername("username")
	}
	if len(password) > 0 {
		opts.SetPassword("password")
	}
	opts.SetWill(topic, lwt, qos, retained)
	var client mqtt.Client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		logprint("E", "连接失败： "+token.Error().Error()+" 。")
		os.Exit(1)
	} else {
		logprint("I", "已连接。")
	}
	return client
}

// MQTTDisconnect : 断开连接
func MQTTDisconnect(client mqtt.Client, timeout uint) {
	client.Disconnect(timeout)
	logprint("I", "主动断开连接。")
}

// MQTTPublish : 发送消息
func MQTTPublish(client mqtt.Client, topic string, pubmsg string, qos byte, retained bool) bool {
	if len(pubmsg) == 0 {
		logprint("E", "消息不能为空！")
		return false
	}
	logprint("I", "正在发送信息...")
	if token := client.Publish(topic, qos, retained, pubmsg); token.Wait() && token.Error() != nil {
		logprint("E", "消息发送失败： "+token.Error().Error()+" 。")
		return false
	}
	logprint("I", "已发送消息。")
	return true
}

// MQTTSubscribe : 订阅
func MQTTSubscribe(client mqtt.Client, topic string, qos byte, msgRcvd mqtt.MessageHandler) bool {

	if token := client.Subscribe(topic, qos, msgRcvd); token.Wait() && token.Error() != nil {
		logprint("E", "订阅失败： "+token.Error().Error()+" 。")
		return false
	}
	logprint("I", "已订阅 "+topic+" 。")
	return true
}

// MQTTUnsubscribe : 取消订阅
func MQTTUnsubscribe(client mqtt.Client) bool {
	if token := client.Unsubscribe("example/topic"); token.Wait() && token.Error() != nil {
		logprint("E", "取消订阅失败： "+token.Error().Error()+" 。")
		return false
	}
	logprint("I", "已取消订阅。")
	return true
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
