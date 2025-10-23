package configparser

import (
	// "log"

	"log"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func InitConfigParser() (*viper.Viper, error) {
	var err error
	v := viper.New()
	v.AddConfigPath("./configs")
	v.SetConfigType("yaml")
	v.SetConfigName("configuration")
	v.OnConfigChange(func(e fsnotify.Event) {
		log.Println("Config file has changed:", e.Name)
	})
	v.WatchConfig()
	// If a config file is found, read it in.
	if err = v.ReadInConfig(); err == nil {
		log.Println("Using config file:", v.ConfigFileUsed())
		// fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	}
	return v, err
}
