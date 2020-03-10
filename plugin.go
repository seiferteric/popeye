package popeye

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	// "github.com/derailed/popeye/internal/report"
	"fmt"
	"github.com/derailed/popeye/internal/issues"
	"github.com/derailed/popeye/pkg"
	"github.com/derailed/popeye/pkg/config"
	"strconv"
	"strings"
)

type PopeyePlugin struct {
	Name string
	Pop  *pkg.Popeye
}

func NewPlugin() *PopeyePlugin {
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
	pop.Builder.ToJSON()

	plug := PopeyePlugin{}
	plug.Pop = pop
	return &plug
}

func (p *PopeyePlugin) PrintErrors() {
	// for _,s := range p.Pop.Builder.Report.Sections {
	// 	for _,i := range s.Outcome {
	// 		for _,k := range i {
	// 			fmt.Println(k.Message)
	// 		}
	// 	}
	// }
	codes, _ := issues.LoadCodes()
	for _, s := range p.Pop.Builder.Report.Sections {

		for k, v := range s.Outcome {
			for _, issue := range v {
				if issue.Level == 3 {
					// fmt.Println("Error: ", s.Title, k, issue.Message)
					PopeyToAction(s.Title, k, issue.Message, codes)
				}
			}
		}
	}

}

func PopeyCodeFromMsg(msg string) (config.ID, string, error) {
	if msg[0] != '[' {
		return 0, msg, fmt.Errorf("No Code Found")
	}
	stop := strings.IndexByte(msg, ']')
	imsg, err := strconv.Atoi(msg[5:stop])
	if err != nil {
		return 0, msg, err
	}
	return config.ID(imsg), msg[stop+2:], nil
}
func PopeyToAction(section string, item string, msg string, codes *issues.Codes) {

	code, new_msg, err := PopeyCodeFromMsg(msg)
	if err == nil {
		fmt.Println(section, item, "Error:")
		fmt.Printf("Severity: %v : %v\n\n", codes.Glossary[code].TailwindSeverity, new_msg)
	} else {
		fmt.Errorf(err.Error())
	}

}
