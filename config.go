package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"strings"

	"gopkg.in/yaml.v2"
)

type ConfigNamespace struct {
	Organization string // optional additional namespace for orgs.
	Namespace    string // usually project name.
}

type Config interface {
	Load(src string, dst interface{}) error
}

const UserBase string = "~/.config/"
const EtcDir string = "/etc/"

// Load expands the provided src path using config.ExpandUser, then reads
// the file and unmarshals into dst using go-yaml.
func Load(src string, dst interface{}) (err error) {
	dstv := reflect.ValueOf(dst)

	if dstv.Kind() != reflect.Ptr {
		err = errors.New("config: not a pointer.")
	} else if dstv.IsNil() {
		err = fmt.Errorf("nil %s.", reflect.TypeOf(dstv).String())
	}
	if err != nil {
		return
	}

	path, err := ExpandUser(src)
	if err != nil {
		return
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(data, dst)
	return
}

func ExpandUser(path string) (exPath string, err error) {
	// Acts kind of like os.path.expanduser in Python, except only supports
	// expanding "~/" or "$HOME"
	usr, err := user.Current()
	if err != nil {
		return
	}

	dir := fmt.Sprintf("%s/", usr.HomeDir)

	exPath = strings.Replace(path, "~/", dir, 1)
	exPath = strings.Replace(exPath, "$HOME", dir, 1)

	if err != nil {
		return "", err
	}

	exPath, err = filepath.Abs(filepath.Clean(exPath))

	if err != nil {
		return "", err
	}

	return
}

// Returns path to config, chosen by hierarchy and checked for existence:
// 1. User config (~/.config/podhub/canary/config.yaml)
// 2. System config (/etc/podhub/canary/config.yaml)
func (c ConfigNamespace) Path() (path string, err error) {
	systemPath, _ := c.systemPath()
	if _, err := os.Stat(systemPath); err == nil {
		path, _ = c.systemPath()
	}

	userPath, _ := c.userPath()
	if _, err := os.Stat(userPath); err == nil {
		path, _ = c.userPath()
	}
	return
}

func (c ConfigNamespace) systemPath() (path string, err error) {
	path = filepath.Join(EtcDir, c.Organization, c.Namespace, "config.yaml")
	return
}

func (c ConfigNamespace) userPath() (path string, err error) {
	userBase, err := ExpandUser(UserBase)
	if err != nil {
		return "", err
	}

	path = filepath.Join(userBase, c.Organization, c.Namespace, "config.yaml")
	return
}

func (c ConfigNamespace) Load(dst interface{}) (err error) {
	cfgPath, err := c.Path()
	if err != nil {
		return
	}

	err = Load(cfgPath, dst)
	return
}
