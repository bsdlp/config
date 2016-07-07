package config

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"os/user"
	"path/filepath"

	"gopkg.in/yaml.v2"

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
	correctEnvVar  string = "TESTORG_TESTSERVICE_CONFIG_URI"
	systemBaseDir  string = "/etc/"
	systemDir      string = "/etc/testorg/testservice/"
	systemPath     string = "/etc/testorg/testservice/config.yaml"
	organization   string = "testorg"
	service        string = "testservice"
	testUsername   string = "testuser"
	yamlExtension  string = "yaml"
	testConfigData string = `---
has_burrito: true
favorite_hero: roadhog`
)

type configData struct {
	HasBurrito   bool   `yaml:"has_burrito"`
	FavoriteHero string `yaml:"favorite_hero"`
}

var testConfigDataUnmarshalled = &configData{
	HasBurrito:   true,
	FavoriteHero: "roadhog",
}

const (
	dirMode  os.FileMode = 0755
	fileMode os.FileMode = 0644
)

func testUnmarshaller(data []byte, v interface{}) (err error) {
	return
}

var _ = Describe("Config", func() {
	var (
		cfg         Config
		testUser    *user.User
		testHomeDir string
		tmpDir      string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "config_test")
		Ω(err).Should(BeNil())
		testHomeDir = tmpDir
		testUser = &user.User{
			HomeDir: testHomeDir,
		}

		cfg = Config{
			Organization: organization,
			Service:      service,
			FileFormat: &FileFormat{
				Extension:    yamlExtension,
				Unmarshaller: yaml.Unmarshal,
			},
			pathExpander: func(p string) string { return expandUser(testUser, p) },
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

	It("returns the right file name", func() {
		Ω(cfg.fileName()).Should(Equal("config.yaml"))
	})

	It("returns a default file name if FileFormat is nil", func() {
		cfg.FileFormat = nil
		Ω(cfg.fileName()).Should(Equal("config"))
	})

	It("returns a default file name if file extension is empty", func() {
		cfg.FileFormat.Extension = ""
		Ω(cfg.fileName()).Should(Equal("config"))
	})

	It("returns the right path for the system config", func() {
		Ω(cfg.systemURI().Path).Should(Equal(systemPath))
	})

	It("returns the right path for the user config", func() {
		Ω(cfg.userURI().Path).Should(Equal(filepath.Join(testHomeDir, ".config", organization, service, "config.yaml")))
	})

	Describe("Loader", func() {
		var (
			ts *httptest.Server
			f  *os.File
		)
		BeforeEach(func() {
			ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				_, err := fmt.Fprint(w, testConfigData)
				Ω(err).Should(BeNil())
			}))

			var err error
			f, err = ioutil.TempFile(tmpDir, "")
			Ω(err).Should(BeNil())
			err = ioutil.WriteFile(f.Name(), []byte(testConfigData), 0640)
			Ω(err).Should(BeNil())
		})

		AfterEach(func() {
			ts.Close()
			err := f.Close()
			Ω(err).Should(BeNil())
		})

		It("parses http correctly", func() {
			data, parseErr := uriParser(ts.URL)
			Ω(parseErr).Should(BeNil())
			Ω(data).Should(Equal([]byte(testConfigData)))
		})

		It("parses file correctly", func() {
			data, parseErr := uriParser(f.Name())
			Ω(parseErr).Should(BeNil())
			Ω(data).Should(Equal([]byte(testConfigData)))
		})

		It("checks to see if unmarshaller is set correctly", func() {
			td := new(configData)
			err := load(nil, f.Name(), td)
			Ω(err).Should(Equal(ErrNilUnmarshaller))
		})

		It("checks to make sure dst is a pointer", func() {
			td := new(configData)
			err := load(yaml.Unmarshal, f.Name(), *td)
			Ω(err).Should(Equal(ErrNotAPointer))
		})

		It("loads config data", func() {
			td := new(configData)
			err := load(yaml.Unmarshal, f.Name(), td)
			Ω(err).Should(BeNil())
			Ω(td).Should(Equal(testConfigDataUnmarshalled))
		})
	})
})
