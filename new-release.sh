#!/bin/bash

# Script per automatizzare il rilascio di una nuova versione su GitHub.
# Utilizzo: ./new-release.sh <versione>
# Esempio: ./new-release.sh 1.2.0

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

# --- Estrazione note di rilascio ---
echo "--- Estrazione note di rilascio da CHANGELOG.md ---"
RELEASE_NOTES_FILE="release_notes_tmp.md"
# Escape dots for regex
ESCAPED_VERSION=$(echo "$VERSION" | sed 's/\./\\./g')
# Estrae il testo tra l'header della versione corrente e il prossimo header
awk "/^## \[$ESCAPED_VERSION\]/{flag=1; next} /^## [^\/]/{flag=0} flag" CHANGELOG.md > "$RELEASE_NOTES_FILE"

# Verifica se abbiamo trovato delle note (file non vuoto e con caratteri non-whitespace)
if [ ! -s "$RELEASE_NOTES_FILE" ] || ! grep -q "[^[:space:]]" "$RELEASE_NOTES_FILE"; then
    echo "⚠️  Attenzione: Nessuna nota trovata nel CHANGELOG.md per la versione $VERSION."
    echo "    Verranno usate le note generate automaticamente da GitHub."
    USE_GENERATED_NOTES=true
else
    echo "✔ Note di rilascio trovate:"
    cat "$RELEASE_NOTES_FILE"
    USE_GENERATED_NOTES=false
fi

# 1. Aggiorna la versione in build_macos.sh e update.go
echo "--- Aggiornamento versione nei file sorgente ---"

# Aggiorna build_macos.sh
# Cerca "CFBundleShortVersionString" e sostituisce la riga successiva con la nuova versione
sed -i '' "/CFBundleShortVersionString/{n;s/<string>.*<\/string>/<string>$VERSION<\/string>/;}" build_macos.sh
if [ $? -ne 0 ]; then
    echo "Errore: Impossibile aggiornare la versione in build_macos.sh"
    rm -f "$RELEASE_NOTES_FILE"
    exit 1
fi
echo "✔ Versione aggiornata in build_macos.sh"

# Aggiorna update.go
# Cerca CurrentVersion = "..." e lo sostituisce
sed -i '' "s/CurrentVersion = \"[0-9]*\.[0-9]*\.[0-9]*\"/CurrentVersion = \"$VERSION\"/" update.go
if [ $? -ne 0 ]; then
    echo "Errore: Impossibile aggiornare la versione in update.go"
    rm -f "$RELEASE_NOTES_FILE"
    exit 1
fi
echo "✔ Versione aggiornata in update.go"

# 2. Esegui lo script di build
echo "--- Esecuzione build_macos.sh ---"
./build_macos.sh
if [ $? -ne 0 ]; then
    echo "Errore: La build è fallita."
    rm -f "$RELEASE_NOTES_FILE"
    exit 1
fi
echo "✔ Build completata con successo"

# 3. Comprimi l'applicazione
echo "--- Creazione archivio ZIP ---"
ZIP_PATH="build/CriptoMenu.app.zip"
# Rimuovi zip precedente se esiste per sicurezza
rm -f "$ZIP_PATH"

# Spostati nella directory build per zippare solo il contenuto
pushd build > /dev/null
zip -r "CriptoMenu.app.zip" "CriptoMenu.app" > /dev/null
if [ $? -ne 0 ]; then
    echo "Errore: Creazione zip fallita."
    popd > /dev/null
    rm -f "$RELEASE_NOTES_FILE"
    exit 1
fi
popd > /dev/null

echo "✔ Archivio creato: $ZIP_PATH"

# 4. Operazioni Git
echo "--- Operazioni Git ---"

# Controlla se ci sono cambiamenti in build_macos.sh o update.go da committare
if git diff --quiet build_macos.sh update.go; then
    echo "Nessun cambiamento rilevato in build_macos.sh o update.go (la versione era già aggiornata?)"
else
    git add build_macos.sh update.go
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

if [ "$USE_GENERATED_NOTES" = true ]; then
    # Usa --generate-notes se non abbiamo note dal changelog
    gh release create "v$VERSION" "$ZIP_PATH" --title "v$VERSION" --generate-notes
else
    # Usa il file delle note estratto
    gh release create "v$VERSION" "$ZIP_PATH" --title "v$VERSION" --notes-file "$RELEASE_NOTES_FILE"
fi

EXIT_CODE=$?

# Cleanup
rm -f "$RELEASE_NOTES_FILE"

if [ $EXIT_CODE -eq 0 ]; then
    echo "=== Rilascio v$VERSION completato con successo! ==="
else
    echo "Errore: Creazione release GitHub fallita."
    exit 1
fi