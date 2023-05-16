package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

const (
	DEFAULT_PORT    = 7837
	DEFAULT_API_URL = "https://srink.co"
)

type config struct {
	confPath string
	mu       sync.RWMutex
	data     map[string]any
	log      *log.Logger
}

func newConfig(confPath string, l *log.Logger) *config {
	return &config{
		confPath: confPath,
		data:     make(map[string]any),
		log:      l,
	}
}

// tryAdd adds the key if it doesn't exist already
func (c *config) tryAdd(key string, value any) {
	if _, ok := c.get(key); ok {
		return
	}
	c.add(key, value)
}

func (c *config) add(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = value
}

func (c *config) get(key string) (value any, ok bool) {
	c.mu.RLock()
	value, ok = c.data[key]
	c.mu.RUnlock()
	return
}

func (c *config) getString(key string) string {
	v, ok := c.get(key)
	if !ok {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case int64:
		return strconv.FormatInt(val, 10)
	case int:
		return strconv.FormatInt(int64(val), 10)
	default:
		return fmt.Sprint(val)
	}
}

func (c *config) getInt(key string) int {
	v, ok := c.get(key)
	if !ok {
		return 0
	}
	switch val := v.(type) {
	case int:
		return val
	case string:
		value, _ := strconv.Atoi(val)
		return value
	case int64:
		// might be corrupted
		return int(val)
	default:
		return 0
	}
}

func (c *config) getInt64(key string) int64 {
	v, ok := c.get(key)
	if !ok {
		return 0
	}
	switch val := v.(type) {
	case int64:
		return val
	case int:
		return int64(val)
	case string:
		value, _ := strconv.ParseInt(val, 10, 64)
		return value
	default:
		return 0
	}
}

func (c *config) getBool(key string) bool {
	v, ok := c.get(key)
	if !ok {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case string:
		value, _ := strconv.ParseBool(val)
		return value
	default:
		return false
	}
}

func (c *config) init(in []byte) error {
	return yaml.Unmarshal(in, c.data)
}

func (c *config) build() []byte {
	buf, err := yaml.Marshal(c.data)
	if err != nil {
		c.log.Fatalln("Failed to build config:", err)
	}
	return buf
}

func (c *config) write() {
	err := os.WriteFile(
		c.confPath, c.build(), os.ModePerm,
	)
	if err != nil {
		c.log.Fatalln("Failed to write config:", err)
	}
}

func getUserConfigDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		log.Fatalln("Failed to read user_config_dir:", err)
	}
	return dir
}

func getPath(fs ...string) string {
	return strings.Join(fs, "/")
}

func isDirExist(name string) bool {
	_, err := os.ReadDir(name)
	return !os.IsNotExist(err)
}

func checkConfigDir(userConfDir string) {
	confDir := getPath(userConfDir, "srink")
	if !isDirExist(confDir) {
		err := os.Mkdir(confDir, os.ModePerm)
		if err != nil {
			log.Fatalln("Failed to create config dir:", err)
		}
		log.Println("Created new config dir at", confDir)
	}
}

func readUserConfig(name string, l *log.Logger) *config {
	userConfDir := getUserConfigDir()
	checkConfigDir(userConfDir)
	fPath := getPath(userConfDir, "srink", name)
	conf := newConfig(fPath, l)
	buf, err := os.ReadFile(fPath)
	if err != nil {
		log.Println("Failed to read conf file:", err)
		if os.IsNotExist(err) {
			log.Println("Creating a new", name, "in config dir")
			conf.write()
		} else {
			os.Exit(1)
		}
	}
	err = conf.init(buf)
	if err != nil {
		log.Fatalln("Failed to initialise config:", err)
	}
	return conf
}
