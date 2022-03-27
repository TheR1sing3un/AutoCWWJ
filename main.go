package main

import (
	"encoding/json"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/robfig/cron"
	"github.com/spf13/viper"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var users Users

func main() {
	config, err := initConfig()
	if config == nil {
		log.Printf("create config err : %v\n", err)
		return
	}
	cron := cron.New()
	//每天早上8点,中午1点执行
	err = cron.AddFunc("0 0 8,13 * * *", auto)
	if err != nil {
		log.Printf("err add func in cron : %v\n", err)
		return
	}
	cron.Start()
	defer cron.Stop()
	select {}
}

func init() {
	file := "./" + "project-log" + ".log"
	logFile, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	log.SetOutput(logFile) // 将文件设置为log输出的文件
	log.SetPrefix("[AutoCWWJ]")
	log.SetFlags(log.LstdFlags | log.Lshortfile | log.LUTC)
	return
}

func initConfig() (*viper.Viper, error) {
	//新建一个viper
	getwd, _ := os.Getwd()
	v := viper.New()
	//v.SetConfigFile("./config.yaml")
	v.SetConfigName("config") // 设置文件名称（无后缀）
	v.SetConfigType("yaml")   // 设置后缀名 {"1.6以后的版本可以不设置该后缀"}
	v.AddConfigPath(getwd)    // 设置文件所在路径
	v.ReadInConfig()
	err := v.Unmarshal(&users)
	fmt.Println(users)
	if err != nil {
		log.Printf("err unmarshall users : %v\n", err)
		return nil, err
	}
	//监控配置更改
	v.WatchConfig()
	v.OnConfigChange(func(in fsnotify.Event) {
		log.Printf("Config file changed : %v\n", in.Name)
		v.ReadInConfig()
		err := v.Unmarshal(&users)
		fmt.Println(users)
		if err != nil {
			log.Printf("err unmarshall users : %v\n", err)
			return
		}
	})
	return v, nil
}

func auto() {
	for _, user := range users.Users {
		go autoDeal(user)
	}
}

func autoDeal(user User) {
	vals := url.Values{}
	vals.Add("username", user.Username)
	vals.Add("password", user.Password)
	m := make(map[string]string)
	resp, err := http.PostForm("https://xxcapp.xidian.edu.cn/uc/wap/login/check", vals)
	data, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(data, &m)
	respStr := m["m"]
	if err != nil || resp.StatusCode != 200 || !strings.Contains(respStr, "操作成功") {
		log.Printf("err post to login : %v;resp code : %v ; resp str : %v\n", err, resp.StatusCode, respStr)
		return
	}
	cookies := resp.Cookies()
	vals2 := url.Values{}
	vals2.Add("sfzx", "1")
	vals2.Add("tw", "1")
	vals2.Add("sfcyglq", "0")
	vals2.Add("sfyzz", "0")
	vals2.Add("qtqk", "")
	vals2.Add("ymtys", "0")
	vals2.Add("area", "陕西省 西安市 长安区")
	vals2.Add("city", "西安市")
	vals2.Add("province", "陕西省")
	vals2.Add("address", "陕西省西安市长安区兴隆街道西安电子科技大学南校区")
	vals2.Add("geo_api_info", "{\"type\":\"complete\",\"position\":{\"Q\":34.125585394966,\"R\":108.83212402343798,\"lng\":108.832124,\"lat\":34.125585},\"location_type\":\"html5\",\"message\":\"Get ipLocation failed.Get geolocation success.Convert Success.Get address success.\",\"accuracy\":35,\"isConverted\":true,\"status\":1,\"addressComponent\":{\"citycode\":\"029\",\"adcode\":\"610116\",\"businessAreas\":[],\"neighborhoodType\":\"\",\"neighborhood\":\"\",\"building\":\"\",\"buildingType\":\"\",\"street\":\"雷甘路\",\"streetNumber\":\"266#\",\"country\":\"中国\",\"province\":\"陕西省\",\"city\":\"西安市\",\"district\":\"长安区\",\"township\":\"兴隆街道\"},\"formattedAddress\":\"陕西省西安市长安区兴隆街道西安电子科技大学长安校区西二楼B西安电子科技大学南校区\",\"roads\":[],\"crosses\":[],\"pois\":[],\"info\":\"SUCCESS\"}")
	request, err := http.NewRequest("POST", "https://xxcapp.xidian.edu.cn/xisuncov/wap/open-report/save", strings.NewReader(vals2.Encode()))
	if err != nil {
		log.Printf("err create request : %v\n", err)
		return
	}
	for _, cookie := range cookies {
		request.AddCookie(cookie)
	}
	client := http.Client{}
	response, err := client.Do(request)
	defer response.Body.Close()
	if err != nil {
		log.Printf("err request : %v\n", err)
		return
	}
	data, err = ioutil.ReadAll(response.Body)
	if err != nil {
		log.Printf("err read response body : %v\n", err)
		return
	}
	json.Unmarshal(data, &m)
	res := m["m"]
	req, err := http.NewRequest(http.MethodGet, "https://push.gh.117503445.top:20000/push/text/v2", nil)
	if err != nil {
		return
	}
	query := req.URL.Query()
	query.Add("name", user.PusherName)
	query.Add("text", res)
	req.URL.RawQuery = query.Encode()
	pushResp, err := client.Do(req)
	if err != nil {
		log.Printf("err push : %v\n", err)
		return
	}
	log.Printf("push to %v || push content : %v || push result : %v\n", user.Username, res, pushResp)
}

type Users struct {
	Users []User `yaml:"users"`
}

type User struct {
	Username   string `yaml:"username"`
	Password   string `yaml:"password"`
	PusherName string `yaml:"pusherName"`
}
