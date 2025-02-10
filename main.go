// main.go
 package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "sync"
    "time"

    "github.com/nsf/termbox-go"

    "github.com/war41k/nospyfon2/ui"
    "github.com/war41k/nospyfon2/managers"
)
   

var logger *log.Logger
var logFile *os.File

const (
    maxLogSize = 1024 * 1024 // 1MB
    maxBackups = 3
)

func initLogger() error {
    var err error
    logFile, err = os.OpenFile("nospyfon.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        return err
    }
    
    logger = log.New(logFile, "[NoSpyFon] ", log.Ldate|log.Ltime|log.Lshortfile)
    go rotateLogs()
    return nil
}

func rotateLogs() {
    ticker := time.NewTicker(time.Minute)
    defer ticker.Stop()
    
    for range ticker.C {
        if logFile != nil {
            logFile.Close()
            
            fileInfo, err := os.Stat("nospyfon.log")
            if err != nil || fileInfo.Size() <= maxLogSize {
                logFile, _ = os.OpenFile("nospyfon.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
                continue
            }
            
            for i := maxBackups; i > 0; i-- {
                oldName := fmt.Sprintf("nospyfon.log.%d", i-1)
                newName := fmt.Sprintf("nospyfon.log.%d", i)
                os.Rename(oldName, newName)
            }
            
            os.Rename("nospyfon.log", "nospyfon.log.0")
            logFile, _ = os.OpenFile("nospyfon.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        }
    }
}

func cleanup() {
    if logFile != nil {
        logFile.Close()
    }
}

func checkTool(name string) error {
    _, err := exec.LookPath(name)
    if err != nil {
        return fmt.Errorf("необходимый инструмент %q не найден: %w", name, err)
    }
    return nil
}

func initManagers() (*NoSpyFon, error) {
    if err := checkTool("adb"); err != nil {
        return nil, err
    }
    if err := checkTool("apktool"); err != nil {
        return nil, err
    }
    
    ns := &NoSpyFon{
        backup:        NewBackupManager(),
        kernel:        NewKernelManager(),
        drivers:       NewDriversManager(),
        apps:          NewAppsManager(),
        virusScanner:  NewVirusScanner("ваш_ключ_api", "apktool"),
        androidSelector: NewAndroidSelector(),
        builder:       NewBuilder(),
        apkScanner:    NewApkToolScanner(),
    }
    
    return ns, nil

func main() {
    if err := initLogger(); err != nil {
        log.Fatal(err)
    }
    
    ns, err := initManagers()
    if err != nil {
        logger.Printf("Ошибка инициализации менеджеров: %v", err)
        cleanup()
        return
    }
    
    ctx, cancel := context.WithCancel(context.Background())
    ui := NewConsoleUI(ns)
    defer func() {
        cancel()
        cleanup()
    }()
    
    if err := ui.Run(ctx); err != nil {
        logger.Printf("Ошибка выполнения: %v", err)
    }
}
