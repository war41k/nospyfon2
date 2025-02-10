// managers/nospyfon.go
package managers

import (
    "context"
    "fmt"
    "sync"
)

type NoSpyFon struct {
    backup        *BackupManager
    kernel        *KernelManager
    drivers       *DriversManager
    apps          *AppsManager
    virusScanner  *VirusScanner
    androidSelector *AndroidSelector
    builder       *Builder
    apkScanner    *ApkToolScanner
}
