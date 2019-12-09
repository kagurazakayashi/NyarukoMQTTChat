package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kagurazakayashi/NyarukoMQTTChat/mqttmgr"
)

//Msgjson : 传输的 JSON 结构
type Msgjson struct {
	S int8   //状态码
	V string //版本
	N string //昵称
	T int64  //时间戳
	M string //消息内容
}

func main() {
	var confServer string = "tcp://127.0.0.1:1883"
	var confMQTTvar uint = 4
	var confUsername string = ""
	var confPassword string = ""
	var confQos byte = 0
	var confRetained bool = false
	var confTopic string = "test/topic"
	var confNick string = "神楽坂雅詩"
	var confConnect string = "加入会话" //S=1
	var confLWT string = "连接中断"     //S=-2
	var confExit string = "离开会话"    //S=-1
	var confClientID string = "NyarukoMQTTChat/" + randString(16)

	var pubmsg string = ""
	var client mqtt.Client = mqttmgr.MQTTConnect(confServer, confClientID, confUsername, confPassword, confQos, confRetained, confTopic, newmsg(-2, confNick, ""), confMQTTvar)

	//数据接收代理
	var msgRcvd mqtt.MessageHandler = func(nclient mqtt.Client, msg mqtt.Message) {
		var msgstr string = string(msg.Payload())
		var msgstrdata []byte = []byte(msgstr)
		tmsg := Msgjson{}
		err := json.Unmarshal(msgstrdata, &tmsg)
		if err != nil {
			logprint("E", "JSON 解码失败： "+err.Error())
		} else {
			if tmsg.V != "NMC_1" {
				logprint("E", "对方的 NyarukoMQTTChat 版本不匹配。")
			} else {
				var msgtimei int64 = int64(tmsg.T)
				var msgtimet time.Time = time.Unix(msgtimei, 0)
				var msgtimes string = timestr(msgtimet)
				if tmsg.S == 0 {
					logprint("M", "----------")
					logprint("M", tmsg.N+" ("+msgtimes+") : ")
					var msgchars []string = strings.Split(tmsg.M, "\n")
					for i := 0; i < len(msgchars); i++ {
						logprint("M", "    "+msgchars[i])
						if len(pubmsg) > 0 {
							fmt.Println(pubmsg)
						}
					}
					logprint("M", "----------")
				} else {
					var alertstr string = ""
					switch tmsg.S {
					case 1:
						alertstr = confConnect
						break
					case -2:
						alertstr = confLWT
						break
					case -1:
						alertstr = confExit
						break
					}
					if tmsg.S == -2 {
						logprint("S", tmsg.N+" "+alertstr)
					} else {
						logprint("S", tmsg.N+" "+alertstr)
					}
				}
			}
		}
	}
	mqttmgr.MQTTPublish(client, confTopic, newmsg(1, confNick, ""), confQos, confRetained)
	mqttmgr.MQTTSubscribe(client, confTopic, confQos, msgRcvd)
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
				mqttmgr.MQTTPublish(client, confTopic, newmsg(-1, confNick, ""), confQos, confRetained)
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
		if len(pubmsg) > 0 {
			mqttmgr.MQTTPublish(client, confTopic, newmsg(0, confNick, pubmsg), confQos, confRetained)
			pubmsg = ""
		}
	}
}

//创建新消息格式
func newmsg(status int8, nick string, msg string) string {
	smsg := Msgjson{
		S: status,
		V: "NMC_1",
		N: nick,
		T: time.Now().Unix(),
		M: msg,
	}
	jsonbytes, err := json.Marshal(smsg)
	if err != nil {
		logprint("E", "JSON 编码失败： "+err.Error())
		return ""
	}
	return string(jsonbytes)
}

//创建随机字符串
func randString(len int) string {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := r.Intn(26) + 65
		bytes[i] = byte(b)
	}
	return string(bytes)
}

// 在控制台输出日志
func logprint(mtype string, msg string) {
	fmt.Println(fmt.Sprintf("[%s %s] %s", timestr(time.Now()), mtype, msg))
}

// 取得时间日期字符串
func timestr(ntime time.Time) string {
	curHour, curMinute, curSecond := ntime.Clock()
	curYear, curMonth, curDay := ntime.Date()
	var dates string = fmt.Sprintf("%d-%02d-%02d %d:%02d:%02d", curYear, curMonth, curDay, curHour, curMinute, curSecond)
	return dates
}
