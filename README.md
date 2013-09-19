Vandal
======

http://eatponies.com source code

Install dependencies
--------------------

 * mandatory:

    ```shell
    apt-get install libcairo2-dev
    go get github.com/garyburd/go-websocket/websocket github.com/ugorji/go/codec github.com/zvin/gocairo
    go install github.com/garyburd/go-websocket/websocket github.com/ugorji/go/codec github.com/zvin/gocairo
    ```

 * optionnal (for javascript minification):

    ```shell
    wget http://closure-compiler.googlecode.com/files/compiler-latest.zip
    unzip compiler-latest.zip
    ```

Build
-----

 * put the domain you are going to use into the DOMAIN file (defaults to localhost)

 * put the http port you are going to use into the HTTP_PORT file (defaults to 8000)

 * put the https port you are going to use into the HTTPS_PORT file (defaults to 4430)

 * set the content of the DEBUG file to 'true' or 'false' (true if you want no minification)

 * make:

    ```shell
    make
    ```

Run the server
--------------

 * generate a certificate and a key for https

    ```shell
    openssl req -x509 -nodes -days 365 -newkey rsa:2048 -keyout build/key -out build/cert
    ```

 * run the server

    ```shell
    cd build
    ./vandal -p 8000 -sp 4430 -cert /path/to/certfile -key /path/to/keyfile -f
    ```

 * open your browser at http://localhost:8000
 * -f means print the logs on stdout
 * -p HTTP_PORT
 * -sp HTTPS_PORT (https is needed to be able to draw over https websites)
 * -cert defaults to "cert"
 * -key defaults to "key"

License
-------

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but
WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the GNU
Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public
License along with this program, in the file "COPYING".  If not, see
<http://www.gnu.org/licenses/>.

js/jscolor.js is available under the GNU LESSER GENERAL PUBLIC LICENSE.

js/msgpack.codec.js is available under the MIT LICENSE.

static/LifeSavers-Regular.woff is available under the SIL Open Font License (OFL)
