package config_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/fly/config"
)

// In this example our organization is named "podhub", and our project
// namespace is "canary".
//
// In this example we have a file located at /Users/jchen/.config/podhub/canary/config.yaml,
// with the following contents:
//  example:
//    - "a"
//    - "b"
//    - "c"
func ExampleNamespace() {
	type Config struct {
		Example []string `yaml:"example"`
	}

	var err error
	var cfg Config
	var cfgNS = config.Namespace{
		Organization: "podhub",
		System:       "canary",
	}

	path, _ = cfgNS.Path()
	fmt.Println("Path to config " + path)

	err = cfgNS.Load(&cfg)
	fmt.Println("Contents of cfg " + fmt.Sprint(cfg))
	// Output: Path to config: /Users/jchen/.config/podhub/canary/config.yaml
	// Output: Contents of cfg: {[a b c]}
}

func TestExpandUser(t *testing.T) {
	const correctPath = "/home/travis/.config/fly/config/testconfig.yaml"
	var err error
	var path string
	path, err = config.ExpandUser("~/.config/fly/config/testconfig.yaml")

	if err != nil {
		t.Error("Got an error: ", err, ", expecting nil")
	}

	// docs say not to trust /home/travis to be homedir. We'll need to
	// revisit this later.
	if path != correctPath {
		t.Error("Expected ", correctPath, ", got ", path)
	}
}

func TestLoad(t *testing.T) {
	type configExample struct {
		Location string
		Burritos bool
	}

	var correctCfg = configExample{
		Location: "Señor Sisig",
		Burritos: true,
	}
	var err error
	var cfg configExample

	err = config.Load("/home/travis/.config/fly/config/testconfig.yaml", &cfg)
	if err != nil {
		t.Error("Got an error: ", err, ", expecting nil")
	}

	if cfg != correctCfg {
		t.Error("Expecting ", correctCfg, ", got ", cfg)
	}
}

func TestUserBase(t *testing.T) {
	const correctUserBase = "~/.config/"
	var userBase string
	userBase = config.UserBase

	if userBase != correctUserBase {
		t.Error("Expecting ", correctUserBase, ", got ", userBase)
	}
}

func TestSystemBase(t *testing.T) {
	var systemBase string
	var correctSystemBase = "~/.config/"
	systemBase = config.SystemBase

	if systemBase != correctSystemBase {
		t.Error("Expecting ", correctSystemBase, ", got ", systemBase)
	}
}
