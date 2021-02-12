#!/bin/bash
start() {
    docker run \
    -d \
    --name go-playground-sandbox \
    -p 127.0.0.1:8080:8080 \
    --rm \
    xiam/go-playground-sandbox

    # Running unsafebox
    # docker run \
    #  -d \
    #  --name go-playground-unsafebox \
    #  -p 127.0.0.1:8080:8080 \
    #  xiam/go-playground-unsafebox

    # Running web editor
    docker run \
    -d \
    --link go-playground-sandbox:compiler \
    --name go-playground \
    -p 0.0.0.0:3000:3000 \
    --rm \
    xiam/go-playground \
        bash -c \
        'webapp -c http://compiler:8080/compile?output=json -allow-share'
}

stop() {
    docker stop go-playground
    docker stop go-playground-sandbox
}


case "$1" in
        start)
                $1
                ;;
        stop)
                stop
                ;;
        *)
                echo "Usage: $0 {start|stop}"
                exit 1
                ;;
esac
exit $RETVAL 
