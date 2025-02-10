// ui/ui.go
package ui

import (
    "context"
    "fmt"
    "sync"
    "github.com/nsf/termbox-go"
)

type UI interface {
    Run(ctx context.Context) error
    ShowProgress(title string, progress float64)
    CancelOperation()
    ShowMenu()
    HandleSelection()
}

type ConsoleUI struct {
    ns        *NoSpyFon
    selected  int
    menuItems []string
    executing bool
    ctx       context.Context
    cancel    func()
    mu        sync.Mutex
}

func NewConsoleUI(ns *NoSpyFon) UI {
    ctx, cancel := context.WithCancel(context.Background())
    return &ConsoleUI{
        ns:        ns,
        selected:  0,
        menuItems: []string{
            "0 - Сделать резервную копию",
            "1 - Сохранить ядро",
            "2 - Сохранить драйверы",
            "3 - Выгрузить приложения",
            "4 - Просканировать на вирусы",
            "5 - Найти чистую версию Android",
            "6 - Сборка",
            "7 - Сканер APK",
            "q - Выход",
        },
        ctx:     ctx,
        cancel:  cancel,
        executing: false,
    }
}

func (c *ConsoleUI) ShowProgress(title string, progress float64) {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
    fmt.Println(title)
    fmt.Printf("\nПрогресс: [")
    
    width := 40
    filled := int(float64(width) * progress / 100)
    for i := 0; i < width; i++ {
        if i < filled {
            fmt.Print("=")
        } else {
            fmt.Print("-")
        }
    }
    
    fmt.Printf("] %.1f%%\n", progress)
    termbox.Flush()
}

func (c *ConsoleUI) CancelOperation() {
    c.mu.Lock()
    defer c.mu.Unlock()
    if c.executing {
        c.cancel()
        c.executing = false
    }
}

func (c *ConsoleUI) showMenu() {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
    fmt.Println("Меню:")
    fmt.Println("-----")
    
    for i, item := range c.menuItems {
        if i == c.selected {
            fmt.Printf("=> %s\n", item)
        } else {
            fmt.Printf("   %s\n", item)
        }
    }
    termbox.Flush()
}

func (c *ConsoleUI) executeAsync(action func(ctx context.Context, progress chan float64) error) error {
    if c.executing {
        return fmt.Errorf("другая операция уже выполняется")
    }
    
    c.executing = true
    progressChan := make(chan float64)
    
    go func() {
        defer func() {
            c.executing = false
            close(progressChan)
            if err := recover(); err != nil {
                logger.Printf("Паника в горутине: %v", err)
            }
        }()
        
        err := action(c.ctx, progressChan)
        c.mu.Lock()
        defer c.mu.Unlock()
        
        termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
        if err != nil {
            fmt.Printf("Ошибка: %v\n", err)
        } else {
            fmt.Println("Действие выполнено успешно")
        }
        
        fmt.Println("\nНажмите любую клавишу для продолжения...")
        termbox.Flush()
        
        termbox.PollEvent()
        c.showMenu()
    }()
    
    return nil
}

func (c *ConsoleUI) handleSelection() {
    switch c.selected {
    case 0:
        c.executeAsync(func(ctx context.Context, progress chan float64) error {
            return c.ns.backup.Execute(ctx, progress)
        })
    case 1:
        c.executeAsync(func(ctx context.Context, progress chan float64) error {
            return c.ns.kernel.Execute(ctx, progress)
        })
    case 2:
        c.executeAsync(func(ctx context.Context, progress chan float64) error {
            return c.ns.drivers.Execute(ctx, progress)
        })
    case 3:
        c.executeAsync(func(ctx context.Context, progress chan float64) error {
            return c.ns.apps.Execute(ctx, progress)
        })
    case 4:
        c.executeAsync(func(ctx context.Context, progress chan float64) error {
            return c.ns.virusScanner.Execute(ctx, progress)
        })
    case 5:
        c.executeAsync(func(ctx context.Context, progress chan float64) error {
            return c.ns.androidSelector.Execute(ctx, progress)
        })
    case 6:
        c.executeAsync(func(ctx context.Context, progress chan float64) error {
            return c.ns.builder.Execute(ctx, progress)
        })
    case 7:
        c.executeAsync(func(ctx context.Context, progress chan float64) error {
            return c.ns.apkScanner.Execute(ctx, progress)
        })
    }
}

func mainLoop(ui UI) {
    for {
        switch ev := termbox.PollEvent(); ev.Type {
        case termbox.EventKey:
            switch ev.Ch {
            case 0: // Специальные клавиши
                switch ev.Key {
                case termbox.KeyArrowUp:
                    ui.(*ConsoleUI).selected = (ui.(*ConsoleUI).selected - 1 + len(ui.(*ConsoleUI).menuItems)) % len(ui.(*ConsoleUI).menuItems)
                    ui.ShowMenu()
                case termbox.KeyArrowDown:
                    ui.(*ConsoleUI).selected = (ui.(*ConsoleUI).selected + 1) % len(ui.(*ConsoleUI).menuItems)
                    ui.ShowMenu()
                case termbox.KeyEnter:
                    ui.HandleSelection()
                }
            case 'q', 'Q':
                return
            }
        }
    }
}

func (c *ConsoleUI) Run(ctx context.Context) error {
    err := termbox.Init()
    if err != nil {
        return fmt.Errorf("ошибка инициализации терминала: %w", err)
    }
    defer termbox.Close()
    
    c.showMenu()
    mainLoop(c)
    return nil
}
