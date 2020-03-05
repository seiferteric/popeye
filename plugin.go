package popeye

import (
	"github.com/rs/zerolog"
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


	flags.AllNamespaces = true
	pop, err := pkg.NewPopeye(flags, &log.Logger)
	if err != nil {
		bomb(fmt.Sprintf("Popeye configuration load failed %v", err))
	}
	if err := pop.Init(); err != nil {
		bomb(err.Error())
	}
	if err := pop.Sanitize(); err != nil {
		bomb(err.Error())
	}
	return pop
}
