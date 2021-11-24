.DEFAULT_GOAL := all

SRC_DIR := src/
EXE_NAME := ftu
INSTALLATION_DIR := /usr/local/bin/
RELEASE_DIR := release
LICENSE_FILE := COPYING
INSTALLATION_SCRIPT := install.sh

all:
	cd $(SRC_DIR) && go build && mv $(EXE_NAME) ..

pkgrelease:
	rm -rf $(RELEASE_DIR)

	mkdir $(RELEASE_DIR)
	mkdir $(RELEASE_DIR)/linux_amd64
	mkdir $(RELEASE_DIR)/darwin_amd64
	mkdir $(RELEASE_DIR)/windows_amd64

	cd $(SRC_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ftu && mv ftu ../$(RELEASE_DIR)/linux_amd64
	cd $(SRC_DIR) && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ftu && mv ftu ../$(RELEASE_DIR)/darwin_amd64
	cd $(SRC_DIR) && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ftu.exe && mv ftu.exe ../$(RELEASE_DIR)/windows_amd64

	cp $(LICENSE_FILE) $(RELEASE_DIR)/linux_amd64
	cp $(INSTALLATION_SCRIPT) $(RELEASE_DIR)/linux_amd64
	
	cp $(LICENSE_FILE) $(RELEASE_DIR)/darwin_amd64
	cp $(LICENSE_FILE) $(RELEASE_DIR)/windows_amd64

	cd $(RELEASE_DIR) && zip -r linux_amd64 linux_amd64/
	cd $(RELEASE_DIR) && zip -r darwin_amd64 darwin_amd64/
	cd $(RELEASE_DIR) && zip -r windows_amd64 windows_amd64/

	rm -rf $(RELEASE_DIR)/linux_amd64
	rm -rf $(RELEASE_DIR)/darwin_amd64
	rm -rf $(RELEASE_DIR)/windows_amd64

race:
	cd $(SRC_DIR) && go build -race && mv $(EXE_NAME) ..

install: all
	cp $(EXE_NAME) $(INSTALLATION_DIR)

test:
	cd $(SRC_DIR) && go test ./...

clean:
	rm -rf $(EXE_NAME) $(RELEASE_DIR)