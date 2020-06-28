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
	"encoding/json"
	"bufio"
	"bytes"
	"github.com/derailed/popeye/internal/report"
)

type tally struct {
	OK    int `yaml:"ok"`
	Info  int `yaml:"info"`
	Warn  int `yaml:"warning"`
	Error int `yaml:"error"`
	Score int `yaml:"score"`
}


type Error struct {
	Type       string `json:"type"`
	Category   string `json:"category"`
	Severity   int    `json:"priority"`
	Msg        string `json:"message"`
}


type PopeyePlugin struct {
	Name string
	Pop *pkg.Popeye
}

func NewPlugin(cluster *string) *PopeyePlugin {
	fmt.Println(*cluster)
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
	flags := config.NewFlags()


	f := true
	flags.AllNamespaces = &f
	if cluster != nil {
		//Setting the context instead of clustername.
		//TODO: Remove following commented line in future.
		//flags.ClusterName = cluster
		flags.Context = cluster
	}
	pop, err := pkg.NewPopeye(flags, &log.Logger)
	if err != nil {
		return nil
	}

	plug := PopeyePlugin{}
	plug.Pop = pop
	return &plug
}

func (p *PopeyePlugin) GetHappyCLIVisibilityReport() {
	if err := p.Pop.Init(); err != nil {
		return
	}
	if err := p.Pop.Sanitize(); err != nil {
		return
	}

	p.Pop.Dump(true)
}

func (p *PopeyePlugin) GettExpeditionCLIAndUIVisibilityReport() string {
	if err := p.Pop.Init(); err != nil {
		return ""
	}
	if err := p.Pop.Sanitize(); err != nil {
		return ""
	}

	var (
		buf bytes.Buffer
		w = bufio.NewWriter(&buf)
		s = report.NewSanitizer(w, false)
	)

	/*
	p.Pop.Builder.PrintHeader(s)
	mx, err := p.Pop.Client.ClusterHasMetrics()
	if err != nil {
		mx = false
	}*/

	//p.Pop.Builder.PrintClusterInfo(s, p.Pop.Client.ActiveCluster(), mx)
	p.Pop.Builder.PrintReport(config.Level(p.Pop.Config.LinterLevel()), s)
	//p.Pop.Builder.PrintSummary(s)
	returnStr := buf.String()
	return returnStr
}

func (p *PopeyePlugin) GetExpeditionCLIAndUIVisibilityErrors() ([]*Error, int, int) {
	var totalSanitizeCnt int
	var totalErrorCnt int
	result := make([]*Error, 0)
	if err := p.Pop.Init(); err != nil {
		return nil, 0, 0
	}
	if err := p.Pop.Sanitize(); err != nil {
		return nil, 0, 0
	}
	p.Pop.Builder.ToJSON()
	codes, _ := issues.LoadCodes()
	rows := make(map[int][]*Error)
	for _,s := range p.Pop.Builder.Report.Sections {
		output, _ := s.Tally.MarshalJSON()
		var t tally
		json.Unmarshal(output, &t)
		totalSanitizeCnt += (t.OK + t.Info + t.Warn + t.Error)
		totalErrorCnt += t.Error
		for k,v := range s.Outcome {
			for _,issue := range v {
				if(issue.Level == 3) {
					row := PopeyToAction(s.Title, k, issue.Message, codes)
					if len(row) == 4 {
						//w.Append(row)
						code,_ := strconv.Atoi(row[2])
						errList := rows[code]
						errList = append(errList, getNewError(s.Title, k, issue.Message, codes))
						rows[code] = errList
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
		errList := rows[k]
		for _, err := range errList {
			result = append(result, err)
		}
	}
	return result, totalSanitizeCnt, totalErrorCnt
}

func (p *PopeyePlugin) GetHappyCLIVisibilityErrors(w *tablewriter.Table) {
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

func getNewError(section string, item string, msg string, codes *issues.Codes) *Error {
	code, new_msg, err := PopeyCodeFromMsg(msg)
	if err == nil {
		e := Error{section, item, int(codes.Glossary[code].TailwindSeverity), new_msg}
		return &e
	}
	return &Error{}
}
