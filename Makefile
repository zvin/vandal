DOMAIN=$(shell cat DOMAIN)
HTTP_PORT=$(shell cat HTTP_PORT)
HTTPS_PORT=$(shell cat HTTPS_PORT)
DEBUG=$(shell cat DEBUG)
COMMIT=$(shell git log -n1 --pretty=format:%H)

dir_guard=@mkdir -p $(@D)  # create the folder of the target if it doesn't exist
CLOSURECOMPILER=compiler.jar
ifeq ($(wildcard $(CLOSURECOMPILER)),)
	DEBUG=true
endif
BUILD=build
GOOUT=$(BUILD)/vandal
GOFILES=eventtype.go location.go main.go user.go messageslog.go utils.go currently_used_sites.go
JSOUT=$(BUILD)/static/script.js
JSFILES=js/jscolor.js js/header.js js/msgpack.codec.js js/user.js js/ui.js js/client.js js/script.js js/footer.js
HTMLOUT=$(BUILD)/templates/index.html
HTMLFILES=templates/index.html

ifeq ($(DEBUG),true)
	GOFLAGS=-race
else
	GOFLAGS=
endif


all: $(GOOUT) $(JSOUT) $(HTMLOUT) staticfiles version

$(GOOUT): $(GOFILES) DEBUG
	$(dir_guard)
	go build $(GOFLAGS) -o $(GOOUT) $(GOFILES)

$(HTMLOUT): $(HTMLFILES) DOMAIN HTTPS_PORT
	$(dir_guard)
	sed 's/DOMAIN_PLACEHOLDER/'$(DOMAIN)'/g' $(HTMLFILES) > $(HTMLOUT)
	sed -i 's/HTTPS_PORT_PLACEHOLDER/'$(HTTPS_PORT)'/g' $(HTMLOUT)

$(JSOUT): $(JSFILES) DOMAIN DEBUG HTTPS_PORT HTTP_PORT
	$(dir_guard)
	cat $(JSFILES) > $(JSOUT)
ifeq ($(DEBUG),true)
	@echo "You sould get google closure-compiler from http://code.google.com/p/closure-compiler/"
	@echo "and put compiler.jar in this folder."
	@echo "I will only concatenate script files (no minifying)."
else
	java -jar $(CLOSURECOMPILER) --js $(JSOUT) --compilation_level SIMPLE_OPTIMIZATIONS --js_output_file $(JSOUT).min
	mv $(JSOUT).min $(JSOUT)
endif
	sed -i 's/DOMAIN_PLACEHOLDER/'$(DOMAIN)'/g' $(JSOUT)
	sed -i 's/HTTP_PORT_PLACEHOLDER/'$(HTTP_PORT)'/g' $(JSOUT)
	sed -i 's/HTTPS_PORT_PLACEHOLDER/'$(HTTPS_PORT)'/g' $(JSOUT)

.PHONY: version staticfiles

staticfiles:
	rsync -r static $(BUILD)/

version:
	sed -i 's/COMMIT_PLACEHOLDER/'$(COMMIT)'/g' $(HTMLOUT)

mrproper:
	rm -rf $(BUILD)
