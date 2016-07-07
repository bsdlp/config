// Package config version v1.0.0
package config

import (
	"errors"
	"fmt"
	"io"
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

// FileFormat is the type of config file to unmarshal
type FileFormat struct {
	// file extension, i.e. "yaml" for a file named "config.yaml"
	Extension string

	// Unmarshaller used to unmarshal config data
	Unmarshaller Unmarshaller
}

// Config implements Loader
type Config struct {
	// optional additional namespace for orgs.
	Organization string

	// Name of the service associated with this config.
	Service string

	// describes the type of config file to unmarshal
	FileFormat *FileFormat

	// used for mocking expanduser
	pathExpander func(p string) string
}

var (
	// ErrNilUnmarshaller is returned when an undefined unmarshaller is passed to
	// load()
	ErrNilUnmarshaller = errors.New("config: nil unmarshaller")

	// ErrNilFileFormat is returned when the config struct contains a nil FileFormat
	ErrNilFileFormat = errors.New("config: file format undefined")

	// ErrConfigFileNotFound is returned when config files at $HOME/:organization/:system/config.{extension}
	// or /etc/:organization/:systems/config.{extension} are missing
	ErrConfigFileNotFound = errors.New("config: missing config files")

	// ErrNotAPointer is returned when a non-pointer is passed into load
	ErrNotAPointer = errors.New("config: not a pointer")
)

// readHTTP reads *http.Response.ContentLength bytes of *http.Response.Body
// and returns as []byte
func readHTTP(uri string) (data []byte, err error) {
	resp, err := http.Get(uri)
	if err != nil {
		return
	}

	defer func() {
		closeErr := resp.Body.Close()
		if err == nil {
			err = closeErr
		}
		return
	}()

	data = make([]byte, 0, resp.ContentLength)
	bytesRead, err := io.ReadFull(resp.Body, data)
	if err != nil {
		return
	}
	if int64(bytesRead) != resp.ContentLength {
		err = fmt.Errorf("config: incomplete http response read: %d bytes read of content-length %d", bytesRead, resp.ContentLength)
		return
	}
	return
}

func uriParser(src string) (data []byte, err error) {
	uri, err := url.Parse(src)
	if err != nil {
		return
	}

	switch {
	case uri.Scheme == "file" || uri.Scheme == "":
		data, err = ioutil.ReadFile(uri.Path)
		if err != nil {
			return
		}
		return
	case uri.Scheme == "http" || uri.Scheme == "https":
		data, err = readHTTP(uri.String())
		if err != nil {
			return
		}
		return
	}
	return
}

// load reads the contents of the file at the provided src uri and uses the
// provided unmarshaller to
func load(unmarshaller Unmarshaller, src string, dst interface{}) (err error) {
	if unmarshaller == nil {
		err = ErrNilUnmarshaller
		return
	}

	if reflect.ValueOf(dst).Kind() != reflect.Ptr {
		err = ErrNotAPointer
		return
	}

	data, err := uriParser(src)
	if err != nil {
		return
	}

	err = unmarshaller(data, dst)
	if err != nil {
		return
	}
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
// 2. User config (~/.config/podhub/canary/config.{extension})
//
// 3. System config (/etc/podhub/canary/config.{extension})
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

const fileNamePrefix = "config"

func (c Config) fileName() string {
	if c.FileFormat == nil || c.FileFormat.Extension == "" {
		return fileNamePrefix
	}
	return fmt.Sprintf("%s.%s", fileNamePrefix, c.FileFormat.Extension)
}

func (c Config) systemURI() (uri *url.URL) {
	path := filepath.Join(SystemBase, c.Organization, c.Service, c.fileName())
	uri = &url.URL{Path: path, Scheme: "file"}
	return
}

func (c Config) userURI() (uri *url.URL) {
	var userBase string
	if c.pathExpander == nil {
		userBase = ExpandUser(UserBase)
	} else {
		userBase = c.pathExpander(UserBase)
	}

	path := filepath.Join(userBase, c.Organization, c.Service, c.fileName())
	uri = &url.URL{Path: path, Scheme: "file"}
	return
}

// EnvVar returns the name of the environment variable containing the URI
// of the config.
// Example: PODHUB_UUIDD_CONFIG_URI
func (c Config) EnvVar() (envvar string) {
	var s []string
	if c.Organization == "" {
		s = []string{c.Service, "CONFIG", "URI"}
	} else {
		s = []string{c.Organization, c.Service, "CONFIG", "URI"}
	}
	envvar = strings.ToUpper(strings.Join(s, "_"))
	return
}

// Load is a convenience function registered to config.Namespace to
// implement Config.Load().
func (c Config) Load(dst interface{}) (err error) {
	if c.FileFormat == nil {
		err = ErrNilFileFormat
		return
	}

	cfgPath := c.Path()

	if cfgPath == "" {
		err = ErrConfigFileNotFound
		return
	}

	err = load(c.FileFormat.Unmarshaller, cfgPath, dst)
	return
}
