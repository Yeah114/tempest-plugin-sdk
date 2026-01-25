package api

type ScoreUpdateEvent struct {
	Scores      map[string]map[string]int `json:"scores"`
	Players     map[string]string         `json:"players"`
	Scoreboards map[string]string         `json:"scoreboards"`
}

type PlayerScoreQueryResult struct {
	ScoreboardName string `json:"ScoreboardName"`
	DisplayName    string `json:"DisplayName"`
	Score          int    `json:"Score"`
}

type RankQueryResult struct {
	PlayerUUID string `json:"PlayerUID"`
	PlayerName string `json:"PlayerName"`
	Score      int    `json:"Score"`
}
type ScoreboardDaemon interface {
	Name() (name string)
	ReConfig(config map[string]interface{}) (err error)
	Config() (config map[string]interface{})

	RegisterWhenScoreUpdate(handler func(event *ScoreUpdateEvent)) (string, error)
	UnregisterWhenScoreUpdate(listenerID string) bool
	QueryScoreByPlayerUUID(uuid string) *[]PlayerScoreQueryResult
	QueryRankByScoreboard(scoreboardName string, descending bool, maxCount int) *[]RankQueryResult
}
