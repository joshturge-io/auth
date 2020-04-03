BUILD_DIR=bin/
EXENAME=auth

GO=$(shell which go)

# Check OS 
UNAME := $(shell uname)
ifeq ($(UNAME), Darwin)
       GOOS=darwin
else
    ifeq ($(UNAME), Linux)
        GOOS=linux
    endif
endif

protoc-check:
	./scripts/protobuf-check.sh

protobuf: protoc-check
	./scripts/protoc_gen.sh

# Build the project
build: protobuf
	 CGO_ENABLED=0 GOOS=$(GOOS) GO111MODULES=auto $(GO) build -a -installsuffix cgo -o \
				 $(BUILD_DIR)$(EXENAME) cmd/auth/main.go

# Clean out the build dir
clean: $(BUILD_DIR)
	 [ -d $< ] && rm -r $< 

# Make build directory if it doesn't already exist
$(BUILD_DIR):
	[ -d $@ ] || mkdir -p $@

