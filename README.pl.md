📘 This README is also available in [English EN](README.md)

---

# RestartAudioGG – restart usług audio + SteelSeries GG (Windows 10/11)

Na moim systemie spotkałem się z takimi problemami jak: po wybudzeniu systemu, brak dźwięku. SteelSeries GG nie widzi słuchawek po ręcznym restarcie usługi Windows Audio. Poniższy program po uruchomieniu rozwiązuje ten problem restartując usługi.

1. Restartuje **Windows Audio** i opcjonalne usługi pokrewne  
2. Zamyka *wszystkie* procesy SteelSeries GG  
3. Uruchamia `SteelSeriesGGClient.exe` i chowa jego okno (ikona ląduje w trayu)

> **Uwaga** – do sterowania usługami potrzebne są uprawnienia administratora.

---

## Spis treści
1. [Wymagania](#wymagania)  
2. [Szybki start](#szybki-start)  
3. [Dostosowanie](#dostosowanie)  
4. [Budowanie](#budowanie)  
5. [Uruchamianie](#uruchamianie)  
6. [Rozwiązywanie problemów](#rozwiązywanie-problemów)  

---

## Wymagania
| Narzędzie | Wersja                     |
|-----------|---------------------------|
| **Go**    | ≥ 1.22 (Windows amd64)    |
| **Git**   | dowolna bieżąca           |
| **rsrc**  | `go install github.com/akavel/rsrc@latest` – do generowania zasobu ikony |

---

## Szybki start
```powershell
# 1) generujemy zasoby (ikonę) – powstaje ggicon.syso
rsrc -ico icon.ico -o ggicon.syso

# 2) budujemy EXE w trybie GUI (bez konsoli)
go build -ldflags "-H=windowsgui" -o RestartAudioGG.exe
```

---

## Odtworzenie z pliku main.go
```powershell
# 1) tworzenie modułu
go mod init restartwindowsaudio

# 2) dodanie zależności do Windows API
go get golang.org/x/sys/windows@latest
```

Uruchom RestartAudioGG.exe jako administrator – usługi audio zostaną
zrestartowane, GG zamknięte i włączone ponownie.

---

## Dostosowanie
| Co zmienić       | Gdzie                                            |
| ---------------- | ------------------------------------------------ |
| Ścieżka do GG    | stała `ggClientExe` w `main.go`                  |
| Inne procesy GG  | tablica `ggProcesses` w `main.go`                |
| Dodatkowe usługi | tablica `dependents` w `main.go`                 |
| Ikona programu   | podmień `icon.ico` **i** ponownie uruchom `rsrc` |

---

## Budowanie
```powershell
# generowanie (po zmianie ikony)
rsrc -ico icon.ico -o ggicon.syso

# finalny build; -s -w usuwa symbole, zmniejsza plik
go build -ldflags "-H=windowsgui -s -w" -o RestartAudioGG.exe
```
Plik ggicon.syso powstaje lokalnie i nie jest commitowany – dzięki temu
repo pozostaje czyste, a każdy deweloper może wygenerować własne zasoby.

---

## Uruchamianie
1. Klik-prawy → Uruchom jako administrator lub z podwyższonego PowerShella:
```powershell
Start-Process .\RestartAudioGG.exe -Verb RunAs
```
2. Program działa w tle, nie wyświetla konsoli.

---

## Rozwiązywanie problemów
| Objaw                                  | Przyczyna / rozwiązanie                                                                    |
| -------------------------------------- | ------------------------------------------------------------------------------------------ |
| **Brak ikony** w pliku EXE             | Upewnij się, że `ggicon.syso` leży obok `main.go` w momencie `go build`.                   |
| **Czarne okno miga** podczas działania | Sprawdź, czy wszystkie wywołania `exec.Command` mają `SysProcAttr{HideWindow:true}`.       |
| **Błąd przy stopie usługi**            | Dodaj brakujące usługi zależne do `dependents` albo uruchom jako administrator.            |
| **Nie widzę GG w trayu**               | Zmień tytuły okien w funkcji `closeGGWindows()` – GG mogło zmienić caption w nowej wersji. |
