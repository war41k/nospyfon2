package managers

import (
    "bytes"
    "fmt"
    "io"
    "os"
    "path/filepath"
    "time"
    "compress/gzip"
    "archive/tar"
)

type BackupManager struct {
    backupDir string
}

func NewBackupManager() *BackupManager {
    return &BackupManager{
        backupDir: "./backups",
    }
}

func (b *BackupManager) MakeBackup() error {
    timestamp := time.Now().Format("20060102150405")
    backupPath := filepath.Join(b.backupDir, fmt.Sprintf("backup_%s.tar.gz", timestamp))
    
    f, err := os.Create(backupPath)
    if err != nil {
        return fmt.Errorf("ошибка создания файла резервной копии: %w", err)
    }
    defer f.Close()
    
    gw := gzip.NewWriter(f)
    defer gw.Close()
    
    tw := tar.NewWriter(gw)
    defer tw.Close()
    
    err = filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if !info.IsDir() {
            header, err := tar.FileInfoHeader(info, "")
            if err != nil {
                return err
            }
            
            header.Name = path
            err = tw.WriteHeader(header)
            if err != nil {
                return err
            }
            
            f, err := os.Open(path)
            if err != nil {
                return err
            }
            defer f.Close()
            
            _, err = io.Copy(tw, f)
            return err
        }
        
        return nil
    })
    
    return err
}
