package tracking

import (
	"acc-server-manager/local/model"
	"acc-server-manager/local/utl/regex_handler"
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

type StateChange int

const (
	PlayerCount StateChange = iota
	Session
)

var StateChanges = map[StateChange]string{
	PlayerCount: "player-count",
	Session:     "session",
}

type AccServerInstance struct {
	Model         *model.Server
	State         *model.ServerState
	OnStateChange func(*model.ServerState, ...StateChange)
}

func NewAccServerInstance(server *model.Server, onStateChange func(*model.ServerState, ...StateChange)) *AccServerInstance {
	return &AccServerInstance{
		Model:         server,
		State:         &model.ServerState{PlayerCount: 0},
		OnStateChange: onStateChange,
	}
}

type StateRegexHandler struct {
	*regex_handler.RegexHandler
	test string
}

func NewRegexHandler(str string, test string) *StateRegexHandler {
	return &StateRegexHandler{
		RegexHandler: regex_handler.New(str),
		test:         test,
	}
}

func (rh *StateRegexHandler) Test(line string) bool {
	return strings.Contains(line, rh.test)
}

func (rh *StateRegexHandler) Count(line string) int {
	var count int = 0
	rh.Contains(line, func(strs ...string) {
		if len(strs) == 2 {
			if ct, err := strconv.Atoi(strs[1]); err == nil {
				count = ct
			}
		}
	})
	return count
}

func (rh *StateRegexHandler) Change(line string) (string, string) {
	var old string = ""
	var new string = ""
	rh.Contains(line, func(strs ...string) {
		if len(strs) == 3 {
			old = strs[1]
			new = strs[2]
		}
	})
	return old, new
}

func TailLogFile(path string, callback func(string)) {
	file, _ := os.Open(path)
	defer file.Close()

	file.Seek(0, os.SEEK_END) // Start at end of file
	reader := bufio.NewReader(file)

	for {
		line, err := reader.ReadString('\n')
		if err == nil {
			callback(line)
		} else {
			time.Sleep(500 * time.Millisecond) // wait for new data
		}
	}
}

type LogStateType int

const (
	SessionChange LogStateType = iota
	LeaderboardUpdate
	UDPCount
	ClientsOnline
	RemovingDeadConnection
)

var logStateContain = map[LogStateType]string{
	SessionChange:          "Session changed",
	LeaderboardUpdate:      "Updated leaderboard for",
	UDPCount:               "Udp message count",
	ClientsOnline:          "client(s) online",
	RemovingDeadConnection: "Removing dead connection",
}

var sessionChangeRegex = NewRegexHandler(`Session changed: (\w+) -> (\w+)`, logStateContain[SessionChange])
var leaderboardUpdateRegex = NewRegexHandler(`Updated leaderboard for (\d+) clients`, logStateContain[LeaderboardUpdate])
var udpCountRegex = NewRegexHandler(`Udp message count (\d+) client`, logStateContain[UDPCount])
var clientsOnlineRegex = NewRegexHandler(`(\d+) client\(s\) online`, logStateContain[ClientsOnline])
var removingDeadConnectionsRegex = NewRegexHandler(`Removing dead connection`, logStateContain[RemovingDeadConnection])

var logStateRegex = map[LogStateType]*StateRegexHandler{
	SessionChange:          sessionChangeRegex,
	LeaderboardUpdate:      leaderboardUpdateRegex,
	UDPCount:               udpCountRegex,
	ClientsOnline:          clientsOnlineRegex,
	RemovingDeadConnection: removingDeadConnectionsRegex,
}

func (instance *AccServerInstance) HandleLogLine(line string) {
	for logState, regexHandler := range logStateRegex {
		if regexHandler.Test(line) {
			switch logState {
			case LeaderboardUpdate:
			case UDPCount:
			case ClientsOnline:
				count := regexHandler.Count(line)
				instance.UpdatePlayerCount(count)
			case SessionChange:
				_, new := regexHandler.Change(line)

				trackSession := model.ToTrackSession(new)
				instance.UpdateSessionChange(trackSession)
			case RemovingDeadConnection:
				instance.UpdatePlayerCount(instance.State.PlayerCount - 1)
			}
		}
	}
}

func (instance *AccServerInstance) UpdateState(callback func(state *model.ServerState, changes *[]StateChange)) {
	state := instance.State
	changes := []StateChange{}
	state.Lock()
	defer state.Unlock()
	callback(state, &changes)
	if len(changes) > 0 {
		instance.OnStateChange(state, changes...)
	}
}

func (instance *AccServerInstance) UpdatePlayerCount(count int) {
	if count < 0 {
		return
	}
	instance.UpdateState(func(state *model.ServerState, changes *[]StateChange) {
		if count == state.PlayerCount {
			return
		}
		if count > 0 && state.PlayerCount == 0 {
			state.SessionStart = time.Now()
			*changes = append(*changes, Session)
		} else if count == 0 {
			state.SessionStart = time.Time{}
			*changes = append(*changes, Session)
		}
		state.PlayerCount = count
		*changes = append(*changes, PlayerCount)
	})
}

func (instance *AccServerInstance) UpdateSessionChange(session model.TrackSession) {
	instance.UpdateState(func(state *model.ServerState, changes *[]StateChange) {
		if session == state.Session {
			return
		}
		if state.PlayerCount > 0 {
			state.SessionStart = time.Now()
		} else {
			state.SessionStart = time.Time{}
		}
		state.Session = session
		*changes = append(*changes, Session)
	})
}
