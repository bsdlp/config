package config

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type testConfig struct {
	Example []string `yaml:"example"`
}

type burritoConfig struct {
	Location string `yaml:"location"`
	Burritos bool   `yaml:"burritos"`
}

const (
	correctEnvVar string = "TESTORG_TESTSERVICE_CONFIG_URI"
	systemBaseDir string = "/etc/"
	systemDir     string = "/etc/testorg/testservice/"
	systemPath    string = "/etc/testorg/testservice/config.yaml"
	organization  string = "testorg"
	service       string = "testservice"
	testUsername  string = "testuser"
)

const (
	dirMode  os.FileMode = 0755
	fileMode os.FileMode = 0644
)

var _ = Describe("Config", func() {
	var (
		cfg         Config
		testUser    *user.User
		testHomeDir string
	)

	BeforeEach(func() {
		cfg = Config{
			Organization: organization,
			Service:      service,
		}
		tmpDir, err := ioutil.TempDir("", "config_test")
		Ω(err).Should(BeNil())
		testHomeDir = tmpDir
		testUser = &user.User{
			HomeDir: testHomeDir,
		}
	})

	AfterEach(func() {
		err := os.RemoveAll(testHomeDir)
		Ω(err).Should(BeNil())
	})

	It("can expand paths", func() {
		Ω(testUser.HomeDir).Should(Equal(testHomeDir))
		Ω(expandUser(testUser, "~/")).Should(Equal(testHomeDir))
		Ω(expandUser(testUser, "$HOME/")).Should(Equal(testHomeDir))
		Ω(expandUser(testUser, "~/test")).Should(Equal(filepath.Join(testHomeDir, "test")))
		Ω(expandUser(testUser, "$HOME/test")).Should(Equal(filepath.Join(testHomeDir, "test")))
	})

	It("looks for the right envvar", func() {
		Ω(cfg.EnvVar()).Should(Equal("TESTORG_TESTSERVICE_CONFIG_URI"))
		Ω(Config{Service: "testservice"}.EnvVar()).Should(Equal("TESTSERVICE_CONFIG_URI"))
	})
})
