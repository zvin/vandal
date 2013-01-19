#!/bin/bash

DOMAIN="eatponies.com"

which java > /dev/null
if [ "$1" != "debug" ] && [ $? = 0 ] && [ -e compiler.jar ]; then
    echo "okay"
    java -jar compiler.jar --js js/jscolor.js --js js/msgpack.js --js js/script.js --compilation_level SIMPLE_OPTIMIZATIONS --js_output_file static/script.js
else
    echo "You sould get google closure-compiler from http://code.google.com/p/closure-compiler/"
    echo "and put compiler.jar in this folder."
    echo "I will only concatenate script files (no minifyingg)."
    cat js/jscolor.js js/msgpack.js js/script.js > static/script.js
fi

sed -i 's/DOMAIN_PLACEHOLDER/'$DOMAIN'/g' static/script.js
