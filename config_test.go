package config_test

import (
	"fmt"

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
