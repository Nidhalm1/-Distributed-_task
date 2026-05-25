#!/bin/bash

# génère le fichier docker-compose.yml (fichier de configuration de docker compose)

N=${1:-3}  # nombre de serveurs
M=${2:-3}  # nombre de clients
BASE_PORT=1234 # port des serveurs
BASE_ALT=7941
SUBNET="172.20.0" # radical de l'ip des serveurs
SERVER0_IP="$SUBNET.10" # serveur du serveur 0
REF="$SERVER0_IP:$BASE_ALT" # adresse du cluster sur lequel les serveurs se connectent

echo "services:" > docker-compose.yml

# Serveurs
for i in $(seq 0 $((N - 1))); do
  ALT=$((BASE_ALT + i))
  IP="$SUBNET.$((10 + i))"
  echo "  serveur$i:" >> docker-compose.yml
  echo "    build:" >> docker-compose.yml
  echo "      context: ." >> docker-compose.yml
  echo "      dockerfile: Dockerfile.serveur" >> docker-compose.yml
  echo "    image: serveur$i" >> docker-compose.yml
  echo "    container_name: serveur$i" >> docker-compose.yml
  if [ $i -eq 0 ]; then
    echo "    command: [\"/app/serveur\", \"$ALT\", \"$BASE_PORT\", \"$REF\"]" >> docker-compose.yml
  else
    echo "    command: [\"sh\", \"-c\", \"sleep 2 && /app/serveur $ALT $BASE_PORT $REF\"]" >> docker-compose.yml
  fi
  echo "    networks:" >> docker-compose.yml
  echo "      mynet:" >> docker-compose.yml
  echo "        ipv4_address: $IP" >> docker-compose.yml
  echo "" >> docker-compose.yml
done

# Clients
for j in $(seq 0 $((M - 1))); do
  TARGET_IP="$SUBNET.$((10 + (j % N)))"
  echo "  client$j:" >> docker-compose.yml
  echo "    build:" >> docker-compose.yml
  echo "      context: ." >> docker-compose.yml
  echo "      dockerfile: Dockerfile.client" >> docker-compose.yml
  echo "    image: client$j" >> docker-compose.yml
  echo "    container_name: client$j" >> docker-compose.yml
  echo "    stdin_open: true" >> docker-compose.yml
  echo "    tty: true" >> docker-compose.yml
  echo "    command: [\"sh\", \"-c\", \"sleep 4 && /app/client $TARGET_IP:$BASE_PORT\"]" >> docker-compose.yml
  echo "    networks:" >> docker-compose.yml
  echo "      mynet:" >> docker-compose.yml
  echo "    depends_on:" >> docker-compose.yml
  for i in $(seq 0 $((N - 1))); do
    echo "      serveur$i:" >> docker-compose.yml
    echo "        condition: service_started" >> docker-compose.yml
  done
  echo "" >> docker-compose.yml
done

# Réseau
echo "networks:" >> docker-compose.yml
echo "  mynet:" >> docker-compose.yml
echo "    driver: bridge" >> docker-compose.yml
echo "    ipam:" >> docker-compose.yml
echo "      config:" >> docker-compose.yml
echo "        - subnet: $SUBNET.0/24" >> docker-compose.yml

echo "docker-compose.yml généré avec $N serveurs et $M clients"
