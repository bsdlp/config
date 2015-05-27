// Package config version v1.0.0
package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
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

// Load reads the contents of the src URI and unmarshals into dst using
// go-yaml.
func Load(src string, dst interface{}) (err error) {
	dstv := reflect.ValueOf(dst)

	if dstv.Kind() != reflect.Ptr {
		err = errors.New("config: not a pointer")
		return
	}

	uri, err := url.Parse(src)
	if err != nil {
		return err
	}

	var data []byte
	switch {
	case uri.Scheme == "file" || uri.Scheme == "":
		path := ExpandUser(uri.Path)
		data, err = ioutil.ReadFile(path)
		if err != nil {
			return err
		}
	case uri.Scheme == "http" || uri.Scheme == "https":
		resp, err := http.Get(uri.String())
		if err != nil {
			return err
		}

		data, err = ioutil.ReadAll(resp.Body)

		resp.Body.Close()
		if err != nil {
			return err
		}
	}

	err = yaml.Unmarshal(data, dst)
	return
}

// ExpandUser acts kind of like os.path.expanduser in Python, except only
// supports expanding "~/" or "$HOME"
func ExpandUser(path string) (exPath string) {
	usr, _ := user.Current()

	dir := fmt.Sprintf("%s/", usr.HomeDir)

	exPath = path
	if len(path) > 2 && path[:2] == "~/" {
		exPath = strings.Replace(exPath, "~/", dir, 1)
	} else if len(path) > 5 && path[:5] == "$HOME" {
		exPath = strings.Replace(exPath, "$HOME", dir, 1)
	}

	exPath, _ = filepath.Abs(filepath.Clean(exPath))
	return
}

// Path returns path to config, chosen by hierarchy and checked for
// existence:
//
// 1. User config (~/.config/podhub/canary/config.yaml)
//
// 2. System config (/etc/podhub/canary/config.yaml)
func (c Namespace) Path() (path string) {
	systemPath := c.systemURI().Path
	if _, err := os.Stat(systemPath); err == nil {
		path = systemPath
	}

	userPath := c.userURI().Path
	if _, err := os.Stat(userPath); err == nil {
		path = userPath
	}
	return
}

func (c Namespace) systemURI() (uri url.URL) {
	path := filepath.Join(SystemBase, c.Organization, c.System, "config.yaml")
	uri = url.URL{Path: path, Scheme: "file"}
	return
}

func (c Namespace) userURI() (uri url.URL) {
	userBase := ExpandUser(UserBase)

	path := filepath.Join(userBase, c.Organization, c.System, "config.yaml")
	uri = url.URL{Path: path, Scheme: "file"}
	return
}

// EnvVar returns the name of the environment variable containing the URI
// of the config.
// Example: PODHUB_UUIDD_CONFIG_URI
func (c Namespace) EnvVar() (envvar string) {
	s := []string{c.Organization, c.System, "CONFIG", "URI"}
	envvar = strings.ToUpper(strings.Join(s, "_"))
	return
}

// Load is a convenience function registered to config.Namespace to
// implement Config.Load().
func (c Namespace) Load(dst interface{}) (err error) {
	cfgPath := c.Path()

	err = Load(cfgPath, dst)
	return
}
