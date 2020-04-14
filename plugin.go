package popeye

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/derailed/popeye/pkg"
	"github.com/derailed/popeye/pkg/config"
	"github.com/derailed/popeye/internal/issues"
	"fmt"
	"strings"
	"strconv"
	"github.com/olekukonko/tablewriter"
    "sort"
)

type PopeyePlugin struct {
	Name string
	Pop *pkg.Popeye
}

func NewPlugin(cluster *string) *PopeyePlugin {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	flags := config.NewFlags()


	f := true
	flags.AllNamespaces = &f
	if cluster != nil {
		flags.ClusterName = cluster
	}
	pop, err := pkg.NewPopeye(flags, &log.Logger)
	if err != nil {
		return nil
	}
	
	plug := PopeyePlugin{}
	plug.Pop = pop
	return &plug
}

func (p *PopeyePlugin) PrintErrors(w *tablewriter.Table) {
	if err := p.Pop.Init(); err != nil {
		return
	}
	if err := p.Pop.Sanitize(); err != nil {
		return
	}
	p.Pop.Builder.ToJSON()
    codes, _ := issues.LoadCodes()
    w.SetHeader([]string{"Section", "Item", "Priority", "Message"})
    rows := make(map[int][][]string)
    for _,s := range p.Pop.Builder.Report.Sections {

        for k,v := range s.Outcome {
            for _,issue := range v {
                if(issue.Level == 3) {
                    row := PopeyToAction(s.Title, k, issue.Message, codes)
                    if len(row) == 4 {
                        //w.Append(row)
                        code,_ := strconv.Atoi(row[2])
                        rows[code] = append(rows[code],row)
                    }
                }
            }
        }
    }
    keys := make([]int, 0)
    for k, _ := range rows {
        keys = append(keys, k)
    }
    sort.Ints(keys)
    for _, k := range keys {
        w.AppendBulk(rows[k])
    }
    w.Render()

}


func PopeyCodeFromMsg(msg string) (config.ID,string,error) {
        if msg[0] != '[' {
                return 0,msg,fmt.Errorf("No Code Found")
        }
        stop := strings.IndexByte(msg, ']')
        imsg,err := strconv.Atoi(msg[5:stop])
        if err != nil {
                return 0,msg,err
        }
        return config.ID(imsg),msg[stop+2:],nil
}
func PopeyToAction(section string, item string, msg string, codes *issues.Codes) []string{
    
    code, new_msg, err := PopeyCodeFromMsg(msg)
    if err == nil {
    		// fmt.Fprintln(w, section, item, "Error:")
            // fmt.Fprintf(w, "Severity: %v : %v\n\n", codes.Glossary[code].TailwindSeverity, new_msg)
            return []string{section, item, strconv.Itoa(int(codes.Glossary[code].TailwindSeverity)), new_msg}
    } else {
            fmt.Errorf(err.Error())
    }
    return []string{}
}
