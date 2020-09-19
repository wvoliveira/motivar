default: build

dep:
	go get github.com/markbates/pkger/cmd/pkger

build:
	pkger -include /data && go build
