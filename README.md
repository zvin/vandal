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

 * run the server

    ```shell
    cd build
    ./vandal -p 8000 -f
    ```

 * open your browser at http://localhost:8000
 * -f means print the logs on stdout

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

js/msgpack.js is available under the MIT LICENSE.
