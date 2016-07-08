package ini

import "github.com/go-ini/ini"

// Unmarshal implements config.Unmarshaller for ini
func Unmarshal(data []byte, v interface{}) error { return ini.MapTo(v, data) }
