作业要求：
1.增加0-2秒延时
2.添加Metrics
3.将httpserver部署集群，并完成prometheus配置
4.查看延时指标
5.创建Grafana DashBoard展示延时分配情况

1.增加0-2秒延时：
```
func rootHandler(w http.ResponseWriter, r *http.Request) {
	log.Info("entering root handler")
	//定义指标类型
	timer := metrics.NewTimer()
	defer timer.ObserveTotal()
	user := r.URL.Query().Get("user")
	delay := randInt(10, 2000)
   ...	
}

```
2.添加Metrics：
```
func CreateExecutionTimeMetric(namespace string, help string) *prometheus.HistogramVec {
	return prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Name:      "execution_latency_seconds",
			Help:      help,
			Buckets:   prometheus.ExponentialBuckets(0.001, 2, 15),
		}, []string{"step"},
	)
}

```

3.部署到集群
```
Deploy中新增prometheus的注解：
    metadata:
      annotations:
        prometheus.io/port: "8080"
        prometheus.io/scrape: "true"

```
root@master:~# k get pod -n mangodemo -owide
NAME                          READY   STATUS    RESTARTS      AGE   IP                NODE   NOMINATED NODE   READINESS GATES
httpserver-58f8875d98-phg94   1/1     Running   4 (15m ago)   11d   192.168.167.180   node   <none>           <none>

curl 192.168.167.180/metricsroot@master:~# curl 192.168.167.180:8080/metrics
# HELP go_gc_duration_seconds A summary of the pause duration of garbage collection cycles.
# TYPE go_gc_duration_seconds summary
go_gc_duration_seconds{quantile="0"} 1.8877e-05
go_gc_duration_seconds{quantile="0.25"} 8.501e-05
go_gc_duration_seconds{quantile="0.5"} 0.000181263
go_gc_duration_seconds{quantile="0.75"} 0.000970719
go_gc_duration_seconds{quantile="1"} 0.004781161
...

4.安装grafana
安装helm:
wget https://get.helm.sh/helm-v3.8.1-linux-amd64.tar.gz
tar zxvf helm-v3.8.1-linux-amd64.tar.gz
sudo cp linux-amd64/helm /usr/local/bin/
helm version

安装grafana:
helm repo add grafana https://grafana.github.io/helm-charts
helm pull grafana/loki-stack
tar -xvf loki-stack-2.6.1.tgz
helm upgrade --install loki ./loki-stack --set grafana.enabled=true,prometheus.enabled=true,prometheus.alertmanager.persistentVolume.enabled=false,prometheus.server.persistentVolume.enabled=falase

获取grafana密码：
kubectl get secret loki-grafana -oyaml -n default

root@master:~# echo YWRtaW4=|base64 -d
admin
root@master:~# echo emdIeWZpZENyS0xYVlJEVGdzZFI1aWtmbmxINjJDWmhQdFdvR1RTUg==|base64 -d
zgHyfidCrKLXVRDTgsdR5ikfnlH62CZhPtWoGTSR

查看grafana地址：
root@master:~# k get svc/loki-grafana
NAME           TYPE       CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
loki-grafana   NodePort   10.108.247.79   <none>        80:30855/TCP   15d




