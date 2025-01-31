package qinglong

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/beego/beego/v2/adapter/httplib"
	"github.com/beego/beego/v2/core/logs"
	"github.com/buger/jsonparser"
	"github.com/cdle/sillyGirl/core"
)

type QingLong struct {
	Host           string `json:"host"`
	ClientID       string `json:"client_id"`
	ClientSecret   string `json:"client_secret"`
	Token          string `json:"-"`
	Error          error  `json:"-"`
	Default        bool   `json:"default"`
	AggregatedMode bool   `json:"aggregated_mode"`
	sync.RWMutex
	idSqlite bool   `json:"-"`
	Name     string `json:"name"`
	Number   int    `json:"-"`
	try      int    `json:"-"`
}

// var Config *QingLong
var qinglong = core.NewBucket("qinglong")
var qLS = []*QingLong{}
var qLSLock = new(sync.RWMutex)

func GetQLS() []*QingLong {
	qLSLock.RLock()
	defer qLSLock.RUnlock()
	return qLS
}

func GetQLSLen() int {
	qLSLock.RLock()
	defer qLSLock.RUnlock()
	return len(qLS)
}

func SetQLS(qls []*QingLong) {
	qLSLock.Lock()
	defer qLSLock.Unlock()
	qLS = qls
}

var expiration int64
var GET = "GET"
var PUT = "PUT"
var POST = "POST"
var DELETE = "DELETE"
var ENVS = "envs"
var CRONS = "crons"
var CONFIG = "configs"

type Carrier struct {
	Get   string
	Value string
}

func init() {
	go func() {
		if !qinglong.GetBool("enable_qinglong", true) {
			return
		}
		initConfig()
		initTask()
		initEnv()
		initqls()
		core.AddCommand("", []core.Function{
			{
				Rules: []string{"青龙管理"},
				Admin: true,
				Handle: func(s core.Sender) interface{} {
					var ql *QingLong
					var ls []string
					nn := []*QingLong{}
					sss := qinglong.Get("QLS")
					json.Unmarshal([]byte(sss), &nn)
					t := ""
					ju := ""
				hh:
					ls = []string{}
					cs := []chan bool{}
					for i := range nn {
						c := make(chan bool)
						cs = append(cs, c)
						go func(i int) {
							nn[i].GetToken()
							close(c)
						}(i)
					}
					for _, c := range cs {
						o, k := <-c
						if o == k {

						}
					}
					for i := range nn {
						t := []string{}
						if nn[i].Token == "" {
							t = append(t, "异常")
						}
						nn[i].Token = ""
						if nn[i].Default {
							t = append(t, "默认")
						}
						if nn[i].AggregatedMode {
							t = append(t, "聚合")
						}

						s := ""
						if len(t) > 0 {
							s = fmt.Sprintf("[%s]", strings.Join(t, ","))
						}
						ls = append(ls, fmt.Sprintf("%d. %s %s", i+1, nn[i].Name, s))
					}
					s.Reply("请选择容器进行编辑：(-删除，0增加，q退出, wq保存)\n" + strings.Join(ls, "\n"))
					r := s.Await(s, nil)
					is := r.(string)
					i := 0
					if is == "wq" || is == "qw" {
						goto save
					}
					if is == "q" {
						goto stop
					}
					if is == "!q" || is == "q!" {
						return "强制退出。"
					}
					if is == "0" {
						ql = &QingLong{}
						nn = append(nn, ql)
					}
					i = core.Int(is)
					if i < 0 && i >= -len(nn) {
						for j := range nn {
							if j == -i-1 {
								nn = append(nn[:j], nn[j+1:]...)
								break
							}
						}
						goto hh
					}
					if i > 0 && i <= len(nn) {
						ql = nn[i-1]
					}
					if ql == nil {
						goto hh
					}
					if ql.Host == "" {
					oo:
						s.Reply("请输入青龙面板地址：")
						ql.Host = regexp.MustCompile(`^(https?://[\.\w]+:?\d*)`).FindString(s.Await(s, nil).(string))
						if ql.Host == "" {
							goto oo
						}
					}
					if ql.ClientID == "" {
						s.Reply("请输入ClientID：")
						ql.ClientID = s.Await(s, nil).(string)
					}
					if ql.ClientSecret == "" {
						s.Reply("请输入ClientSecret：")
						ql.ClientSecret = s.Await(s, nil).(string)
					}
					if ql.Name == "" {
						s.Reply("请输入备注：")
						ql.Name = s.Await(s, nil).(string)
					}
					for {
						if ql.Default {
							t = "移除默认标记"
						} else {
							t = "设置默认标记"
						}
						if ql.AggregatedMode {
							ju = "关闭聚合模式"
						} else {
							ju = "开启聚合模式"
						}
						s.Reply(fmt.Sprintf("请选择要编辑的属性(u返回,q退出,wq保存)：\n%s", strings.Join(
							[]string{
								fmt.Sprintf("1. 容器备注 - %s", ql.Name),
								fmt.Sprintf("2. 面板地址 - %s", ql.Host),
								fmt.Sprintf("3. ClientID - %s", ql.ClientID),
								fmt.Sprintf("4. ClientSecret - %s", ql.ClientSecret),
								fmt.Sprintf("5. %s", t),
								fmt.Sprintf("6. %s", ju),
							}, "\n")))
						switch s.Await(s, nil) {
						default:
							goto hh
						case "1":
							s.Reply("请输入备注：")
							ql.Name = s.Await(s, nil).(string)
						case "2":
						oo1:
							s.Reply("请输入青龙面板地址：")
							ql.Host = regexp.MustCompile(`^(https?://[\.\w]+:?\d*)`).FindString(s.Await(s, nil).(string))
							if ql.Host == "" {
								goto oo1
							}
						case "3":
							s.Reply("请输入ClientID：")
							ql.ClientID = s.Await(s, nil).(string)
						case "4":
							s.Reply("请输入ClientSecret：")
							ql.ClientSecret = s.Await(s, nil).(string)
						case "5":
							ql.Default = !ql.Default
						case "6":
							ql.AggregatedMode = !ql.AggregatedMode
						case "u":
							goto hh
						case "q":
							goto stop
						case "!q", "q!":
							return "强制退出。"
						case "wq", "qw":
							goto save
						}
					}
				stop:
					s.Reply("是否保存修改？(Y/n)")
					if s.Await(s, func(s core.Sender) interface{} {
						return core.YesNo
					}) == core.No {
						return "未作修改。"
					}
				save:
					SetQLS(nn)
					d, _ := json.Marshal(nn)
					qinglong.Set("QLS", string(d))
					return "已保存修改。"

				},
			},
		})
		initCron()
	}()
}

func initqls() {
	s := qinglong.Get("QLS")
	nn := []*QingLong{}
	json.Unmarshal([]byte(s), &nn)
	if len(nn) == 0 {
		Config := &QingLong{}
		Config.Host = regexp.MustCompile(`^(https?://[\.\w]+:?\d*)`).FindString(qinglong.Get("host"))
		Config.ClientID = qinglong.Get("client_id")
		Config.ClientSecret = qinglong.Get("client_secret")
		if Config.Host != "" {
			nn = append(nn, Config)
			d, _ := json.Marshal(nn)
			qinglong.Set("QLS", string(d))
		}
	}
	for _, ql := range nn {
		if ql.Name == "" {
			ql.Name = ql.Host
		}
		if ql.Host == "" {

		}
		_, err := ql.GetToken()
		if err == nil {
			logs.Info("青龙面板(%s)连接成功。", ql.Name)
		} else {
			logs.Warn("青龙面板(%s)连接错误，%v", ql.Name, err)
		}
	}
	SetQLS(nn)
	logs.Info("青龙360安全卫士为您保驾护航，杜绝一切流氓脚本！")
}

func (ql *QingLong) SetNumber(i int) {
	ql.Lock()
	defer ql.Unlock()
	ql.Number = i
}

func (ql *QingLong) GetNumber() int {
	ql.RLock()
	defer ql.RUnlock()
	return ql.Number
}

func (ql *QingLong) SetClientID(i string) {
	ql.Lock()
	defer ql.Unlock()
	ql.ClientID = i
}

func (ql *QingLong) GetClientID() string {
	ql.RLock()
	defer ql.RUnlock()
	return ql.ClientID
}

func (ql *QingLong) SetClientSecret(i string) {
	ql.Lock()
	defer ql.Unlock()
	ql.ClientSecret = i
}

func (ql *QingLong) GetClientSecret() string {
	ql.RLock()
	defer ql.RUnlock()
	return ql.ClientSecret
}

func (ql *QingLong) SetHost(i string) {
	ql.Lock()
	defer ql.Unlock()
	ql.Host = i
}

func (ql *QingLong) GetHost() string {
	ql.RLock()
	defer ql.RUnlock()
	return ql.Host
}

func (ql *QingLong) SetName(i string) {
	ql.Lock()
	defer ql.Unlock()
	ql.Name = i
}

func (ql *QingLong) SetIsSqlite() {
	ql.Lock()
	defer ql.Unlock()
	ql.idSqlite = true
}

func (ql *QingLong) IsSqlite() bool {
	ql.RLock()
	defer ql.RUnlock()
	return ql.idSqlite
}

func (ql *QingLong) GetName() string {
	ql.RLock()
	defer ql.RUnlock()
	return ql.Name
}

func (ql *QingLong) SetToken(i string) {
	ql.Lock()
	defer ql.Unlock()
	ql.Token = i
}

func (ql *QingLong) AddTry() {
	ql.Lock()
	defer ql.Unlock()
	ql.try = ql.try + 1
}

func (ql *QingLong) SetTry(i int) {
	ql.Lock()
	defer ql.Unlock()
	ql.try = i
}

func (ql *QingLong) GetToken() (string, error) {
	ql.RLock()
	defer ql.RUnlock()

	// if ql.try >= 2 {
	// 	return "", errors.New(fmt.Sprintf("%s异常。", ql.Name))
	// }

	if ql.Token != "" && expiration > time.Now().Unix() {
		return ql.Token, nil
	}
	req := httplib.Get(fmt.Sprintf("%s/open/auth/token?client_id=%s&client_secret=%s", ql.Host, ql.ClientID, ql.ClientSecret))
	req.SetTimeout(time.Second*2, time.Second*2)
	data, err := req.Bytes()
	if err != nil {
		msg := fmt.Sprintf("青龙连接失败：%v", err)
		// logs.Warn(msg)
		return "", errors.New(msg)
	}
	code, _ := jsonparser.GetInt(data, "code")
	if code != 200 {
		msg := fmt.Sprintf("青龙登录失败：%v", string(data))
		// logs.Warn(msg)
		return "", errors.New(msg)
	}
	ql.Token, _ = jsonparser.GetString(data, "data", "token")
	// if ql.Token == "" {
	// 	go ql.SetTry(0)
	// } else {
	// 	go ql.AddTry()
	// }
	expiration, _ = jsonparser.GetInt(data, "data", "expiration")
	return ql.Token, nil
}

func Req(p interface{}, ps ...interface{}) (*QingLong, error) {
	var nn = GetQLS()
	if len(nn) == 0 {
		return nil, errors.New("未配置容器。")
	}
	var s core.Sender
	var ql *QingLong
	var qls []*QingLong
	switch p.(type) {
	case core.Sender:
		s = p.(core.Sender)
	case *QingLong:
		ql = p.(*QingLong)
	case []*QingLong:
		qls = p.([]*QingLong)
	}
	if qls != nil {
		for i := range qls {
			Req(qls[i], ps...)
		}
		return nil, nil
	}
	if s != nil && !s.IsAdmin() { //普通用户自动分配
		for i := range nn {
			if nn[i].Default {
				ql = nn[i]
				break
			}
		}
		if ql == nil {
			ql = nn[0]
		}
	}
	if ql == nil {
		if len(nn) > 1 {
			if s != nil {
				ls := []string{}
				for i := range nn {
					ls = append(ls, fmt.Sprintf("%d. %s", i+1, nn[i].Name))
				}
				ls = append(ls, fmt.Sprintf("%d. %s", len(nn)+1, "所有容器"))
				s.Reply("请选择容器：\n" + strings.Join(ls, "\n"))
				r := s.Await(s, func(s core.Sender) interface{} {
					return core.Range([]int{1, len(nn) + 1})
				}, time.Second*10)
				switch r {
				case nil:
				default:
					ql = nn[r.(int)-1]
				}
			}
		} else {
			ql = nn[0]
		}
	}

	if ql == nil {
		for i := range nn {
			if nn[i].Default {
				ql = nn[i]
				if s != nil {
					s.Reply(fmt.Sprintf("已默认选择容器%s", ql.Name))
				}
				break
			}
		}
	}

	if ql == nil {
		return nil, errors.New("未选择容器。")
	}
	// start:
	token, err := ql.GetToken()
	if err != nil {
		return nil, err
	}
	method := GET
	body := []byte{}
	api := ENVS
	apd := ""
	var get *string
	var c *Carrier
	var toParse interface{}
	for _, p := range ps {
		switch p.(type) {
		case string:
			switch p.(string) {
			case GET, POST, DELETE, PUT:
				method = p.(string)
			case ENVS, CRONS, CONFIG:
				api = p.(string)
			default:
				apd = p.(string)
			}
		case []byte:
			body = p.([]byte)
		case *Carrier:
			c = p.(*Carrier)
		case *string:
			get = p.(*string)
		default:
			if strings.Contains(reflect.TypeOf(p).String(), "*") {
				toParse = p
			} else {
				body, _ = json.Marshal(p)
			}
		}
	}
	var req *httplib.BeegoHTTPRequest
	api += apd
	api = strings.Trim(api, " ")
	switch method {
	case GET:
		req = httplib.Get(ql.GetHost() + "/open/" + api)
	case POST:
		req = httplib.Post(ql.GetHost() + "/open/" + api)
	case DELETE:
		req = httplib.Delete(ql.GetHost() + "/open/" + api)
	case PUT:
		req = httplib.Put(ql.GetHost() + "/open/" + api)
	}
	req.Header("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header("Content-Type", "application/json;charset=UTF-8")
	req.SetTimeout(time.Second*5, time.Second*5)
	if method != GET {
		if ql.IsSqlite() {
			s := string(body)
			for _, v := range regexp.MustCompile(`"_id":"(\d+)",`).FindAllStringSubmatch(s, -1) {
				s = strings.Replace(s, v[0], `"id":`+v[1]+`,`, -1)
			}
			body = []byte(s)
		}
		req.Body(body)
	}
	data, err := req.Bytes()

	// if strings.Contains(string(data), "UnauthorizedError") {
	// 	ql.SetToken("")
	// 	goto start
	// }

	if err != nil {
		return nil, err
	}
	if strings.Contains(string(data), `"id"`) {
		s := string(data)
		for _, v := range regexp.MustCompile(`"id":(\d+),`).FindAllStringSubmatch(s, -1) {
			s = strings.Replace(s, v[0], `"_id":"`+v[1]+`",`, -1)
		}
		data = []byte(s)
		if !ql.IsSqlite() {
			ql.SetIsSqlite()
		}
	}

	code, _ := jsonparser.GetInt(data, "code")
	if code != 200 {
		return nil, errors.New(string(data))
	}
	if toParse != nil {
		if err := json.Unmarshal(data, toParse); err != nil {
			return nil, err
		}
	}
	if get != nil {
		if *get, err = jsonparser.GetString(data, *get); err != nil {
			return nil, err
		}
	}
	if c != nil {
		c.Value, _ = jsonparser.GetString(data, strings.Split(c.Get, ".")...)
	}
	return ql, nil
}

func GetQinglongByClientID(s string) (error, *QingLong) {
	nn := GetQLS()
	for i := range nn {
		if nn[i].ClientID == s {
			return nil, nn[i]
		}
	}
	if len(nn) == 0 {
		return errors.New("未配置容器。"), nil
	}
	var ql *QingLong
	min := 10000000
	for i := range nn {
		if num := nn[i].GetNumber(); num <= min {
			min = num
			ql = nn[i]
		}
	}
	return errors.New("默认获取了一个容器。"), ql
}

func QinglongSC(s core.Sender) (error, []*QingLong) {
	nn := GetQLS()
	if len(nn) == 0 {
		return errors.New("未配置容器。"), nil
	}
	if len(nn) == 1 {
		return nil, nn
	}
	var ql *QingLong
	if s != nil && !s.IsAdmin() { //普通用户自动分配
		for i := range nn {
			if nn[i].Default {
				ql = nn[i]
				break
			}
		}
		if ql == nil {
			ql = nn[0]
		}
	}
	if ql != nil {
		return nil, []*QingLong{ql}
	}
	ls := []string{}
	for i := range nn {
		ls = append(ls, fmt.Sprintf("%d. %s", i+1, nn[i].Name))
	}
	s.Reply("请选择容器：\n" + strings.Join(ls, "\n"))
	r := s.Await(s, func(s core.Sender) interface{} {
		return core.Range([]int{1, len(nn)})
	}, time.Second*10)
	switch r {
	case nil:
		s.Reply()
		return errors.New("你没有选择容器。"), []*QingLong{}
	default:
		index := r.(int) - 1
		if index != len(nn) {
			return nil, []*QingLong{nn[index]}
		} else {
			return nil, nn
		}
	}
}
