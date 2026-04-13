CC=g++
CFLAGS=-Wall

APP=http_server

all: $(APP).o

$(APP).o: $(APP).cpp
	$(CC) -o $@ $< $(CFLAGS)

clean:
	rm -f *.o