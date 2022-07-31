package hash

import (
	"hash/crc32"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// StringKey returns a schema.SchemaSetFunc able to hash a string value
// from map accessed by k.
func StringKey(k string) schema.SchemaSetFunc {
	return func(v interface{}) int {
		m, ok := v.(map[string]interface{})
		if !ok {
			return 0
		}
		if v, ok := m[k].(string); ok {
			return stringToHashCode(v)
		}
		return 0
	}
}

func stringToHashCode(s string) int {
	v := int(crc32.ChecksumIEEE([]byte(s)))
	if v >= 0 {
		return v
	}
	if -v >= 0 {
		return -v
	}
	return 0
}
