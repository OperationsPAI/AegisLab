package gateway

import "fmt"

func missingRemoteDependency(name string) error {
	return fmt.Errorf("%s remote client is not configured for api-gateway", name)
}
