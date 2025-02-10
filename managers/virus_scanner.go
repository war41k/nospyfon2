// managers/virus_scanner.go
package managers

import (
    "context"
    "fmt"
    "io"
    "net/http"
    "os"
    "os/exec"
    "path/filepath"
    "strings"
    "sync"
    "time"
    "log"
)

var logger *log.Logger

func init() {
    file, err := os.OpenFile("nospyfon.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
    if err != nil {
        log.Fatal(err)
    }
    logger = log.New(file, "[NoSpyFon] ", log.Ldate|log.Ltime|log.Lshortfile)
}

type VirusScanner struct {
    apiKey string
    mu     sync.Mutex
    client *http.Client
    apktoolPath string
}

func (s *VirusScanner) getFilesToScan() ([]string, error) {
    var files []string
    
    err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if !info.IsDir() {
            files = append(files, path)
        }
        
        return nil
    })
    
    return files, err
}

func checkTool(name string) error {
    _, err := exec.LookPath(name)
    if err != nil {
        return fmt.Errorf("необходимый инструмент %q не найден: %w", name, err)
    }
    return nil
}

func NewVirusScanner(apiKey, apktoolPath string) *VirusScanner {
    return &VirusScanner{
        apiKey: apiKey,
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
        apktoolPath: apktoolPath,
    }
}

func (s *VirusScanner) CheckTools() error {
    if err := checkTool("curl"); err != nil {
        return err
    }
    if err := checkTool(s.apktoolPath); err != nil {
        return fmt.Errorf("apktool не найден: %w", err)
    }
    return nil
}

func (s *VirusScanner) Execute(ctx context.Context, progress chan float64) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    files, err := s.getFilesToScan()
    if err != nil {
        return fmt.Errorf("ошибка получения списка файлов: %w", err)
    }
    
    totalFiles := len(files)
    processedFiles := 0
    
    for _, file := range files {
        select {
        case <-ctx.Done():
            return ctx.Err()
        default:
            err = s.scanFile(file)
            if err != nil {
                logger.Printf("Ошибка сканирования файла %s: %v", file, err)
            }
            
            processedFiles++
            progressPercentage := (float64(processedFiles) / float64(totalFiles)) * 100
            progress <- progressPercentage
        }
    }
    
    return nil
}

func (s *VirusScanner) scanFile(filePath string) error {
    // Проверяем, что это APK файл
    if !strings.HasSuffix(filePath, ".apk") {
        return nil
    }
    
    // Создаем временную директорию для декомпиляции
    tempDir, err := os.MkdirTemp("", "apk_decompilation")
    if err != nil {
        return fmt.Errorf("ошибка создания временной директории: %w", err)
    }
    defer os.RemoveAll(tempDir)
    
    // Декомпилируем APK
    cmd := exec.Command(
        s.apktoolPath,
        "d",
        "-s", // Без декомпиляции в smali
        "-f", // Принудительное перезаписывание
        filePath,
        "-o",
        tempDir,
    )
    
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("ошибка декомпиляции APK: %w\n%s", err, string(output))
    }
    
    // Сканируем декомпилированные файлы
    return s.scanDecompiledFiles(tempDir)
}

func (s *VirusScanner) scanDecompiledFiles(dir string) error {
    files, err := filepath.Glob(filepath.Join(dir, "**/*"))
    if err != nil {
        return fmt.Errorf("ошибка получения списка файлов: %w", err)
    }
    
    for _, file := range files {
        if err := s.scanFileWithVirusTotal(file); err != nil {
            return fmt.Errorf("ошибка сканирования файла %s: %w", file, err)
        }
    }
    
    return nil
}

func (s *VirusScanner) scanFileWithVirusTotal(filePath string) error {
    file, err := os.Open(filePath)
    if err != nil {
        return fmt.Errorf("ошибка открытия файла: %w", err)
    }
    defer file.Close()
    
    url := fmt.Sprintf("https://www.virustotal.com/api/v3/files/scan")
    
    req, err := http.NewRequestWithContext(context.Background(), "POST", url, file)
    if err != nil {
        return fmt.Errorf("ошибка создания запроса: %w", err)
    }
    
    req.Header.Set("x-apikey", s.apiKey)
    req.Header.Set("Content-Type", "application/octet-stream")
    
    resp, err := s.client.Do(req)
    if err != nil {
        return fmt.Errorf("ошибка отправки файла: %w", err)
    }
    defer resp.Body.Close()
    
    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("ошибка чтения ответа: %w", err)
    }
    
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("ошибка API: %s", string(body))
    }
    
    return nil
}
