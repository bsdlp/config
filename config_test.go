package config_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"testing"

	"github.com/bsdlp/config"
)

type testConfig struct {
	Example []string `yaml:"example"`
}

type burritoConfig struct {
	Location string `yaml:"location"`
	Burritos bool   `yaml:"burritos"`
}

// TODO: come up with a way to test homedir without hardcoding travis
const (
	correctEnvVar string = "TESTORG_TESTSYSTEM_CONFIG_URI"
	systemBaseDir string = "/etc/"
	systemDir     string = "/etc/testorg/testsystem/"
	systemPath    string = "/etc/testorg/testsystem/config.yaml"
	userBaseDir   string = "/home/travis/.config/"
	userDir       string = "/home/travis/"
	userPath      string = "/home/travis/.config/testorg/testsystem/config.yaml"
	organization  string = "testorg"
	system        string = "testsystem"
)

const (
	dirMode  os.FileMode = 0755
	fileMode os.FileMode = 0644
)

var cfgNS = config.Namespace{
	Organization: organization,
	System:       system,
}

// In this example our organization is named "testorganization", and our project
// namespace is "testsystem".
//
// In this example we have a file located at
// /Users/jchen/.config/testorg/testsystem/config.yaml with the
// following contents:
//  example:
//    - "a"
//    - "b"
//    - "c"
func ExampleNamespace() {
	var err error
	var cfg testConfig
	var path string

	path = cfgNS.Path()
	fmt.Println("Path to config " + path)

	err = cfgNS.Load(&cfg)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Contents of cfg " + fmt.Sprint(cfg))
}

func TestExpandUser(t *testing.T) {
	var path string
	correctPath := userPath

	path = config.ExpandUser("~/.config/testorg/testsystem/config.yaml")

	if path != correctPath {
		t.Error("Expected ", correctPath, ", got ", path)
	}

	path = config.ExpandUser("$HOME/.config/testorg/testsystem/config.yaml")

	if path != correctPath {
		t.Error("Expected ", correctPath, ", got ", path)
	}
}

func TestUserBase(t *testing.T) {
	userbase := "~/.config/"
	if config.UserBase != userbase {
		t.Error("Expecting ", userbase, ", got ", config.UserBase)
	}
}

func TestSystemBase(t *testing.T) {
	if config.SystemBase != systemBaseDir {
		t.Error("Expecting ", systemBaseDir, ", got ", config.SystemBase)
	}
}

func TestNamespacePath(t *testing.T) {
	var err error
	var path string

	correctDir := systemDir
	correctPath := systemPath

	// Setup
	err = os.RemoveAll(userPath)
	if err != nil {
		t.Error("Unable to remove file ", correctPath, ", got an error: ", err)
	}

	err = os.MkdirAll(correctDir, dirMode)
	if err != nil {
		t.Error("Unable to create directory ", correctPath, ", got an error: ", err)
	}

	_, err = os.Create(correctPath)
	if err != nil {
		t.Error("Unable to create file ", correctPath, ", got an error: ", err)
	}

	// Test
	path = cfgNS.Path()
	if path != correctPath {
		t.Error("Expecting ", correctPath, ", got ", path)
	}

	// Teardown
	err = os.RemoveAll(correctPath)
	if err != nil {
		t.Error("Unable to remove file ", correctPath,
			" in teardown, got an error: ", err)
	}

	// test homedir
	// Setup
	correctDir = userDir
	correctPath = userPath
	err = os.RemoveAll(systemPath)
	if err != nil {
		t.Error("Unable to remove file ", correctPath, ", got an error: ", err)
	}

	err = os.MkdirAll(correctDir, dirMode)
	if err != nil {
		t.Error("Unable to create directory ", correctPath, ", got an error: ", err)
	}

	_, err = os.Create(correctPath)
	if err != nil {
		t.Error("Unable to create file ", correctPath, ", got an error: ", err)
	}

	// Test
	path = cfgNS.Path()
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

func TestNamespaceEnvVar(t *testing.T) {
	if cfgNS.EnvVar() != correctEnvVar {
		t.Error("Expecting ", correctEnvVar, ", got ", cfgNS.EnvVar())
	}
}

func TestNewConfigFromNamespace(t *testing.T) {
	const (
		organization = "testorg"
		system       = "testsystem"
	)

	var err error

	correctDir := systemDir
	correctPath := systemPath

	// Setup
	err = os.RemoveAll(userPath)
	if err != nil {
		t.Error("Unable to remove file ", userPath, ", got an error: ", err)
	}

	err = os.MkdirAll(correctDir, dirMode)
	if err != nil {
		t.Error("Unable to create directory ", correctPath, ", got an error: ", err)
	}

	_, err = os.Create(correctPath)
	if err != nil {
		t.Error("Unable to create file ", correctPath, ", got an error: ", err)
	}

	err = os.RemoveAll(correctPath)
	if err != nil {
		t.Error("Unable to remove file ", correctPath,
			" in teardown, got an error: ", err)
	}
}

func TestNamespaceLoad(t *testing.T) {
	correctDir := systemDir
	correctPath := systemPath

	correctCfgText := `location: Señor Sisig
burritos: true`
	correctCfg := burritoConfig{
		Location: "Señor Sisig",
		Burritos: true,
	}
	var err error
	var cfg burritoConfig

	// Setup
	err = os.RemoveAll(userPath)
	if err != nil {
		t.Error("Unable to remove path ", correctPath, ", got an error: ", err)
	}

	err = os.MkdirAll(correctDir, dirMode)
	if err != nil {
		t.Error("Unable to create directory ", correctPath, ", got an error: ", err)
	}

	err = ioutil.WriteFile(correctPath, []byte(correctCfgText), fileMode)
	if err != nil {
		t.Error("Got an error writing ", correctPath, ": ", err)
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
