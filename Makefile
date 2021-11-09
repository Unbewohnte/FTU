.DEFAULT_GOAL := all

SRC_DIR := src/
EXE_NAME := ftu
INSTALLATION_DIR := /usr/local/bin/

all:
	cd $(SRC_DIR) && go build && mv $(EXE_NAME) ..

race:
	cd $(SRC_DIR) && go build -race && mv $(EXE_NAME) ..

install: all
	cp $(EXE_NAME) $(INSTALLATION_DIR)

test:
	cd $(SRC_DIR) && go test ./... ; cd ..

clean:
	rm $(EXE_NAME)