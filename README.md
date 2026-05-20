## Compilation des programmes

Compilez d'abord vos deux programmes pour générer les exécutables :

```bash
make
```

---

## 2. Lancement du Cluster (Les Serveurs)

Ouvrez plusieurs terminaux pour simuler votre cluster.

**Terminal 1 : Lancement du Premier Nœud (Dispatcher)**  
Ce nœud va créer le réseau (sur le port `7941`) et écouter les requêtes du client (sur le port `1234`).

```bash
./serveur.exe 7941 1234
```

**Terminal 2 : Lancement d'un Nœud Worker**  
Ce nœud (port `7942`) rejoint le cluster en contactant le premier nœud (`127.0.0.1:7941`). Il écoutera d'éventuels autres clients sur le port `1235`.

```bash
./serveur.exe 7942 1235 127.0.0.1:7941
```

> **Note :** L'affichage `client disconnected or decode error: EOF` dans les logs des serveurs est normal, c'est simplement la fin de la connexion rapide après les "probes" de vérification.

---

## 3. Utilisation du Client

Ouvrez un 3ème terminal pour lancer le client et connectez-le au premier serveur (port `1234`) :

```bash
./client.exe localhost:1234
```

Une fois dans le menu `>`, vous pouvez soumettre une commande et interroger son résultat.

### Étape A : Soumettre une tâche

Indiquez l'estimation CPU et RAM, puis la commande.  
Par exemple, pour lister les fichiers (`ls -l`) :

```
> submit cpu=10 mem=15 ls -l
ID recu 48db280b-8bcd-48a5-8aaf-9
```

Le dispatcher va automatiquement trouver un nœud disponible (ex : le nœud `7942`) et lui envoyer la tâche.

### Étape B : Récupérer le résultat

Utilisez la commande `result` suivie de l'ID généré à l'étape précédente pour voir le retour de votre commande :

```
> result 48db280b-8bcd-48a5-8aaf-9
etat ID recu total 26000
drwxrwxrwx 1 nmoussa nmoussa      Clients
-rwxrwxrwx 1 nmoussa nmoussa      Makefile
-rwxrwxrwx 1 nmoussa nmoussa      README.md
...
```

Pour quitter le client :

```
> exit
```