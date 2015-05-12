// Package config version v1.0.0
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

// Namespace holds the pieces joined together to form the
// directory namespacing for config.
type Namespace struct {
	Organization string // optional additional namespace for orgs.
	System       string // Name of the system associated with this config.
}

// The Config interface implements config.
type Config interface {
	Load(src string, dst interface{}) error
}

// UserBase and SystemBase are the prefixes for the user and system config
// paths, respectively.
const (
	UserBase   string = "~/.config/"
	SystemBase string = "/etc/"
)

// Load expands the provided src path using config.ExpandUser, then reads
// the file and unmarshals into dst using go-yaml.
func Load(src string, dst interface{}) (err error) {
	dstv := reflect.ValueOf(dst)

	if dstv.Kind() != reflect.Ptr {
		err = errors.New("config: not a pointer")
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

// ExpandUser acts kind of like os.path.expanduser in Python, except only
// supports expanding "~/" or "$HOME"
func ExpandUser(path string) (exPath string, err error) {
	usr, err := user.Current()
	if err != nil {
		return
	}

	dir := fmt.Sprintf("%s/", usr.HomeDir)

	exPath = path
	if len(path) > 2 && path[:2] == "~/" {
		exPath = strings.Replace(exPath, "~/", dir, 1)
	} else if len(path) > 5 && path[:5] == "$HOME" {
		exPath = strings.Replace(exPath, "$HOME", dir, 1)
	}

	exPath, err = filepath.Abs(filepath.Clean(exPath))

	if err != nil {
		return "", err
	}

	return
}

// Path returns path to config, chosen by hierarchy and checked for
// existence:
//
// 1. User config (~/.config/podhub/canary/config.yaml)
//
// 2. System config (/etc/podhub/canary/config.yaml)
func (c Namespace) Path() (path string) {
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

func (c Namespace) systemPath() (path string, err error) {
	path = filepath.Join(SystemBase, c.Organization, c.System, "config.yaml")
	return
}

func (c Namespace) userPath() (path string, err error) {
	userBase, err := ExpandUser(UserBase)
	if err != nil {
		return "", err
	}

	path = filepath.Join(userBase, c.Organization, c.System, "config.yaml")
	return
}

// Load is a convenience function registered to config.Namespace to
// implement Config.Load().
func (c Namespace) Load(dst interface{}) (err error) {
	cfgPath := c.Path()

	err = Load(cfgPath, dst)
	return
}
