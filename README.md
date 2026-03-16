# btop

Un utilitaire CLI léger en Go pour surveiller les processus du système en temps réel et identifier rapidement ceux qui consomment le plus de CPU et de RAM.

## 🚀 Installation

### Option 1 : Compiler localement
```sh
go mod tidy
go build -o btop
./btop
```

### Option 2 : Installation globale
Si votre environnement Go est bien configuré (avec `$GOPATH/bin` dans votre `$PATH`), vous pouvez l'installer globalement sur votre système :

```sh
go install
```
Vous pourrez ensuite l'utiliser de n'importe où en tapant `btop`.

## 🛠 Commandes & Utilisation

### `btop`
Affiche un tableau en temps réel des processus les plus gourmands (rafraîchi toutes les 2 secondes), ainsi qu'un résumé de la charge CPU et RAM globale du système.

**Options :**
- `--limit` : Nombre de processus à afficher (défaut : 10)
- `--interval` : Intervalle de rafraîchissement en secondes (défaut : 2)
- `--sort` : Champ de tri : `cpu`, `ram`, `mem`, `pid`, `name` (défaut : `cpu`)
- `--name` : Filtrer la liste par nom de processus (ex: `--name chrome`)
- `--user` : Filtrer la liste par propriétaire du processus

Exemple : `btop --interval 1 --limit 15 --sort ram --name firefox`

### `btop kill <pid>`
Tuer un processus avec son PID.

### `btop watch`
Surveille les processus en arrière-plan et affiche une alerte si l'un d'eux dépasse 80% (CPU ou RAM).

**Options :**
- `--auto-kill` : Tue automatiquement les processus déviants.
- `--interval` : Fréquence de vérification.

Exemple : `btop watch --auto-kill`

## 🧰 Informations Techniques
- **Langage**: Go
- **Packages principaux**: Cobra (CLI), Gopsutil (supervision), Fatih Color (couleurs terminal).
# btop
