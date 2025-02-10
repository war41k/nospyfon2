package managers

import (
    "fmt"
    "os/exec"
)

type AndroidSelector struct{}

func NewAndroidSelector() *AndroidSelector {
    return &AndroidSelector{}
}

func (a *AndroidSelector) FindCleanVersion() error {
    cmd := exec.Command("adb", "shell", "getprop", "ro.build.version.release")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("ошибка получения версии Android: %w", err)
    }
    
    fmt.Printf("Текущая версия Android: %s\n", string(output))
    return nil
}
