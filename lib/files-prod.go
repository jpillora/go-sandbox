package sandbox

import "fmt"

func Asset(name string) ([]byte, error) {
	return nil, fmt.Errorf("Asset %s not found", name)
}
