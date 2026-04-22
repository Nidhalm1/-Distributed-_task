CC=g++
CFLAGS=-Wall

APP=cpu_charge
APP2=exec_cmd

all: $(APP).o

$(APP).o: $(APP).cpp
	$(CC) -o $@ $< $(CFLAGS)

clean:
	rm -f *.o