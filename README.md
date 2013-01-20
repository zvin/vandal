Vandal
======

eatponies.com source code

Install dependencies
--------------------

 * mandatory:

    ```shell
    apt-get install libcairo2-dev
    go get code.google.com/p/go.net/websocket github.com/ugorji/go-msgpack github.com/zvin/gocairo
    go install code.google.com/p/go.net/websocket github.com/ugorji/go-msgpack github.com/zvin/gocairo
    ```

 * optionnal (for javascript minification):

    ```shell
    wget http://closure-compiler.googlecode.com/files/compiler-latest.zip
    unzip compiler-latest.zip
    ```

Build
-----

 * set the DOMAIN and make:

    ```shell
    DOMAIN=localhost:8000 make
    ```

 * if you do not want the javascript files to be minified, set DEBUG to true:

    ```shell
    DOMAIN=localhost:8000 DEBUG=true make
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
