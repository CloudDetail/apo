package integration

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type CompressMD string

func (c CompressMD) Value() (driver.Value, error) {
	compressed := string(c)
	compressed = strings.ReplaceAll(compressed, "\n", "\\n")
	return compressed, nil
}

func (c *CompressMD) Scan(value interface{}) error {
	if value == nil {
		*c = ""
		return nil
	}
	switch v := value.(type) {
	case []byte:
		str := string(v)
		str = strings.ReplaceAll(str, "\\n", "\n")
		*c = CompressMD(str)
		return nil
	case string:
		// 将特定字符串还原为换行符
		str := strings.ReplaceAll(v, "\\n", "\n")
		*c = CompressMD(str)
		return nil
	default:
		return fmt.Errorf("can not scan %T to CompressMD", value)
	}
}
