// +build windows

// Restart Audio + SteelSeries GG
//  1. Restarts Windows Audio services (Audiosrv + RtkAudioUniversalService)
//  2. Kills all SteelSeries GG-related processes
//  3. Launches SteelSeriesGGClient.exe
//  4. Closes the GG window after a few seconds, leaving the tray icon running

package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "os/exec"
    "strings"
    "syscall"
    "time"
    "unsafe"

    "golang.org/x/sys/windows"
    "golang.org/x/sys/windows/svc"
    "golang.org/x/sys/windows/svc/mgr"
)

/* ---------- Configuration (Edit for your system) ---------- */

const (
    baseService = "Audiosrv" // Main audio service name
    ggClientExe = `C:\Program Files\SteelSeries\GG\SteelSeriesGGClient.exe` // Full path to SteelSeriesGGClient.exe
    WM_CLOSE    = 0x0010 // Message used to close windows
)

// Additional audio-related services to restart (modify if needed)
var dependents = []string{
    "RtkAudioUniversalService",
}

// SteelSeries GG processes to terminate
var ggProcesses = []string{
    "SteelSeriesGGClient.exe",
    "SteelSeriesGG.exe",
    "SteelSeriesSonar.exe",
    "SteelSeriesEngine.exe",
    "SteelSeriesEngine3.exe",
}

/* ---------- Main ---------- */

func main() {
    // 1) Restart audio services
    if err := restartAudio(); err != nil {
        log.Fatalf("Audio restart failed: %v", err)
    }

    // 2) Kill all GG-related processes
    fmt.Println("Terminating SteelSeries GG processes...")
    killAllGG()

    // 3) Wait until all processes are gone
    waitUntilGone(ggProcesses, 10*time.Second)

    // 4) Start GG launcher hidden
    fmt.Println("Starting SteelSeries GG...")
    if err := startHidden(ggClientExe); err != nil {
        log.Fatalf("Failed to start GG: %v", err)
    }

    // 5) Wait, then close the GG window (remains in tray)
    time.Sleep(5 * time.Second)
    closeGGWindows()

    fmt.Println("✓ SteelSeries GG is running in the system tray.")
    fmt.Print("Press Enter to exit...")
    _, _ = bufio.NewReader(os.Stdin).ReadBytes('\n')
}

/* ---------- Audio service control ---------- */

func restartAudio() error {
    m, err := mgr.Connect()
    if err != nil {
        return err
    }
    defer m.Disconnect()

    for _, name := range dependents {
        if err := stopSvc(m, name); err != nil {
            return err
        }
    }

    if err := stopSvc(m, baseService); err != nil {
        return err
    }

    if err := startSvc(m, baseService); err != nil {
        return err
    }

    for i := len(dependents) - 1; i >= 0; i-- {
        if err := startSvc(m, dependents[i]); err != nil {
            return err
        }
    }

    return nil
}

func stopSvc(m *mgr.Mgr, name string) error {
    s, err := m.OpenService(name)
    if err != nil {
        return err
    }
    defer s.Close()

    status, err := s.Query()
    if err != nil {
        return err
    }
    if status.State == svc.Stopped {
        return nil
    }

    fmt.Printf("• Stopping %s...\n", name)
    if _, err := s.Control(svc.Stop); err != nil {
        return err
    }
    return waitState(s, svc.Stopped, 30*time.Second)
}

func startSvc(m *mgr.Mgr, name string) error {
    s, err := m.OpenService(name)
    if err != nil {
        return err
    }
    defer s.Close()

    status, err := s.Query()
    if err != nil {
        return err
    }
    if status.State == svc.Running {
        return nil
    }

    fmt.Printf("• Starting %s...\n", name)
    if err := s.Start(); err != nil {
        return err
    }
    return waitState(s, svc.Running, 30*time.Second)
}

func waitState(s *mgr.Service, desired svc.State, timeout time.Duration) error {
    deadline := time.Now().Add(timeout)
    for {
        status, err := s.Query()
        if err != nil {
            return err
        }
        if status.State == desired {
            return nil
        }
        if time.Now().After(deadline) {
            return fmt.Errorf("timeout waiting for state %v", desired)
        }
        time.Sleep(300 * time.Millisecond)
    }
}

/* ---------- GG process management ---------- */

func killAllGG() {
    // taskkill /F /T /IM proc1 /IM proc2 …
    args := []string{"/F", "/T"}
    for _, img := range ggProcesses {
        args = append(args, "/IM", img)
    }
    cmd := exec.Command("taskkill", args...)
    cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true} // ← tu magia
    _ = cmd.Run()
}

func waitUntilGone(processList []string, timeout time.Duration) {
    deadline := time.Now().Add(timeout)
    for time.Now().Before(deadline) {
        allGone := true
        for _, img := range processList {
            if isRunning(img) {
                allGone = false
                break
            }
        }
        if allGone {
            return
        }
        time.Sleep(400 * time.Millisecond)
    }
}

func isRunning(img string) bool {
    cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq "+img)
    cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
    out, _ := cmd.CombinedOutput()
    return strings.Contains(strings.ToLower(string(out)), strings.ToLower(img))
}

/* ---------- Launch & window management ---------- */

// startHidden launches a program with no visible window
func startHidden(path string) error {
    cmd := exec.Command(path)
    cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
    return cmd.Start()
}

// closeGGWindows scans for GG-related windows and sends WM_CLOSE to hide them
func closeGGWindows() {
    user32 := windows.NewLazySystemDLL("user32.dll")
    enumWindows := user32.NewProc("EnumWindows")
    getWindowTextW := user32.NewProc("GetWindowTextW")
    isWindowVisible := user32.NewProc("IsWindowVisible")
    sendMessageW := user32.NewProc("SendMessageW")

    targets := []string{
        "steelseries gg",
        "steelseries sonar",
        "steelseries engine",
    }

    cb := syscall.NewCallback(func(hwnd uintptr, _ uintptr) uintptr {
        vis, _, _ := isWindowVisible.Call(hwnd)
        if vis == 0 {
            return 1 // skip invisible windows
        }

        buf := make([]uint16, 260)
        getWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
        title := strings.ToLower(windows.UTF16ToString(buf))
        if title == "" {
            return 1
        }

        for _, t := range targets {
            if strings.Contains(title, t) {
                sendMessageW.Call(hwnd, WM_CLOSE, 0, 0)
                break
            }
        }
        return 1
    })

    enumWindows.Call(cb, 0)
}
