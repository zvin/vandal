vandal
======

eatponies.com source code

BUILD THE SERVER
================

 * install dependencies:

    apt-get install libcairo2-dev
    go get code.google.com/p/go.net/websocket github.com/ugorji/go-msgpack github.com/zvin/gocairo
    go install code.google.com/p/go.net/websocket github.com/ugorji/go-msgpack github.com/zvin/gocairo

 * build:

    go build -o vandal *.go


CREATE JAVASCRIPT FILE
======================

 * get closure compiler (optionnal, for minification):

    wget http://closure-compiler.googlecode.com/files/compiler-latest.zip
    unzip compiler-latest.zip

 * replace "eatponies.com" with your domain in concat_scripts.sh and template/index.html:

    sed -i 's/eatponies.com/localhost:8000/g' templates/index.html concat_scripts.sh

 * concatenate scripts:

    ./concat_scripts.sh


RUN THE SERVER
==============

 * create necessary folders:

    mkdir log img

 * run the server

    ./vandal -p 8000

 * open your browser at http://localhost:8000
