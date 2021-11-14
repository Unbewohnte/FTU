.DEFAULT_GOAL := all

SRC_DIR := src/
EXE_NAME := ftu
INSTALLATION_DIR := /usr/local/bin/

all:
	cd $(SRC_DIR) && go build && mv $(EXE_NAME) ..

cross:
	cd $(SRC_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ftu_linux_amd64 && mv ftu_linux_amd64 ..
	cd $(SRC_DIR) && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ftu_darwin_amd64 && mv ftu_darwin_amd64 ..
	cd $(SRC_DIR) && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ftu_windows_amd64.exe && mv ftu_windows_amd64.exe ..

race:
	cd $(SRC_DIR) && go build -race && mv $(EXE_NAME) ..

install: all
	cp $(EXE_NAME) $(INSTALLATION_DIR)

test:
	cd $(SRC_DIR) && go test ./... ; cd ..

clean:
	rm $(EXE_NAME)