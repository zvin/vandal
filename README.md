Vandal
======

eatponies.com source code

Build the server
----------------

 * install dependencies:

    ```shell
    apt-get install libcairo2-dev
    go get code.google.com/p/go.net/websocket github.com/ugorji/go-msgpack github.com/zvin/gocairo
    go install code.google.com/p/go.net/websocket github.com/ugorji/go-msgpack github.com/zvin/gocairo
    ```

 * build:

    ```shell
    go build -o vandal *.go
    ```


Create javascript file
----------------------

 * get closure compiler (optionnal, for minification):

    ```shell
    wget http://closure-compiler.googlecode.com/files/compiler-latest.zip
    unzip compiler-latest.zip
    ```

 * replace "eatponies.com" with your domain in concat_scripts.sh and template/index.html:

    ```shell
    sed -i 's/eatponies.com/localhost:8000/g' templates/index.html concat_scripts.sh
    ```

 * concatenate scripts:

    ```shell
    ./concat_scripts.sh
    ```


Run the server
--------------

 * create necessary folders:

    ```shell
    mkdir log img
    ```

 * run the server

    ```shell
    ./vandal -p 8000
    ```

 * open your browser at http://localhost:8000
