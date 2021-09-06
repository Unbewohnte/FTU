#!/bin/bash

EXECUTABLE_NAME=ftu
DESTDIR=/usr/bin/

# if ftu is in the same directory - copy it to $DESTDIR
if [ -e $EXECUTABLE_NAME ]
then
    cp $EXECUTABLE_NAME $DESTDIR
else
    echo "No '${EXECUTABLE_NAME}' in current directory !"
fi