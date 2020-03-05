package popeye

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	// "github.com/derailed/popeye/internal/report"
	"github.com/derailed/popeye/pkg"
	"github.com/derailed/popeye/pkg/config"
)

type Plugin struct {
	Name string
}

func NewPlugin() *pkg.Popeye {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	flags := config.NewFlags()


	f := true
	flags.AllNamespaces = &f
	pop, err := pkg.NewPopeye(flags, &log.Logger)
	if err != nil {
		return nil
	}
	if err := pop.Init(); err != nil {
		return nil
	}
	if err := pop.Sanitize(); err != nil {
		return nil
	}
	return pop
}
