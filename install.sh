#!/bin/bash

EXE_NAME=ftu
INSTALLATION_DIR=/usr/local/bin/


if [ -f $EXE_NAME ]; then
    cp $EXE_NAME $INSTALLATION_DIR
else
    echo "No $EXE_NAME found in current directory"
fi