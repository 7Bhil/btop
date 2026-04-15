# killerprocess 🚀

**killerprocess** est un petit utilitaire simple, rapide et interactif pour surveiller votre ordinateur. Il vous permet de voir instantanément quels programmes consomment le plus de ressources (processeur et mémoire), et de les arrêter d'une simple touche si nécessaire. 

Il a été conçu pour être très facile à lire, avec des couleurs et un système de navigation interactif au clavier.

---

## 📥 Installation facile (Pour tout le monde)

Vous n'avez pas besoin d'être un expert en informatique pour l'installer. Suivez juste ces étapes !

### Étape 1 : Prérequis (Avoir "Go")
Ce programme est écrit dans un langage appelé "Go". Pour le transformer en application, votre ordinateur doit connaître ce langage.
1. Ouvrez votre navigateur et allez sur : [https://go.dev/doc/install](https://go.dev/doc/install).
2. Téléchargez et installez la version correspondant à votre ordinateur (Windows, Mac, ou Linux).

### Étape 2 : Télécharger et installer `killerprocess`
Nous avons créé un petit script pour tout installer automatiquement à votre place. 
1. Ouvrez votre terminal (par exemple en cherchant "Terminal" dans vos applications).
2. Téléchargez le code source sur votre ordinateur (cela va créer un dossier `killerprocess`) :
   ```sh
   git clone https://github.com/7Bhil/killerprocess.git
   ```
3. Entrez dans ce nouveau dossier :
   ```sh
   cd killerprocess
   ```
4. Rendez le script d'installation exécutable et lancez-le :
   ```sh
   chmod +x install.sh
   ./install.sh
   ```
*(Note : Il se peut que l'ordinateur vous demande votre mot de passe administrateur, c'est normal, c'est pour placer la commande `killerprocess` au même endroit que vos autres commandes système).*

✅ **C'est fini !** Vous pouvez maintenant utiliser `killerprocess` n'importe où, n'importe quand !

---

## 🖥 Comment utiliser `killerprocess` ?

Maintenant que le programme est installé, pour le lancer, il vous suffit d'ouvrir un terminal (n'importe lequel) et de taper :
```sh
killerprocess
```

Un écran interactif (comme une vraie application) va s'ouvrir. 

### ⌨️ Les contrôles (très simples)
Tout se fait au clavier, pas besoin de souris :
- **Flèche du Haut `↑`** et **Flèche du Bas `↓`** : Naviguer dans la liste des processus.
- **Touche `x` ou `Entrée`** : Tuer (arrêter de force) le programme que vous avez sélectionné d'un seul coup. *Adieu l'application qui plante !*
- **Touche `<` ou `>`** : Changer l'ordre de la liste. Par exemple, vous pouvez trier du plus gourmand en `RAM` (mémoire) au plus gourmand en `CPU` (processeur).
- **Touche `q` ou `Ctrl+C`** : Quitter élégamment l'application et revenir à votre terminal normal.

---

## 🛠 Pour les utilisateurs avancés (Fonctionnalités "Pro")

Si vous êtes un peu plus à l'aise, la commande `killerprocess` possède des options invisibles bien pratiques :

- `killerprocess --sort ram` : Ouvre directement l'application en triant par l'utilisation de la RAM.
- `killerprocess --name chrome` : Ouvre l'application, mais n'affiche que les processus dont le nom contient "chrome".
- `killerprocess --user admin` : Filtre la liste pour n'afficher que les processus lancés par un utilisateur spécifique.
- `killerprocess watch --auto-kill` : (Commande de surveillance) Ce mode tourne en arrière-plan sans interface et surveille en silence. Si un programme dépasse brutalement 80% d'utilisation, `killerprocess` va s'en rendre compte, vous alerter et le tuer si l'option est activée !

---

## 📦 Informations techniques
- Construit entièrement en **Go**.
- Interface avec `github.com/charmbracelet/bubbletea` (pour l'interactivité 60 FPS fluide).
- Récupération des données avec `github.com/shirou/gopsutil/v3`.
