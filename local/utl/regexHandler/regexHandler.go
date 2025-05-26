package regexHandler

import (
	"acc-server-manager/local/model"
	"regexp"
)

type AccServerInstance struct {
    Model     *model.Server
    State     *model.ServerState
}

type RegexHandler struct {
	regex *regexp.Regexp
}

func (rh *RegexHandler) Contains(line string, callback func(...string)) {
	match := rh.regex.FindStringSubmatch(line)
	callback(match...)
}

func New(str string) *RegexHandler {
	return &RegexHandler{
		regex: regexp.MustCompile(str),
	}
}