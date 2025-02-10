package managers

import (
    "fmt"
    "os/exec"
    "time"
    "strings"
)

type ApkToolScanner struct {
    apktoolPath string
}

func NewApkToolScanner() (*ApkToolScanner, error) {
    _, err := exec.LookPath("apktool")
    if err != nil {
        return nil, fmt.Errorf("apktool не найден: %w", err)
    }
    
    return &ApkToolScanner{
        apktoolPath: "apktool",
    }, nil
}

func (s *ApkToolScanner) ScanApk(apkPath string) (*ApkInfo, error) {
    infoCmd := exec.Command(s.apktoolPath, "d", "--no-src", "-f", apkPath)
    output, err := infoCmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("ошибка декомпиляции APK: %w", err)
    }

    info := &ApkInfo{}
    for _, line := range strings.Split(string(output), "\n") {
        line = strings.TrimSpace(line)
        if strings.HasPrefix(line, "apktool:") {
            continue
        }
        
        if strings.HasPrefix(line, "package:") {
            info.PackageName = strings.TrimPrefix(line, "package:")
        }
        
        if strings.HasPrefix(line, "versionCode:") {
            info.VersionCode = strings.TrimPrefix(line, "versionCode:")
        }
        
        if strings.HasPrefix(line, "versionName:") {
            info.VersionName = strings.TrimPrefix(line, "versionName:")
        }
        
        if strings.HasPrefix(line, "sdkVersion:") {
            info.SdkVersion = strings.TrimPrefix(line, "sdkVersion:")
        }
        
        if strings.HasPrefix(line, "targetSdkVersion:") {
            info.TargetSdkVersion = strings.TrimPrefix(line, "targetSdkVersion:")
        }
    }

    return info, nil
}

type ApkInfo struct {
    PackageName      string
    VersionCode      string
    VersionName      string
    SdkVersion       string
    TargetSdkVersion string
}
