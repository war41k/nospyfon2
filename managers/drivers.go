package managers

import (
    "context"
    "fmt"
    "os/exec"
    "sync"
    "time"
    "ioutil"
)
type DriversManager struct{}

func NewDriversManager() *DriversManager {
    return &DriversManager{}
}

func (d *DriversManager) SaveDrivers() error {
    cmd := exec.Command("adb", "shell", "ls", "/sys/class/")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("ошибка получения списка драйверов: %w", err)
    }
    
    err = ioutil.WriteFile("drivers.txt", output, 0644)
    if err != nil {
        return fmt.Errorf("ошибка сохранения списка драйверов: %w", err)
    }
    
    return nil
}
