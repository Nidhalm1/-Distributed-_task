EXT := .exe

all: serveur client

serveur:
	go build -o serveur$(EXT) ./Serveur

client:
	go build -o client$(EXT) ./Clients

clean:
	del /Q serveur$(EXT) client$(EXT)

.PHONY: all serveur client clean