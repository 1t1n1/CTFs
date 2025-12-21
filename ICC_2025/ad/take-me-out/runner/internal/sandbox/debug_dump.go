package sandbox

import (
    "fmt"
    "io/fs"
    "os"
    "path/filepath"
    "strings"
)

func dumpDir(prefix, root string, maxDepth int) {
    if maxDepth < 0 {
        return
    }
    filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
        if err != nil {
            fmt.Fprintf(os.Stderr, "%s ERR %s: %v\n", prefix, path, err)
            return nil
        }
        rel := strings.TrimPrefix(path, root)
        if rel == "" {
            rel = "/"
        }
        info, _ := d.Info()
        mode := ""
        if info != nil {
            mode = info.Mode().String()
        }
        fmt.Fprintf(os.Stderr, "%s %s %s\n", prefix, rel, mode)
        if d.IsDir() && strings.Count(rel, string(os.PathSeparator)) >= maxDepth {
            return filepath.SkipDir
        }
        return nil
    })
}
