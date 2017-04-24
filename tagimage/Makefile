all: compile docker push clean

compile:
	CGO_ENABLED=0 go build -o tagimage

docker:
	sudo docker build --no-cache=true --force-rm=true -t dwhitena/tagimage .

push:
	sudo docker push dwhitena/tagimage

clean:
	rm tagimage
