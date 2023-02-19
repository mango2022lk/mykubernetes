FROM ubuntu
RUN apt update && apt install -y net-tools \
	vim \
	curl
COPY  httpserver /
COPY  httpserver.properties /etc/httpserver/httpserver.properties
EXPOSE 8080
ENTRYPOINT ["/httpserver"]