package fileformat

import (
	"encoding/json"

	"github.com/BurntSushi/toml"
	"github.com/bsdlp/config"
	"github.com/bsdlp/config/fileformat/ini"
	"github.com/hashicorp/hcl"
	"gopkg.in/yaml.v2"
)

// YAML is a FileFormat for yaml
var YAML = &config.FileFormat{
	Unmarshaller: yaml.Unmarshal,
	Extension:    "yaml",
}

// JSON is FileFormat for json
var JSON = &config.FileFormat{
	Unmarshaller: json.Unmarshal,
	Extension:    "json",
}

// TOML is FileFormat for toml
var TOML = &config.FileFormat{
	Unmarshaller: toml.Unmarshal,
	Extension:    "toml",
}

// HCL is FileFormat for hcl
var HCL = &config.FileFormat{
	Unmarshaller: hcl.Unmarshal,
	Extension:    "hcl",
}

// INI is FileFormat for ini
var INI = &config.FileFormat{
	Unmarshaller: ini.Unmarshal,
	Extension:    "ini",
}
