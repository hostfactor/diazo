package envvar

import "os"

const (
	HostFactorServerIdKey = "HOST_FACTOR_SERVER_ID"
)

var ServerId = os.Getenv(HostFactorServerIdKey)
