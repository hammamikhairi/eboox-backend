#!/bin/bash

# Check if both arguments are provided
if [ $# -ne 2 ]; then
    echo "Usage: $0 <command> <book_uuid>"
    exit 1
fi

# Read the arguments
command=$1
target=$2

# Execute the corresponding command
case $command in
    progress)
        curl -X POST -H "Content-Type: application/json" -d "{\"book_uuid\": \"$target\", \"progress\": \"page34\"}" http://localhost:5050/progress
        ;;

    add_bookmark)
        curl -X POST -H "Content-Type: application/json" -d "{\"book_uuid\": \"$target\", \"bookmark\": \"69\", \"action\":\"add\"}" http://localhost:5050/bookmark
        ;;
    remove_bookmark)
        curl -X POST -H "Content-Type: application/json" -d "{\"book_uuid\": \"$target\", \"bookmark\": \"69\", \"action\":\"remove\"}" http://localhost:5050/bookmark
        ;;

    add_highlight)
        curl -X POST -H "Content-Type: application/json" -d "{\"book_uuid\": \"$target\", \"bounds\": \"6:9\", \"content\":\"death to me\", \"action\":\"add\"}" http://localhost:5050/highlight
        ;;
    remove_highlight)
        curl -X POST -H "Content-Type: application/json" -d "{\"book_uuid\": \"$target\", \"bounds\": \"6:9\", \"action\":\"remove\"}" http://localhost:5050/highlight
        ;;

    add_note)
        curl -X POST -H "Content-Type: application/json" -d "{\"book_uuid\": \"$target\", \"highlight\": \"6:9\", \"content\":\"This is an awesome quote\", \"action\":\"add\"}" http://localhost:5050/note
        ;;
    modify_note)
        curl -X POST -H "Content-Type: application/json" -d "{\"book_uuid\": \"$target\", \"highlight\": \"6:9\", \"content\":\"Oh man!\", \"action\":\"update\"}" http://localhost:5050/note
        ;;
    remove_note)
        curl -X POST -H "Content-Type: application/json" -d "{\"book_uuid\": \"$target\", \"highlight\": \"6:9\", \"action\":\"remove\"}" http://localhost:5050/note
        ;;
    *)
        echo "Invalid command: $command"
        ;;
esac
