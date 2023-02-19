package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"io"
	"math/rand"
	"mykubernetes/module10/httpserver/metrics"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var logLevel string
var viperInstance *viper.Viper

func main() {
	flag.Set("v", "4")
	log.Info("Starting http server...")
	metrics.Register()

	httpserverConf := os.Getenv("HTTPSERVER_CONF")
	log.Info("configFile from env is " + httpserverConf)
	if httpserverConf == "" {
		httpserverConf = "/etc/httpserver/httpserver.properties"
	}
	log.Info("confFile is " + httpserverConf)
	viperInstance = viper.New()                 // viper实例
	viperInstance.SetConfigFile(httpserverConf) // 指定配置文件路径

	err := viperInstance.ReadInConfig()
	if err != nil { // 处理读取配置文件的错误
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}
	viperInstance.WatchConfig()
	viperInstance.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Detect config change: %s \n", e.String())
		log.Warn("Config file updated.")
		viperLoadConf(viperInstance) // 加载配置的方法
	})
	viperLoadConf(viperInstance)

	port := 8080
	portstr := ":" + strconv.Itoa(port)
	log.Info("httpserver listend port: ", port)

	mux := http.NewServeMux()
	mux.HandleFunc("/", HttpHandler)
	mux.HandleFunc("/healthz", HealthZ)
	mux.HandleFunc("/sleep", SleepTest)
	mux.HandleFunc("/hello", rootHandler)
	mux.Handle("/metrics", promhttp.Handler())
	server := &http.Server{
		Addr:    portstr,
		Handler: mux,
	}
	go server.ListenAndServe()
	listenSignal(context.Background(), server)
}

func dynamicConfig() {
	viperInstance.WatchConfig()
	viperInstance.OnConfigChange(func(e fsnotify.Event) {
		fmt.Println("Detect config change: %s \n", e.String())
		log.Warn("Config file updated.")
		viperLoadConf(viperInstance) // 加载配置的方法
	})
}

func viperLoadConf(viperInstance *viper.Viper) {
	logLevel = viperInstance.GetString("log_level")
	level, err := log.ParseLevel(logLevel)
	if err != nil {
		level = log.GetLevel()
	}
	log.SetLevel(level)
	myconf := viperInstance.GetString("my_conf")
	log.Trace(myconf + " in viperLoadConf")
	log.Debug(myconf + " in viperLoadConf")
	log.Info(myconf + " in viperLoadConf")
	log.Warn(myconf + " in viperLoadConf")
	log.Error(myconf + " in viperLoadConf")
}

func HealthZ(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}
func HttpHandler(w http.ResponseWriter, r *http.Request) {
	myconf := viperInstance.GetString("my_conf")
	log.Trace(myconf)
	log.Debug(myconf)
	log.Info(myconf)
	log.Warn(myconf)
	log.Error(myconf)
	ip := ClientIP(r)
	httpcode := 200
	log.Info("reqLog: clientIP : " + ip + " httpcode " + strconv.Itoa(httpcode))
	headers := r.Header
	log.Info("req headers : %+v", headers)
	for k, v := range headers {
		for _, val := range v {
			w.Header().Set(k, val)
		}
	}
	version := os.Getenv("VERSION")
	if version != "" {
		w.Header().Set("VERSION", version)
	}
	w.WriteHeader(httpcode)
	io.WriteString(w, "ok")
}

func ClientIP(r *http.Request) string {
	xForwardedFor := r.Header.Get("X-Forwarded-For")
	ip := strings.TrimSpace(strings.Split(xForwardedFor, ",")[0])
	if ip != "" {
		return ip
	}
	ip = strings.TrimSpace(r.Header.Get("X-Real-Ip"))
	if ip != "" {
		return ip
	}
	if ip, _, err := net.SplitHostPort(strings.TrimSpace(r.RemoteAddr)); err == nil {
		return ip
	}
	return ""
}

func SleepTest(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()
	sleeptimeStr := values.Get("sleep")
	sleeptime, err := strconv.Atoi(sleeptimeStr)
	if err != nil {
		sleeptime = 1
	}
	time.Sleep(time.Duration(sleeptime) * time.Second)
	fmt.Fprintln(w, "Hello world, sleep "+strconv.Itoa(sleeptime)+"s")
	log.Printf("Hello world, sleep %+vs", sleeptime)
}

//定义随机的访问延迟
func rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("entering root handler")
	//定义指标类型
	timer := metrics.NewTimer()
	defer timer.ObserveTotal()
	user := r.URL.Query().Get("user")
	delay := randInt(10, 2000)
	time.Sleep(time.Millisecond * time.Duration(delay))
	if user != "" {
		io.WriteString(w, fmt.Sprintf("hello [#{user}]\n"))
	} else {
		io.WriteString(w, "hello [stranger]\n")
	}
	io.WriteString(w, "==========================Details of the http request header;====================================\n")
	log.Infof("Respond in %d ms", delay)
}

// 优雅终止
func listenSignal(ctx context.Context, httpSrv *http.Server) {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case <-sigs:
		fmt.Println("notify sigs")
		httpSrv.Shutdown(ctx)
		fmt.Println("http shutdown gracefully")
	}
}

func randInt(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
