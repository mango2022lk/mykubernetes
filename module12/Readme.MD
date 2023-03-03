作业描述：
将httpserver以istio ingress gateway的形式发布:
1.实现安全保证
2.七层路由规则
3.考虑open tracing的接入


1.安装istio:
wget https://github.com/istio/istio/releases/download/1.12.0/istio-1.12.0-linux-amd64.tar.gz
tar -zxvf  istio-1.12.0-linux-amd64.tar.gz
cd istio-1.12.0
cp bin/istioctl /usr/local/bin
istioctl install --set profile=demo -y

查看pod:
root@master:~# k get pod -n istio-system
NAME                                   READY   STATUS    RESTARTS      AGE
istio-egressgateway-7f4864f59c-sk8c4   1/1     Running   2 (41m ago)   2d21h
istio-ingressgateway-55d9fb9f-w87rz    1/1     Running   2 (41m ago)   2d21h
istiod-5d5b675975-jx5tk                1/1     Running   2 (41m ago)   2d21h

2.安装jaeger，用来抓取tracing
kubectl apply -f jaeger.yaml

3.部署服务
kubectl create ns tracing
#为tracing下的pod注入sidecar 
kubectl label ns tracing istio-injection=enabled
kubectl -n tracing apply -f service0.yaml
kubectl -n tracing apply -f service1.yaml
kubectl -n tracing apply -f service2.yaml
kubectl apply -f istio-specs.yaml -n tracing

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
k get svc -n istio-system
root@master:~# k get svc -n istio-system
NAME                   TYPE           CLUSTER-IP       EXTERNAL-IP      PORT(S)                                                                      AGE
istio-egressgateway    ClusterIP      10.101.131.87    <none>           80/TCP,443/TCP                                                               2d21h
istio-ingressgateway   LoadBalancer   10.102.235.84    192.168.91.161   15021:32448/TCP,80:30817/TCP,443:32038/TCP,31400:30240/TCP,15443:31480/TCP

root@master:~# curl 192.168.91.161/service0
===================Details of the http request header:============
HTTP/1.1 200 OK
Transfer-Encoding: chunked
Content-Type: text/plain; charset=utf-8
Date: Fri, 03 Mar 2023 13:03:03 GMT
Server: envoy
X-Envoy-Upstream-Service-Time: 101


5.查看tracing：
多次访问service0:
for i in {1..100}; do curl 192.168.91.161/service0; done

#istio通过jaeger dashboard 提供了tracing的可视化支持
root@master:~# istioctl dashboard jaeger
http://localhost:16686


6.实现istio的安全通信
1)生成安全证书
openssl req -x509 -sha256 -nodes -days 365 -newkey rsa:2048 -subj '/O=cncamp Inc./CN=*.cncamp.io' -keyout cncamp.io.key -out cncamp.io.crt
2)创建对应secret
kubectl create -n istio-system secret tls cncamp-credential --key=cncamp.io.key --cert=cncamp.io.crt
3)修改istio-spec.yaml
```
VitrualService:
...
  - match:
      - uri:
          exact: /service0
      - port: 443
    route:
      - destination:
          host: httpserver.tracing.svc.cluster.local
          port:
            number: 80
...
GateWays:
...
  servers:
    - hosts:
        - httpsserver.cncamp.io
      port:
        name: https-default
        number: 443
        protocol: HTTPS
      tls:
        mode: SIMPLE
        credentialName: cncamp-credential            
```
kubectl apply -f istio-specs-https.yaml -n tracing

访问：
curl --resolve httpsserver.cncamp.io:443:$INGRESS_IP https://httpsserver.cncamp.io/service0 -v -k

