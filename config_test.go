package config_test

import (
	"fmt"
	"io/ioutil"
	"log"
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
	var path string
	var cfgNS = config.Namespace{
		Organization: "podhub",
		System:       "canary",
	}

	path, _ = cfgNS.Path()
	fmt.Println("Path to config " + path)

	err = cfgNS.Load(&cfg)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Contents of cfg " + fmt.Sprint(cfg))
}

func TestExpandUser(t *testing.T) {
	var homeDir = os.Getenv("HOME")
	var correctPath = homeDir + "/.config/fly/config/testconfig.yaml"
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
	const correctDir = "/etc/fly/config/"
	const correctPath = correctDir + "testconfig.yaml"
	type configExample struct {
		Location string `yaml:"location"`
		Burritos bool   `yaml:"burritos"`
	}

	var correctCfgText = `location: Señor Sisig
burritos: true`
	var correctCfg = configExample{
		Location: "Señor Sisig",
		Burritos: true,
	}
	var err error
	var cfg configExample
	var homeDir = os.Getenv("HOME")
	var dirMode os.FileMode = 0755
	var fileMode os.FileMode = 0644

	// Setup
	os.RemoveAll(homeDir + "/.config/fly/config/testconfig.yaml")
	os.MkdirAll(correctDir, dirMode)
	err = ioutil.WriteFile(correctPath, []byte(correctCfgText), fileMode)
	if err != nil {
		t.Error("Got an error writing ", correctPath, ", got an error: ", err)
	}

	// Test
	err = config.Load(correctPath, &cfg)
	if err != nil {
		t.Error("Got an error: ", err, ", expecting nil")
	}

	if cfg != correctCfg {
		t.Error("Expecting ", correctCfg, ", got ", cfg)
	}

	// Teardown
	err = os.RemoveAll(correctPath)
	if err != nil {
		t.Error("Unable to remove file ", correctPath,
			" in teardown, got an error: ", err)
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
	const correctSystemBase = "/etc/"
	var systemBase string
	systemBase = config.SystemBase

	if systemBase != correctSystemBase {
		t.Error("Expecting ", correctSystemBase, ", got ", systemBase)
	}
}

func TestNamespacePath(t *testing.T) {
	const correctDir = "/etc/fly/config/"
	const correctPath = correctDir + "testconfig.yaml"
	var cfgNS = config.Namespace{
		Organization: "fly",
		System:       "config",
	}
	var err error
	var path string
	var homeDir = os.Getenv("HOME")
	var dirMode os.FileMode = 0755

	// Setup
	os.RemoveAll(homeDir + "/.config/fly/config/testconfig.yaml")
	os.MkdirAll(correctDir, dirMode)
	_, err = os.Create(correctPath)
	if err != nil {
		t.Error("Unable to create file ", correctPath, ", got an error: ", err)
	}

	// Test
	path, err = cfgNS.Path()
	if err != nil {
		t.Error("Got an error: ", err, " expecting nil")
	}
	if path != correctPath {
		t.Error("Expecting ", correctPath, ", got ", path)
	}

	// Teardown
	err = os.RemoveAll(correctPath)
	if err != nil {
		t.Error("Unable to remove file ", correctPath,
			" in teardown, got an error: ", err)
	}
}

func TestNamespaceLoad(t *testing.T) {
	const correctDir = "/etc/fly/config/"
	const correctPath = correctDir + "testconfig.yaml"
	type configExample struct {
		Location string
		Burritos bool
	}

	var correctCfgText = `location: Señor Sisig
burritos: true`
	var correctCfg = configExample{
		Location: "Señor Sisig",
		Burritos: true,
	}
	var cfgNS = config.Namespace{
		Organization: "fly",
		System:       "config",
	}
	var err error
	var cfg configExample
	var homeDir = os.Getenv("HOME")
	var dirMode os.FileMode = 0755
	var fileMode os.FileMode = 0644

	// Setup
	os.RemoveAll(homeDir + "/.config/fly/config/testconfig.yaml")
	os.MkdirAll(correctDir, dirMode)
	err = ioutil.WriteFile(correctPath, []byte(correctCfgText), fileMode)
	if err != nil {
		t.Error("Got an error writing ", correctPath, ", got an error: ", err)
	}

	// Test
	err = cfgNS.Load(&cfg)
	if err != nil {
		t.Error("Got an error: ", err, ", expecting nil")
	}

	if cfg != correctCfg {
		t.Error("Expecting ", correctCfg, ", got ", cfg)
	}

	// Teardown
	err = os.RemoveAll(correctPath)
	if err != nil {
		t.Error("Unable to remove file ", correctPath,
			" in teardown, got an error: ", err)
	}
}
