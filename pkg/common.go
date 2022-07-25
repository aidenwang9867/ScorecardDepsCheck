package pkg

import (
	"fmt"
	"strings"

	"github.com/ossf/scorecard/v4/checker"
	"github.com/ossf/scorecard/v4/log"
)

func textToMarkdown(s string) string {
	return strings.ReplaceAll(s, "\n", "\n\n")
}

// DetailToString turns a detail information into a string.
func DetailToString(d *checker.CheckDetail, logLevel log.Level) string {
	if d.Type == checker.DetailDebug && logLevel != log.DebugLevel {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s: %s", typeToString(d.Type), d.Msg.Text))

	if d.Msg.Path != "" {
		sb.WriteString(fmt.Sprintf(": %s", d.Msg.Path))
		if d.Msg.Offset != 0 {
			sb.WriteString(fmt.Sprintf(":%d", d.Msg.Offset))
		}
		if d.Msg.EndOffset != 0 && d.Msg.Offset < d.Msg.EndOffset {
			sb.WriteString(fmt.Sprintf("-%d", d.Msg.EndOffset))
		}
	}

	if d.Msg.Remediation != nil {
		sb.WriteString(fmt.Sprintf(": %s", d.Msg.Remediation.HelpText))
	}

	return sb.String()
}

func detailsToString(details []checker.CheckDetail, logLevel log.Level) (string, bool) {
	// UPGRADEv2: change to make([]string, len(details)).
	var sa []string
	for i := range details {
		v := details[i]
		s := DetailToString(&v, logLevel)
		if s != "" {
			sa = append(sa, s)
		}
	}
	return strings.Join(sa, "\n"), len(sa) > 0
}

func typeToString(cd checker.DetailType) string {
	switch cd {
	default:
		panic("invalid detail")
	case checker.DetailInfo:
		return "Info"
	case checker.DetailWarn:
		return "Warn"
	case checker.DetailDebug:
		return "Debug"
	}
}
