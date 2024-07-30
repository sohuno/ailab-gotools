package configutils

import (
	"k8s.io/klog/v2"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

func LoadConfig(v *viper.Viper, cfgFile string, defaultCfgName string) {
	if cfgFile != "" {
		// Use config utils file from the flag.
		v.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			klog.Errorf("get homedir failed: %+v", err)
			os.Exit(1)
		}

		// Search config utils in home directory with name "leading_worker" (without extension).
		v.AddConfigPath(home)
		v.SetConfigName(defaultCfgName)
	}

	// If a config utils file is found, read it in.
	if err := v.ReadInConfig(); err == nil {
		klog.V(1).Infof("Using configutils file: %s; configutils name: %s", cfgFile, defaultCfgName)
	}
}
