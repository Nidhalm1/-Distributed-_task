#!/bin/bash

NODE_PORT=$1
NODE_ID=$2

echo "Lancement du noeud..."
./serveur.exe $NODE_PORT $NODE_ID &

SERVER_PID=$!

sleep 2

echo "Lancement du collector..."
./utils/collector "C:\Users\Mo\AppData\Local\Temp\cpu.sock" &

COLLECTOR_PID=$!

echo "Serveur PID : $SERVER_PID"
echo "Collector PID : $COLLECTOR_PID"

wait