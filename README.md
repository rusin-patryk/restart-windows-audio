📘 This README is also available in [Polish 🇵🇱](README.pl.md)

---

# RestartAudioGG – Restart Audio Services + SteelSeries GG (Windows 10/11)

On my system, I encountered issues such as no sound after waking from sleep. SteelSeries GG would not detect the headset after manually restarting the Windows Audio service. Running the program below resolves this problem by restarting the necessary services.

1. Restarts **Windows Audio** and optional related services
2. Terminates *all* SteelSeries GG processes
3. Launches `SteelSeriesGGClient.exe` and hides its window (icon appears in the system tray)

> **Note** – Administrator privileges are required to control system services.

---

## Table of Contents

1. [Requirements](#requirements)
2. [Quick Start](#quick-start)
3. [Customization](#customization)
4. [Building](#building)
5. [Running](#running)
6. [Troubleshooting](#troubleshooting)

---

## Requirements

| Tool     | Version                                                                       |
| -------- | ----------------------------------------------------------------------------- |
| **Go**   | ≥ 1.22 (Windows amd64)                                                        |
| **Git**  | Any current version                                                           |
| **rsrc** | `go install github.com/akavel/rsrc@latest` – for generating the icon resource |

---

## Quick Start

```powershell
# 1) Generate resources (icon) – creates ggicon.syso
rsrc -ico icon.ico -o ggicon.syso

# 2) Build the EXE in GUI mode (no console window)
go build -ldflags "-H=windowsgui" -o RestartAudioGG.exe
```

---

## Recreating from main.go

```powershell
# 1) Create a module
go mod init restartwindowsaudio

# 2) Add Windows API dependency
go get golang.org/x/sys/windows@latest
```

Run `RestartAudioGG.exe` as administrator – audio services will be
restarted, GG will be closed and restarted.

---

## Customization

| What to Change      | Where                                   |
| ------------------- | --------------------------------------- |
| GG Executable Path  | Constant `ggClientExe` in `main.go`     |
| GG Processes        | Array `ggProcesses` in `main.go`        |
| Additional Services | Array `dependents` in `main.go`         |
| App Icon            | Replace `icon.ico` **and** rerun `rsrc` |

---

## Building

```powershell
# Generate (after icon change)
rsrc -ico icon.ico -o ggicon.syso

# Final build; -s -w strips symbols, reduces file size
go build -ldflags "-H=windowsgui -s -w" -o RestartAudioGG.exe
```

The `ggicon.syso` file is created locally and not committed –
this keeps the repo clean and allows each developer to generate their own resources.

---

## Running

1. Right-click → Run as administrator, or use elevated PowerShell:

```powershell
Start-Process .\RestartAudioGG.exe -Verb RunAs
```

2. The app runs in the background, without a console window.

---

## Troubleshooting

| Symptom                             | Cause / Solution                                                                      |
| ----------------------------------- | ------------------------------------------------------------------------------------- |
| **No icon** in the EXE file         | Make sure `ggicon.syso` is next to `main.go` during `go build`.                       |
| **Black window flickers** on launch | Ensure all `exec.Command` calls use `SysProcAttr{HideWindow:true}`.                   |
| **Error stopping service**          | Add missing dependent services to `dependents` or run the app as administrator.       |
| **GG not showing in system tray**   | Update window titles in `closeGGWindows()` – GG may have changed captions in updates. |

---
