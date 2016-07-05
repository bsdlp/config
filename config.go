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
)

// UserBase and SystemBase are the prefixes for the user and system config
// paths, respectively.
const (
	UserBase   string = "~/.config/"
	SystemBase string = "/etc/"
)

// Unmarshaller defines the function signature for unmarshal functions
type Unmarshaller func(data []byte, v interface{}) error

// Config implements Loader
type Config struct {
	// optional additional namespace for orgs.
	Organization string

	// Name of the service associated with this config.
	Service string

	// Unmarshaller used to unmarshal config data
	Unmarshaller Unmarshaller
}

var (
	// ErrNilUnmarshaller is returned when an undefined unmarshaller is passed to
	// load()
	ErrNilUnmarshaller = errors.New("config: nil unmarshaller")

	// ErrConfigFileNotFound is returned when config files at $HOME/:organization/:system/config.yaml
	// or /etc/:organization/:systems/config.yaml are missing
	ErrConfigFileNotFound = errors.New("config: missing config files")
)

// load reads the contents of the file at the provided src uri and uses the
// provided unmarshaller to
func load(unmarshaller Unmarshaller, src string, dst interface{}) (err error) {
	if unmarshaller == nil {
		err = ErrNilUnmarshaller
		return
	}

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
		data, err = ioutil.ReadFile(uri.Path)
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

	err = unmarshaller(data, dst)
	return
}

func expandUser(usr *user.User, path string) (exPath string) {
	dir := fmt.Sprintf("%s/", usr.HomeDir)

	exPath = path
	if len(path) >= 2 && path[:2] == "~/" {
		exPath = strings.Replace(exPath, "~/", dir, 1)
	} else if len(path) >= 5 && path[:5] == "$HOME" {
		exPath = strings.Replace(exPath, "$HOME", dir, 1)
	}

	exPath, _ = filepath.Abs(filepath.Clean(exPath))
	return
}

// ExpandUser acts kind of like os.path.expanduser in Python, except only
// supports expanding "~/" or "$HOME"
func ExpandUser(path string) (exPath string) {
	usr, err := user.Current()
	if usr == nil || err != nil {
		return
	}
	exPath = expandUser(usr, path)
	return
}

// Path returns path to config, chosen by hierarchy and checked for
// existence:
//
// 1. {ORGANIZATION}_{SERVICE}_CONFIG_URI environment variable
//
// 2. User config (~/.config/podhub/canary/config.yaml)
//
// 3. System config (/etc/podhub/canary/config.yaml)
func (c Config) Path() (path string) {
	envVarPath := c.EnvVar()
	if envVarPath != "" {
		if _, err := os.Stat(envVarPath); err == nil {
			path = envVarPath
			return
		}
	}

	userPath := c.userURI().Path
	if _, err := os.Stat(userPath); err == nil {
		path = userPath
		return
	}

	systemPath := c.systemURI().Path
	if _, err := os.Stat(systemPath); err == nil {
		path = systemPath
		return
	}
	return
}

func (c Config) systemURI() (uri url.URL) {
	path := filepath.Join(SystemBase, c.Organization, c.Service, "config.yaml")
	uri = url.URL{Path: path, Scheme: "file"}
	return
}

func (c Config) userURI() (uri url.URL) {
	userBase := ExpandUser(UserBase)

	path := filepath.Join(userBase, c.Organization, c.Service, "config.yaml")
	uri = url.URL{Path: path, Scheme: "file"}
	return
}

// EnvVar returns the name of the environment variable containing the URI
// of the config.
// Example: PODHUB_UUIDD_CONFIG_URI
func (c Config) EnvVar() (envvar string) {
	s := []string{c.Organization, c.Service, "CONFIG", "URI"}
	envvar = strings.ToUpper(strings.Join(s, "_"))
	return
}

// Load is a convenience function registered to config.Namespace to
// implement Config.Load().
func (c Config) Load(dst interface{}) (err error) {
	cfgPath := c.Path()

	if cfgPath == "" {
		err = ErrConfigFileNotFound
		return
	}

	err = load(c.Unmarshaller, cfgPath, dst)
	return
}
