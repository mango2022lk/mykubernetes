package main

/*
编写一个HTTP服务器
1.接收客户端 request，并将 request 中带的 header 写入 response header
2.读取当前系统的环境变量中的 VERSION 配置，并写入 response header
3.Server 端记录访问日志包括客户端 IP，HTTP 返回码，输出到 server 端的标准输出
4.当访问 localhost/healthz 时，应返回200
*/
import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

func HandleServer(res http.ResponseWriter, req *http.Request) {
	// http.Request.Header 本质是：type Header map[string][]string
	//1. 设置res的header
	header := req.Header
	for k, v := range header {
		res.Header().Set(k, v[0])
	}
	//可以手工设置响应码
	res.WriteHeader(http.StatusOK)
	//获取系统VERSION,并设置response header
	res.Header().Set("VERSION", os.Getenv("VERSION"))
	fmt.Fprintf(res, "ok")
	//Server 端记录访问日志包括客户端 IP，HTTP 返回码，输出到 server 端的标准输出
	fmt.Printf("client IP: %s\n", req.RemoteAddr)
	fmt.Printf("client IP: %s\n", req.RequestURI)
	fmt.Printf("Response Status:%d\n", http.StatusOK)
}
func healthzTest(res http.ResponseWriter, req *http.Request) {
	io.WriteString(res, "200")
}
func main() {
	http.HandleFunc("/", HandleServer)
	http.HandleFunc("/healthz", healthzTest)
	err := http.ListenAndServe("localhost:8080", nil)
	if err != nil {
		log.Fatal("listenAndServer: ", err.Error())
	}
}
