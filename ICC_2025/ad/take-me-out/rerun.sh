mkdir -p /tmp/work && cat <<'EOF' > /tmp/work/patch
*** Begin Patch
*** Update File: runner/internal/sandbox/sandbox.go
@@
-import (
-    "context"
-    "errors"
-    "fmt"
-    "io"
-    "io/fs"
-    "log"
-    "os"
-    "os/exec"
-    "path/filepath"
-    "sort"
-    "strings"
-    "sync"
-)
+import (
+    "context"
+    "errors"
+    "fmt"
+    "io"
+    "io/fs"
+    "log"
+    "os"
+    "os/exec"
+    "path/filepath"
+    "sort"
+    "strings"
+    "sync"
+)
@@
-    mounts := []bindMount{{host: envRoot, target: envTarget, readOnly: true}, {host: envRoot, target: rootEnvTarget, readOnly: true}}
+    mounts := []bindMount{
+        {host: envRoot, target: envTarget, readOnly: true},
+        {host: envRoot, target: rootEnvTarget, readOnly: true},
+    }
 
-    bindTargets := map[string]string{
-        filepath.Join(envRoot, "bin"):   filepath.Join(rr.WorkHost, "bin"),
-        filepath.Join(envRoot, "usr"):   filepath.Join(rr.WorkHost, "usr"),
-        filepath.Join(envRoot, "lib"):   filepath.Join(rr.WorkHost, "lib"),
-        filepath.Join(envRoot, "lib64"): filepath.Join(rr.WorkHost, "lib64"),
-        filepath.Join(envRoot, "sbin"):  filepath.Join(rr.WorkHost, "sbin"),
-        filepath.Join(envRoot, "etc"):   filepath.Join(rr.WorkHost, "etc"),
-        filepath.Join(envRoot, "opt"):   filepath.Join(rr.WorkHost, "opt"),
-    }
+    bindTargets := map[string]string{
+        filepath.Join(envRoot, "bin"):   filepath.Join(rr.WorkHost, "bin"),
+        filepath.Join(envRoot, "usr"):   filepath.Join(rr.WorkHost, "usr"),
+        filepath.Join(envRoot, "lib"):   filepath.Join(rr.WorkHost, "lib"),
+        filepath.Join(envRoot, "lib64"): filepath.Join(rr.WorkHost, "lib64"),
+        filepath.Join(envRoot, "sbin"):  filepath.Join(rr.WorkHost, "sbin"),
+        filepath.Join(envRoot, "etc"):   filepath.Join(rr.WorkHost, "etc"),
+        filepath.Join(envRoot, "opt"):   filepath.Join(rr.WorkHost, "opt"),
+    }
 
+    fmt.Fprintf(os.Stderr, "[debug] bindTargets to work: %v\n", bindTargets)
     for host, target := range bindTargets {
         if err := ensureDirWithPerm(target, 0o755); err != nil {
             cleanup()
             return nil, err
         }
         mounts = append(mounts, bindMount{host: host, target: target, readOnly: true})
     }
+    fmt.Fprintf(os.Stderr, "[debug] append workspace bind %s -> %s\n", rr.WorkspaceHost, envWorkspaceTarget)
     mounts = append(mounts, bindMount{host: rr.WorkspaceHost, target: envWorkspaceTarget, readOnly: false})
 
-    rootBindTargets := map[string]string{
-        filepath.Join(envRoot, "bin"):   filepath.Join(rr.Root, "bin"),
-        filepath.Join(envRoot, "usr"):   filepath.Join(rr.Root, "usr"),
-        filepath.Join(envRoot, "lib"):   filepath.Join(rr.Root, "lib"),
-        filepath.Join(envRoot, "lib64"): filepath.Join(rr.Root, "lib64"),
-        filepath.Join(envRoot, "sbin"):  filepath.Join(rr.Root, "sbin"),
-        filepath.Join(envRoot, "etc"):   filepath.Join(rr.Root, "etc"),
-        filepath.Join(envRoot, "opt"):   filepath.Join(rr.Root, "opt"),
-    }
+    rootBindTargets := map[string]string{
+        filepath.Join(envRoot, "bin"):   filepath.Join(rr.Root, "bin"),
+        filepath.Join(envRoot, "usr"):   filepath.Join(rr.Root, "usr"),
+        filepath.Join(envRoot, "lib"):   filepath.Join(rr.Root, "lib"),
+        filepath.Join(envRoot, "lib64"): filepath.Join(rr.Root, "lib64"),
+        filepath.Join(envRoot, "sbin"):  filepath.Join(rr.Root, "sbin"),
+        filepath.Join(envRoot, "etc"):   filepath.Join(rr.Root, "etc"),
+        filepath.Join(envRoot, "opt"):   filepath.Join(rr.Root, "opt"),
+    }
+    fmt.Fprintf(os.Stderr, "[debug] rootBindTargets: %v\n", rootBindTargets)
 
     for host, target := range rootBindTargets {
         if err := ensureDirWithPerm(target, 0o755); err != nil {
             cleanup()
             return nil, err
         }
         mounts = append(mounts, bindMount{host: host, target: target, readOnly: true})
     }
*** End Patch
