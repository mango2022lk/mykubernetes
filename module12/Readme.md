作业描述： 

将httpserver以istio ingress gateway的形式发布: 

1.实现安全保证 

2.七层路由规则 

3.考虑open tracing的接入



1.安装istio: 

```shell
wget https://github.com/istio/istio/releases/download/1.12.0/istio-1.12.0-linux-amd64.tar.gz 
tar -zxvf istio-1.12.0-linux-amd64.tar.gz 
cd istio-1.12.0 
cp bin/istioctl /usr/local/bin 
istioctl install --set profile=demo -y
```

查看pod:

```shell
root@master:~# k get pod -n istio-system
NAME                                   READY   STATUS    RESTARTS        AGE
istio-egressgateway-7f4864f59c-sk8c4   1/1     Running   3 (3m40s ago)   3d21h
istio-ingressgateway-55d9fb9f-w87rz    1/1     Running   3 (3m34s ago)   3d21h
istiod-5d5b675975-jx5tk                1/1     Running   3 (3m34s ago)   3d21h
jaeger-5d44bc5c5d-457kj                1/1     Running   2 (3m34s ago)   2d21h
```



2.安装jaeger，用来抓取tracing 

```shell
kubectl apply -f jaeger.yaml

jaeger-5d44bc5c5d-457kj                1/1     Running   2 (3m34s ago)   2d21h
```

3.部署服务 

```shell
kubectl create ns tracing
#为tracing下的pod注入sidecar 
kubectl label ns tracing istio-injection=enabled
kubectl -n tracing apply -f service0.yaml
kubectl -n tracing apply -f service1.yaml
kubectl -n tracing apply -f service2.yaml
kubectl apply -f istio-specs-https.yaml -n tracing
```





七层路由配置：

```
VirtualService:
 ...
  hosts:
    - '*'
  http:
  - match:
      - uri:
          exact: /service0
    route:
      - destination:
          host: service0
 ...
```

4.查看ingressIP: 

```shell
k get svc -n istio-system
root@master:~# k get svc -n istio-system
NAME                   TYPE           CLUSTER-IP       EXTERNAL-IP      PORT(S)                                                                      AGE
istio-egressgateway    ClusterIP      10.101.131.87    <none>           80/TCP,443/TCP                                                               2d21h
istio-ingressgateway   LoadBalancer   10.102.235.84    192.168.91.161   15021:32448/TCP,80:30817/TCP,443:32038/TCP,31400:30240/TCP,15443:31480/TCP

ingressgateway的LB地址是：
192.168.91.161 
```



5.查看tracing：

 ```shell
 root@master:~# curl 192.168.91.161/service0
 ===================Details of the http request header:============
 HTTP/1.1 200 OK
 Transfer-Encoding: chunked
 Content-Type: text/plain; charset=utf-8
 Date: Fri, 03 Mar 2023 13:03:03 GMT
 Server: envoy
 X-Envoy-Upstream-Service-Time: 101
 
 #多次访问service0:
 for i in {1..100}; do curl 192.168.91.161/service0; done
 ```



6.通过jaeger dashboard查看tracing

```shell
root@master:~# istioctl dashboard jaeger
http://localhost:16686
```



7.实现istio的安全通信 
```shell
#生成身份认证信息
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=mango Inc./CN=*.mango.io' -keyout mango.io.key -out mango.io.crt
kubectl create -n istio-system secret tls mango-credential --key=mango.io.key --cert=mango.io.crt
```

```yaml
VitrualService:
...
apiVersion: networking.istio.io/v1beta1
kind: VirtualService
metadata:
  name: service0
spec:
  gateways:
    - service0
  hosts:
    - httpsserver.mango.io     
  http:
  - match:
      - port: 443
    route:
      - destination:
          host: service0.tracing.svc.cluster.local
          port:
            number: 80
...
GateWays:
...
apiVersion: networking.istio.io/v1beta1
kind: Gateway
metadata:
  name: service0
spec:
  selector:
    istio: ingressgateway
  servers:
    - hosts:
        - httpsserver.mango.io
      port:
        name: https-default
        number: 443
        protocol: HTTPS
      tls:
        mode: SIMPLE
        credentialName: mango-credential         
```

```shell
kubectl apply -f istio-specs-https.yaml -n tracing
#ingressgateway的IP:
INGERSS_IP=192.168.91.161
#访问： 
curl --resolve httpsserver.mango.io:443:$INGRESS_IP https://httpsserver.mango.io/service0 -v -k



===================Details of the http request header:============
HTTP/1.1 200 OK
Content-Length: 669
Content-Type: text/plain; charset=utf-8
Date: Sat, 04 Mar 2023 13:17:30 GMT
Server: envoy
X-Envoy-Upstream-Service-Time: 19

===================Details of the http request header:============
X-B3-Spanid=[c380d997e48a7cdb]
X-B3-Parentspanid=[180f4df5d750408c]
X-Forwarded-Client-Cert=[By=spiffe://cluster.local/ns/tracing/sa/default;Hash=ec09b0560b07841c3f6773d2a0c4243dcfebb5e7ef74a35ad357a848a11d81b6;Subject="";URI=spiffe://cluster.local/ns/tracing/sa/default]
X-Request-Id=[3a1a8682-b97e-9816-bec3-cb9dc4f3510a]
X-Envoy-Internal=[true]
User-Agent=[Go-http-client/1.1,Go-http-client/1.1,curl/7.68.0]
X-Forwarded-For=[192.168.91.131]
Accept-Encoding=[gzip,gzip]
X-Envoy-Attempt-Count=[1]
X-Forwarded-Proto=[https]
X-B3-Traceid=[27f60dce78559c12ee8fe2afa66a7561]
X-B3-Sampled=[1]
Accept=[*/*]
```



