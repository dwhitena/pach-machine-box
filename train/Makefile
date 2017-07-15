all: compile docker push clean

compile:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o gotrain

docker:
	sudo docker build --force-rm=true -t dwhitena/fbtrain .

push:
	sudo docker push dwhitena/fbtrain

clean:
	rm gotrain
