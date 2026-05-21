
# Utilisation de Docker Compose
 
Simulation de N serveurs et M clients sur un réseau Docker interne.
 
## Prérequis
 
- Docker
- Docker Compose
## Utilisation
 
### 1. Générer la configuration
 
```bash
./generate-compose.sh <N> <M>
```
 
- `N` : nombre de serveurs
- `M` : nombre de clients

Exemple :
```bash
./generate-compose.sh 3 6     # 3 serveurs, 6 clients
```
 
### 2. Lancer les conteneurs
 
Lancement + compilation du code :
```bash
docker compose up --build -d
```
 
Lancer sans recompiler :
```bash
docker compose up -d
```
 
### 3. Voir les logs
 
Tous les conteneurs en temps réel :
```bash
docker compose logs -f
```
 
Un seul conteneur :
```bash
docker compose logs -f serveur0
docker compose logs -f client0
```
 
### 4. Interagir avec un client
 
```bash
docker attach client0
```
 
Pour se détacher sans tuer le client : `Ctrl+P` puis `Ctrl+Q`
 
### 5. Arrêter
 
```bash
docker compose down
```
 
## Détails

- Les serveurs démarrent avec un décalage de 2s (sauf serveur0) pour laisser le cluster se former
- Les clients démarrent avec un décalage de 4s pour attendre que tous les serveurs soient prêts
- Le réseau interne Docker est `172.20.0.0/24`, les conteneurs ne sont pas accessibles depuis l'extérieur
