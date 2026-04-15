#!/bin/bash

echo "🚀 Début de l'installation de killerprocess..."

# 1. Vérifier si Go est installé
if ! command -v go &> /dev/null
then
    echo "❌ Erreur : Le langage externe 'Go' n'est pas installé sur votre ordinateur."
    echo "Pour utiliser ce programme source, vous devez installer Go."
    echo "👉 Allez sur : https://go.dev/doc/install et suivez les instructions pour votre système, puis relancez ce script."
    exit 1
fi

# 2. Télécharger les dépendances automatiquement
echo "📦 Téléchargement des dépendances..."
go mod tidy

# 3. Compiler le code en un seul fichier exécutable
echo "⚙️ Compilation du programme..."
go build -o killerprocess

# 4. Vérifier que la compilation a réussi
if [ ! -f ./killerprocess ]; then
    echo "❌ Erreur : La compilation a échoué."
    exit 1
fi

# 5. Installer globalement dans le système pour que 'killerprocess' soit une commande reconnue partout
echo "🔑 Installation globale sur le système..."
echo "(Il est possible qu'on vous demande votre mot de passe pour le copier dans le dossier système)"

# Nettoyer l'ancienne version appelée 'btop' si elle existe
if [ -f /usr/local/bin/btop ]; then
    echo "🧹 Suppression automatique de l'ancienne version (btop)..."
    sudo rm -f /usr/local/bin/btop
fi

sudo mv killerprocess /usr/local/bin/

echo ""
echo "✅ SUCCÈS ! killerprocess est maintenant installé et reconnu comme une commande de votre système."
echo "🎉 Vous pouvez fermer ce terminal, en ouvrir un nouveau, et simplement taper :"
echo "   killerprocess"
echo ""
