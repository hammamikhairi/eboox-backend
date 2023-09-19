#!/bin/bash

c="go"

build="go build -ldflags=-s ."
bin="eboox"

color="\e[93m"
reset="\e[0m"

function build_and_exec() {
    eval $build
    eval "./$bin" &
}

function kill_process() {
    eval "pkill -2 $bin" > /dev/null 2>&1
    wait $!
    echo
}

function wrap_echo() {
    echo -e $color$1$reset
}

wrap_echo "[INFO] starting project"
build_and_exec

while true; do
    read -rsn1 input
    echo "input $input"
    if [ "$input" == "r" ]; then
        wrap_echo "[INFO] restarting project"
        wrap_echo "[INFO] killing process"
        kill_process
        wrap_echo "[INFO] building and executing binary"
        build_and_exec
    elif [ "$input" == "c" ]; then
        clear
        wrap_echo "[INFO] restarting project"
        wrap_echo "[INFO] killing process"
        kill_process
        wrap_echo "[INFO] building and executing binary"
        build_and_exec
    elif [ "$input" == "x" ]; then
        wrap_echo "[INFO] killing processes and exiting"
        kill_process
        break 
    fi
done

wrap_echo "[INFO] exited"