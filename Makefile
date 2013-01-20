DOMAIN?=localhost:8000
DEBUG?=false

CLOSURECOMPILER=compiler.jar
ifeq ($(wildcard $(CLOSURECOMPILER)),)
	DEBUG=true
endif
GOOUT=vandal
GOFILES=eventtype.go location.go main.go user.go
JSOUT=static/script.js
JSFILES=js/jscolor.js js/msgpack.js js/script.js
HTMLOUT=templates/index.html
HTMLFILES=$(HTMLOUT).orig

all: $(GOOUT) $(JSOUT) $(HTMLOUT)

$(GOOUT): $(GOFILES)
	go build -o $(GOOUT) $(GOFILES)

.PHONY: $(HTMLOUT) $(JSOUT) mrproper

$(HTMLOUT): $(HTMLFILES)
	sed 's/DOMAIN_PLACEHOLDER/'$(DOMAIN)'/g' $(HTMLFILES) > $(HTMLOUT)

$(JSOUT): $(JSFILES)
ifeq ($(DEBUG),true)
	cat $(JSFILES) > $(JSOUT)
	echo "You sould get google closure-compiler from http://code.google.com/p/closure-compiler/"
	echo "and put compiler.jar in this folder."
	echo "I will only concatenate script files (no minifyingg)."
else
	java -jar $(CLOSURECOMPILER) --js $(JSFILES) --compilation_level SIMPLE_OPTIMIZATIONS --js_output_file $(JSOUT)
endif
	sed -i 's/DOMAIN_PLACEHOLDER/'$(DOMAIN)'/g' $(JSOUT)

mrproper:
	rm $(JSOUT) $(HTMLOUT) $(GOOUT)
