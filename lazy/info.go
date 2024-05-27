package lazy

import "github.com/dogsays/mo/logger"

var (
	BUILD_TIME = ""
	GO_VERSION = ""
	AUTHOR     = ""
)

func printInfo() {
	logger.Info("BUILD_TIME", BUILD_TIME)
	logger.Info("GO_VERSION", GO_VERSION)
	logger.Info("AUTHOR", AUTHOR)
}
