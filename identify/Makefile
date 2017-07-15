all: compile docker push clean

compile:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o goidentify

docker:
	sudo docker build --force-rm=true -t dwhitena/fbidentify .

push:
	sudo docker push dwhitena/fbidentify

clean:
	rm goidentify
