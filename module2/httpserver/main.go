package main

/*
1.接收客户端 request，并将 request 中带的 header 写入 response
2.header读取当前系统的环境变量中的 VERSION 配置，并写入 response header
3.Server 端记录访问日志包括客户端 IP，HTTP 返回码，输出到 server 端的标准输出
4.当访问 {url}/healthz 时，应返回200
*/

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
)

func index(w http.ResponseWriter, r *http.Request) {
	//2.header读取当前系统的环境变量中的 VERSION 配置，并写入 response header
	os.Setenv("VERSION", "V1")
	version := os.Getenv("VERSION")
	w.Header().Set("VERSION", version)
	fmt.Printf("os version : %s \n", version)
	//1.接收客户端 request，并将 request 中带的 header 写入 response
	for k, v := range r.Header {
		//v是一个slice
		for _, vv := range v {
			//打印下对应的信息
			fmt.Printf("header key :%s, header value : %s\n", k, vv)
			w.Header().Set(k, vv)
		}
	}

	//04.记录日志并输出
	clinetIp := getCurrentIP(r)
	log.Printf("Success! Response code: %d", http.StatusOK)
	log.Printf("Success! clientip : %s", clinetIp)
}

func getCurrentIP(r *http.Request) string {
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = strings.Split(r.RemoteAddr, ":")[0]
	}
	return ip
}

//4.健康检查
func healthz(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "working")
}
func main() {
	//定义一个总的路由来分发请求
	mux := http.NewServeMux()

	mux.HandleFunc("/", index)
	mux.HandleFunc("/healthz", healthz)

	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatalf("start http server fialed,error : %s\n", err.Error())
	}
}
