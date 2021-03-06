CC=g++
GOP = $$HOME/projects/cloud/go

CXXFLAGS=-Wall
LDFLAGS=-Wall

INCPATH=inc
SRCPATH=src
OBJPATH=obj
LIBPATH=lib
BINPATH=bin

#OBJS= $(OBJPATH)/simhash.o $(OBJPATH)/simtable.o $(OBJPATH)/simhashgo_wrap.o
OBJS= $(OBJPATH)/simhash.o $(OBJPATH)/simtable.o
OUT=$(LIBPATH)/libsimhash.so

INCLUDES=-I ./$(INCPATH)

#Set this to your go installation directory
export PATH := bin:$(PATH)

default: $(OUT)

$(OUT): $(OBJS)
	$(CC) $(LDFLAGS) -shared -o $@ $^
	sudo cp $(OUT) /usr/local/lib/

obj/simhashgo_wrap.o: simhashgo_wrap.cxx inc/simtable.h
	$(CC) $(CXXFLAGS) $(INCLUDES) -include ./$(INCPATH)/simtable.h -fpic -c $< -o $@

obj/simtable.o: src/simtable.cpp inc/simhash.h inc/common.h inc/simtable.h
	$(CC) $(CXXFLAGS) $(INCLUDES) -fpic -c $< -o $@

obj/simhash.o: src/simhash.cpp inc/simhash.h inc/common.h
	$(CC) $(CXXFLAGS) $(INCLUDES) -fpic -c $< -o $@

simhashgo_wrap.cxx:
	swig -go -c++ -intgosize 64 -soname libsimhashgo.so simhashgo.swig

.PHONY: clean cleanall

clean:
	rm -f $(OBJPATH)/.o

cleanall: clean
	rm -f $(OUT)
	rm -f .6
	rm -f .a
	rm -f .so
	rm -f .cxx
	rm -f .c

build:
	@echo "Building bindings..."
	go tool 6c -I $(GOP)/pkg/linux_amd64/ -D _64BIT simhashcpp_gc.c
	go tool 6g simhashcpp.go
	go tool pack grc simhashgo.a simhashcpp.6 simhashcpp_gc.6

install:
	@echo "Installing go package..."
	#Rename swig file so go install command does not try to reprocess it
	rm -rf simhashcpp_gc.c simhashgo_wrap.cxx
	mv simhashgo.swig simhashgo.notswig
	go install
	mv simhashgo.notswig simhashgo.swig

	@echo "Installing go shared lib..."
	sudo cp -f $(OUT) /usr/local/lib/
	sudo ldconfig
