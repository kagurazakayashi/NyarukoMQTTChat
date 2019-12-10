package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"flag"
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

var (
	confServer   string
	confMQTTvar  uint = 4
	confUsername string
	confPassword string
	confQos      byte = 0
	confQosS     string
	confQosNum   uint64
	confRetained bool = false
	confTopic    string
	confNick     string
	confConnect  string //S=1
	confLWT      string //S=-2
	confExit     string //S=-1
)

func init() {
	flag.StringVar(&confServer, "server", "tcp://127.0.0.1:1883", "服务器地址")
	flag.UintVar(&confMQTTvar, "mqttvar", 4, "服务器版本")
	flag.StringVar(&confUsername, "user", "", "服务器用户名")
	flag.StringVar(&confPassword, "password", "", "服务器密码")
	flag.Uint64Var(&confQosNum, "qos", 2, "通讯等级(0:最多一次的传输/1:至少一次的传输/2:只有一次的传输)")
	if confQosS == "0" || confQosS == "1" || confQosS == "2" {
		var buf = make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, confQosNum)
		confQos = buf[0]
	}
	flag.BoolVar(&confRetained, "retain", true, "允许离线消息(true/false)")
	flag.StringVar(&confTopic, "topic", "Public", "频道名称")
	confTopic = "NyarukoMQTTChat/" + confTopic
	flag.StringVar(&confNick, "nick", "匿名用户", "你的昵称")
	flag.StringVar(&confConnect, "txtconn", "加入会话", "别人加入会话时的提示信息")
	flag.StringVar(&confLWT, "txtlwt", "连接中断", "别人掉线时的提示信息")
	flag.StringVar(&confExit, "txtexit", "离开会话", "别人离开会话时的提示信息")
}

func main() {
	flag.Parse()
	var pubmsg string = ""
	var confClientID string = "NyarukoMQTTChat" + randString(16)
	var client mqtt.Client = mqttmgr.MQTTConnect(confServer, confClientID, confUsername, confPassword, confQos, confRetained, confTopic, newmsg(-2, confNick, ""), confMQTTvar)
	logprint("I", "昵称： "+confNick)
	//数据接收代理
	var msgRcvd mqtt.MessageHandler = func(nclient mqtt.Client, msg mqtt.Message) {
		var msgstr string = string(msg.Payload())
		var msgstrdata []byte = []byte(msgstr)
		var tmsg Msgjson = Msgjson{}
		var jsonerr error = json.Unmarshal(msgstrdata, &tmsg)
		if jsonerr != nil {
			logprint("E", "JSON 解码失败： "+jsonerr.Error())
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
