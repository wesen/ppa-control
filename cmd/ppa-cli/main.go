package main

import (
	"ppa-control/cmd/ppa-cli/cmds"
	logger "ppa-control/lib/log"
)

func main() {
	logger.InitializeLogger()

	cmds.Execute()

}
