dev:
	mkdir -p dbdata
	go run main.go

build:
	go build -o server .

tools:
	go get github.com/coapcloud/coap-cli/cmd@v0.1.1
	mv ${GOPATH}/bin/cmd ${GOPATH}/bin/coap-cli

clean:
	rm -rf ./dbdata

docker-build:
	docker build -t coap-hooks-router .

docker-run:
	docker run -v container-db:/dbdata -e ADMIN_BEARER=${ADMIN_BEARER} -p 8081:8081 -p 5683:5683 coap-hooks-router