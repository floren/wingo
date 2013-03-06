#!/bin/bash

source env.bash

# build packages
echo BUILD PACKAGES
for i in `ls src`
do
	echo $i
	go install $i
done
echo

