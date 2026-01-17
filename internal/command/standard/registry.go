package standard

import "github.com/ElshadHu/verdis/internal/command"

// RegisterAll adds all command specs into the router.
func RegisterAll(router *command.Router) {
	router.Register(PingSpec())
	router.Register(GetSpec())
	router.Register(SetSpec())
	router.Register(DelSpec())
	router.Register(ExistsSpec())
}
