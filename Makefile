<<<<<<< HEAD
GOOS ?= $(shell go env GOOS)
EXT =
ifeq ($(GOOS),windows)
EXT = .exe
endif

all: serveur client

serveur:
	go build -o serveur$(EXT) ./Serveur

client:
	go build -o client$(EXT) ./Clients

clean:
	rm -f serveur$(EXT) client$(EXT) serveur client
# Variables
CC = gcc
CFLAGS = -Wall -Wextra -O2
SRC = $(wildcard src/*.c)
OBJ = $(SRC:.c=.o)
TARGET = nvProjet

# Règles
all: $(TARGET)

$(TARGET): $(OBJ)
	$(CC) $(CFLAGS) -o $@ $^

src/%.o: src/%.c
	$(CC) $(CFLAGS) -c $< -o $@

clean:
	rm -f src/*.o $(TARGET)

.PHONY: all clean
=======
CC=g++
CFLAGS=-Wall

APP=http_server

all: $(APP).o

$(APP).o: $(APP).cpp
	$(CC) -o $@ $< $(CFLAGS)

clean:
	rm -f *.o
>>>>>>> 02aa59b45f6d58927d32f52e7afa71c28a169a71
