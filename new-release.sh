#!/bin/bash

# Script per automatizzare il rilascio di una nuova versione su GitHub.
# Utilizzo: ./new-release.sh <versione>
# Esempio: ./new-release.sh 1.2.0
# Nota: Ricordati di aggiornare il file CHANGELOG.md prima di lanciare lo script 
# se vuoi che le modifiche siano tracciate lì, anche se la release su GitHub 
# avrà comunque le note generate automaticamente dai commit.

VERSION=$1

# Verifica che sia stato fornito un numero di versione
if [ -z "$VERSION" ]; then
  echo "Errore: Devi specificare una versione."
  echo "Utilizzo: $0 <versione>"
  echo "Esempio: $0 1.2.0"
  exit 1
fi

# Verifica il formato della versione (X.Y.Z)
if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
  echo "Errore: La versione deve essere nel formato X.Y.Z (es. 1.2.0)"
  exit 1
fi

echo "=== Inizio procedura di rilascio per la versione $VERSION ==="

# 1. Aggiorna la versione in build_macos.sh
# Cerca "CFBundleShortVersionString" e sostituisce la riga successiva con la nuova versione
sed -i '' "/CFBundleShortVersionString/{n;s/<string>.*<\/string>/<string>$VERSION<\/string>/;}" build_macos.sh

if [ $? -ne 0 ]; then
    echo "Errore: Impossibile aggiornare la versione in build_macos.sh"
    exit 1
fi
echo "✔ Versione aggiornata in build_macos.sh"

# 2. Esegui lo script di build
echo "--- Esecuzione build_macos.sh ---"
./build_macos.sh
if [ $? -ne 0 ]; then
    echo "Errore: La build è fallita."
    exit 1
fi
echo "✔ Build completata con successo"

# 3. Comprimi l'applicazione
echo "--- Creazione archivio ZIP ---"
ZIP_PATH="build/CriptoMenu.app.zip"
# Rimuovi zip precedente se esiste per sicurezza
rm -f "$ZIP_PATH"
zip -r "$ZIP_PATH" "build/CriptoMenu.app" > /dev/null
if [ $? -ne 0 ]; then
    echo "Errore: Creazione zip fallita."
    exit 1
fi
echo "✔ Archivio creato: $ZIP_PATH"

# 4. Operazioni Git
echo "--- Operazioni Git ---"

# Controlla se ci sono cambiamenti in build_macos.sh da committare
if git diff --quiet build_macos.sh; then
    echo "Nessun cambiamento rilevato in build_macos.sh (la versione era già aggiornata?)"
else
    git add build_macos.sh
    git commit -m "Bump version to $VERSION"
    echo "✔ Commit effettuato"
    git push
    echo "✔ Push effettuato"
fi

# Gestione Tag
if git rev-parse "v$VERSION" >/dev/null 2>&1; then
    echo "Attenzione: Il tag v$VERSION esiste già."
else
    git tag "v$VERSION"
    git push origin "v$VERSION"
    echo "✔ Tag v$VERSION creato e pushato"
fi

# 5. Crea Release su GitHub
echo "--- Creazione Release GitHub ---"
# Usa --generate-notes per creare automaticamente le note basate sui PR/commit
gh release create "v$VERSION" "$ZIP_PATH" --title "v$VERSION" --generate-notes

if [ $? -eq 0 ]; then
    echo "=== Rilascio v$VERSION completato con successo! ==="
else
    echo "Errore: Creazione release GitHub fallita."
    exit 1
fi
