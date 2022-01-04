.DEFAULT_GOAL := all

SRC_DIR := src/
EXE_NAME := ftu
INSTALLATION_DIR := /usr/local/bin/
RELEASE_DIR := release
LICENSE_FILE := COPYING
INSTALLATION_SCRIPT := install.sh

all:
	cd $(SRC_DIR) && go build && mv $(EXE_NAME) ..

release:
	rm -rf $(RELEASE_DIR)

	mkdir $(RELEASE_DIR)
	mkdir $(RELEASE_DIR)/ftu_linux_amd64
	mkdir $(RELEASE_DIR)/ftu_darwin_amd64
	mkdir $(RELEASE_DIR)/ftu_windows_amd64

	cd $(SRC_DIR) && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ftu && mv ftu ../$(RELEASE_DIR)/ftu_linux_amd64
	cd $(SRC_DIR) && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o ftu && mv ftu ../$(RELEASE_DIR)/ftu_darwin_amd64
	cd $(SRC_DIR) && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o ftu.exe && mv ftu.exe ../$(RELEASE_DIR)/ftu_windows_amd64

	cp $(LICENSE_FILE) $(RELEASE_DIR)/ftu_linux_amd64
	cp $(INSTALLATION_SCRIPT) $(RELEASE_DIR)/ftu_linux_amd64
	
	cp $(LICENSE_FILE) $(RELEASE_DIR)/ftu_darwin_amd64
	cp $(LICENSE_FILE) $(RELEASE_DIR)/ftu_windows_amd64

	cd $(RELEASE_DIR) && zip -r ftu_linux_amd64 ftu_linux_amd64/
	cd $(RELEASE_DIR) && zip -r ftu_darwin_amd64 ftu_darwin_amd64/
	cd $(RELEASE_DIR) && zip -r ftu_windows_amd64 ftu_windows_amd64/

	rm -rf $(RELEASE_DIR)/ftu_linux_amd64
	rm -rf $(RELEASE_DIR)/ftu_darwin_amd64
	rm -rf $(RELEASE_DIR)/ftu_windows_amd64

race:
	cd $(SRC_DIR) && go build -race && mv $(EXE_NAME) ..

install: all
	cp $(EXE_NAME) $(INSTALLATION_DIR)

test:
	cd $(SRC_DIR) && go test ./...

clean:
	rm -rf $(EXE_NAME) $(RELEASE_DIR)