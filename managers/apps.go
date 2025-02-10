package managers

import (
    "context"
    "fmt"
    "os/exec"
    "sync"
    "time"
    "strings"
)
type AppsManager struct{}

func NewAppsManager() *AppsManager {
    return &AppsManager{}
}

func (a *AppsManager) UnloadApps() error {
    cmd := exec.Command("adb", "shell", "pm", "list", "packages")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("ошибка получения списка приложений: %w", err)
    }
    
    for _, pkg := range strings.Split(string(output), "\n") {
        if pkg == "" {
            continue
        }
        
        cmd := exec.Command("adb", "shell", "pm", "clear", pkg)
        if err := cmd.Run(); err != nil {
            return fmt.Errorf("ошибка очистки данных приложения %s: %w", pkg, err)
        }
    }
    
    return nil
}
