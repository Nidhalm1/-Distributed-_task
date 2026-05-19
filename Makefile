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
