DOMAIN?=localhost:8000
DEBUG?=false

CLOSURECOMPILER=compiler.jar
ifeq ($(wildcard $(CLOSURECOMPILER)),)
	DEBUG=true
endif
GOFILES=eventtype.go location.go main.go user.go
JSFILES=js/jscolor.js js/msgpack.js js/script.js

all: vandal static/script.js templates/index.html

vandal: $(GOFILES)
	go build -o vandal $(GOFILES)

.PHONY: templates/index.html static/script.js mrproper

templates/index.html: templates/index.html.orig
	sed 's/DOMAIN_PLACEHOLDER/'$(DOMAIN)'/g' templates/index.html.orig > templates/index.html

static/script.js: $(JSFILES)
ifeq ($(DEBUG),true)
		cat $(JSFILES) > static/script.js
		echo "You sould get google closure-compiler from http://code.google.com/p/closure-compiler/"
		echo "and put compiler.jar in this folder."
		echo "I will only concatenate script files (no minifyingg)."
else
		java -jar $(CLOSURECOMPILER) --js $(JSFILES) --compilation_level SIMPLE_OPTIMIZATIONS --js_output_file static/script.js
endif
	sed -i 's/DOMAIN_PLACEHOLDER/'$(DOMAIN)'/g' static/script.js

mrproper:
	rm static/script.js templates/index.html vandal
