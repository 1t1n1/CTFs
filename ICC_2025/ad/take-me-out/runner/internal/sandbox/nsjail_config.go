package sandbox

import (
	_ "embed"
	"os"
	"sync"
)

//go:embed nsjail_config.cfg
var nsjailConfigData string

var (
	configOnce sync.Once
	configPath string
	configErr  error
)

func nsjailConfigPath() (string, error) {
	configOnce.Do(func() {
		f, err := os.CreateTemp("", "nsjail-config-*.cfg")
		if err != nil {
			configErr = err
			return
		}
		if _, err := f.WriteString(nsjailConfigData); err != nil {
			_ = f.Close()
			configErr = err
			return
		}
		if err := f.Close(); err != nil {
			configErr = err
			return
		}
		if err := os.Chmod(f.Name(), 0o644); err != nil {
			configErr = err
			return
		}
		configPath = f.Name()
	})
	return configPath, configErr
}
