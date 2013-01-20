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

 * put the domain you are going to use into the DOMAIN file (defaults to localhost:8000)

 * set the content of the DEBUG file to 'true' or 'false' (true if you want no minification)

 * make:

    ```shell
    make
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
