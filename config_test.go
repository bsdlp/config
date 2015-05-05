package config_test

import (
	"fmt"
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
	var cfgNS = config.Namespace{
		Organization: "podhub",
		System:       "canary",
	}

	type Config struct {
		Example []string `yaml:"example"`
	}

	var cfg Config

	path, err := cfgNS.Path()
	fmt.Println("Path to config " + path)

	err := cfgNS.Load(&cfg)
	fmt.Println("Contents of cfg " + fmt.Sprint(cfg))
	// Output: Path to config: /Users/jchen/.config/podhub/canary/config.yaml
	// Output: Contents of cfg: {[a b c]}
}

func TestExpandUser(t *testing.T) {
	var path string
	correctPath := "/home/travis/.config/fly/config/config.yaml"
	path, err := config.ExpandUser("~/.config/fly/config/config.yaml")

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

	var err error
	var cfg configExample
	var correctCfg = configExample{
		Location: "Señor Sisig",
		Burritos: true,
	}

	err = config.Load("/home/travis/.config/fly/config/config.yaml", &cfg)
	if err != nil {
		t.Error("Got an error: ", err, ", expecting nil")
	}

	if cfg != correctCfg {
		t.Error("Got ", cfg, ", expecting ", correctCfg)
	}
}
