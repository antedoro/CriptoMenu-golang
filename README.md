# CriptoMenu

CriptoMenu è una semplice applicazione per la menubar di macOS che ti permette di monitorare in tempo reale le quotazioni di criptovalute da Binance.

## Caratteristiche

*   **Quotazioni in tempo reale:** Visualizza il prezzo di una coppia di criptovalute selezionata direttamente nella menubar.
*   **Supporto Binance:** Si connette all'API Spot di Binance per ottenere dati sui prezzi.
*   **Prezzi arrotondati:** I prezzi sono visualizzati arrotondati a due cifre decimali.
*   **Configurazione flessibile:** Definisci le coppie di criptovalute da monitorare tramite un file di configurazione JSON.
*   **Menu interattivo:**
    *   **Monitored Pairs:** Seleziona al volo la coppia da visualizzare tra quelle configurate.
    *   **Edit Config:** Apre il file di configurazione `~/.criptomenu.json` nel tuo editor predefinito per una facile modifica.
    *   **Aggiornamento automatico:** Il menu "Monitored Pairs" si aggiorna automaticamente quando salvi le modifiche al file di configurazione.
    *   **Quit:** Esce dall'applicazione.
*   **Applicazione Standalone:** Distribuita come un'applicazione `.app` nativa per macOS.

## Installazione

1.  **Clona il repository (se disponibile) o scarica i file sorgente.**
2.  **Assicurati di avere Go installato.** Puoi scaricarlo da [go.dev](https://go.dev/dl/).
3.  **Costruisci l'applicazione:**
    Apri il terminale nella directory principale del progetto e usa i seguenti comandi:
    ```bash
    go build -o CriptoMenu.app/Contents/MacOS/CriptoMenu
    ```
4.  **Genera l'icona dell'applicazione (.icns):**
    Assicurati di avere un file `icon.png` (idealmente quadrato, di buona qualità) nella directory del progetto. Questo script creerà l'icona `.icns` e la inserirà nel bundle dell'app:
    ```bash
    mkdir CriptoMenu.iconset
    sips -z 16 16     icon.png --out CriptoMenu.iconset/icon_16x16.png
    sips -z 32 32     icon.png --out CriptoMenu.iconset/icon_16x16@2x.png
    sips -z 32 32     icon.png --out CriptoMenu.iconset/icon_32x32.png
    sips -z 64 64     icon.png --out CriptoMenu.iconset/icon_32x32@2x.png
    sips -z 128 128   icon.png --out CriptoMenu.iconset/icon_128x128.png
    sips -z 256 256   icon.png --out CriptoMenu.iconset/icon_128x128@2x.png
    sips -z 256 256   icon.png --out CriptoMenu.iconset/icon_256x256.png
    sips -z 512 512   icon.png --out CriptoMenu.iconset/icon_256x256@2x.png
    sips -z 512 512   icon.png --out CriptoMenu.iconset/icon_512x512.png
    sips -z 1024 1024 icon.png --out CriptoMenu.iconset/icon_512x512@2x.png
    iconutil -c icns CriptoMenu.iconset
    mv CriptoMenu.icns CriptoMenu.app/Contents/Resources/AppIcon.icns
    rm -rf CriptoMenu.iconset
    ```
5.  **Crea il file `Info.plist`:**
    Questo file definisce i metadati dell'applicazione. Crealo come `CriptoMenu.app/Contents/Info.plist` con il seguente contenuto:
    ```xml
    <?xml version="1.0" encoding="UTF-8"?>
    <!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
    <plist version="1.0">
    <dict>
        <key>CFBundleExecutable</key>
        <string>CriptoMenu</string>
        <key>CFBundleIconFile</key>
        <string>AppIcon</string>
        <key>CFBundleIdentifier</key>
        <string>com.antedoro.criptomenu</string>
        <key>CFBundleName</key>
        <string>CriptoMenu</string>
        <key>CFBundlePackageType</key>
        <string>APPL</string>
        <key>CFBundleShortVersionString</key>
        <string>1.0.0</string>
        <key>LSUIElement</key>
        <true/>
        <key>NSHighResolutionCapable</key>
        <true/>
    </dict>
    </plist>
    ```
6.  **Crea il file `PkgInfo`:**
    Questo è un piccolo file per macOS. Crealo come `CriptoMenu.app/Contents/PkgInfo` con il seguente contenuto:
    ```
    APPL????
    ```
7.  **Sposta l'applicazione:**
    Sposta la cartella `CriptoMenu.app` nella cartella `/Applications` o in qualsiasi altra posizione desiderata.

## Utilizzo

1.  **Avvia l'applicazione:** Fai doppio click su `CriptoMenu.app`.
2.  **Configura le coppie da monitorare:**
    *   Clicca sull'icona dell'applicazione nella menubar.
    *   Seleziona "Edit Config". Si aprirà il file `~/.criptomenu.json` nel tuo editor predefinito.
    *   Modifica l'array `"Pairs"` con le coppie di criptovalute che desideri monitorare (es. `["BTCUSDC", "ETHUSDC", "BNBUSDT"]`).
    *   Salva il file. Il menu "Monitored Pairs" si aggiornerà automaticamente.
3.  **Seleziona la coppia da visualizzare:**
    *   Clicca sull'icona dell'applicazione nella menubar.
    *   Passa il mouse su "Monitored Pairs".
    *   Clicca sulla coppia che vuoi visualizzare nella menubar.

## Risoluzione dei Problemi

*   **Icona non visualizzata correttamente:** Se l'icona dell'app non appare o mostra un'icona generica, il sistema potrebbe averla messa in cache. Prova a spostare `CriptoMenu.app` in un'altra cartella e poi di nuovo nella sua posizione originale, oppure esegui il seguente comando nel terminale:
    ```bash
    touch CriptoMenu.app; killall Dock; killall Finder
    ```
*   **Prezzi non aggiornati / Errori:** Assicurati di avere una connessione internet attiva. Verifica che i simboli delle coppie in `~/.criptomenu.json` siano validi su Binance (es. `BTCUSDC`, non `BTC-USDC`).

## Tecnologie Utilizzate

*   **Go Lang:** Il linguaggio di programmazione principale.
*   **systray:** Libreria per la gestione dell'icona e del menu nella system tray.
*   **Binance Connector Go:** Libreria per interagire con l'API di Binance.
*   **AppleScript (osascript):** Utilizzato per aprire il file di configurazione con l'editor predefinito.
