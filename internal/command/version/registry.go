package version

import "github.com/ElshadHu/verdis/internal/command"

// RegisterAll adds all command specs into the router.
func RegisterAll(router *command.Router) {
	router.Register(GetVersionSpec())
	router.Register(HistorySpec())
}
