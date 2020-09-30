package conf

import (
	"errors"
	"flag"
	"io/ioutil"
	"strings"

	"github.com/robert-pkg/micro-go/log"
	jaeger_trace "github.com/robert-pkg/micro-go/trace/jaeger-trace"
	"gopkg.in/yaml.v2"
)

// Conf global variable.
var (
	confPath string
	Conf     = &Config{}
)

// Config struct of conf.
type Config struct {
	Log log.LogConfig `yaml:"log"`

	ServerConfig ServerConfig        `yaml:"server"`
	AuthConfig   AuthConfig          `yaml:"auth"`
	TraceConfig  jaeger_trace.Config `yaml:"jaeger"`
}

// AuthConfig .
type AuthConfig struct {
	VerifyToken      string `yaml:"verify_token"`
	veriryServerName string
	verifyMethod     string

	SkipToken    []string `yaml:"skip_token"`
	skipTokenMap map[string]bool
}

// ServerConfig .
type ServerConfig struct {
	Port int `yaml:"port"`

	SignKey string `yaml:"sign_key"`
}

// IsSkipToken .
func (c *Config) IsSkipToken(name string) bool {
	if _, ok := c.AuthConfig.skipTokenMap[name]; ok {
		return true
	}

	return false
}

// GetVerifyTokenInfo .
func (c *Config) GetVerifyTokenInfo() (string, string) {
	return c.AuthConfig.veriryServerName, c.AuthConfig.verifyMethod
}

func (c *Config) check() error {

	if len(c.ServerConfig.SignKey) <= 0 {
		return errors.New("签名Key未配置")
	}

	if c.ServerConfig.Port <= 0 {
		return errors.New("端口未配置")
	}

	return nil
}

func (c *Config) loadFromFile(path string) error {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal([]byte(data), &c)
	if err != nil {
		return err
	}

	c.AuthConfig.skipTokenMap = make(map[string]bool)
	for _, r := range c.AuthConfig.SkipToken {
		c.AuthConfig.skipTokenMap[r] = true
	}

	if true {
		ss := strings.Split(c.AuthConfig.VerifyToken, "/")
		if len(ss) != 2 {
			return errors.New("verify token config error")
		}

		c.AuthConfig.veriryServerName = ss[0]
		c.AuthConfig.verifyMethod = ss[1]
	}

	return c.check()
}

func init() {
	flag.StringVar(&confPath, "conf", "", "default config path")
}

// Init int config
func Init() error {

	if confPath != "" {
		return Conf.loadFromFile(confPath)
	}

	return errors.New("暂未实现配置中心")

}
