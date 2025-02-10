package managers

import (
    "context"
    "fmt"
    "os/exec"
    "sync"
    "time"
)

type Builder struct{}

func NewBuilder() *Builder {
    return &Builder{}
}

func (b *Builder) Build() error {
    cmd := exec.Command("make", "build")
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("ошибка сборки: %w\n%s", err, string(output))
    }
    return nil
}
