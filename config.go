package config

import (
	"errors"
	"os/user"
	"path/filepath"
	"strings"
)

type ConfigNamespace struct {
	Organization string // optional additional namespace for orgs.
	Namespace    string // usually project name.
}

type Config interface {
	Load() error
}

const UserBase string = "~/.config/"

func ExpandUser(path string) (exPath string, err error) {
	// Acts kind of like os.path.expanduser in Python, except only supports
	// expanding "~/" or "$HOME"
	usr, err := user.Current()
	if err != nil {
		return
	}

	dir := usr.HomeDir

	if path[:2] == "~/" {
		exPath = strings.Replace(path, "~/", dir, 1)
	} else if path[:5] == "$HOME" {
		exPath = strings.Replace(path, "$HOME", dir, 1)
	} else {
		err = errors.New("No expandable path provided.")
	}

	if err != nil {
		return "", err
	}

	exPath, err = filepath.Abs(filepath.Clean(exPath))

	if err != nil {
		return "", err
	}

	return
}

func (c ConfigNamespace) Path() (path string, err error) {
	userBase, err := ExpandUser(UserBase)
	if err != nil {
		return "", err
	}

	path = filepath.Join(userBase, c.Organization, c.Namespace, "config.yaml")
	return
}
