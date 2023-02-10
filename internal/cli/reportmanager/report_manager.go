package reportmanager

type Score struct {
	Category string  `json:"category" toml:"category"`
	Value    float32 `json:"value" toml:"value"`
}

type Headers struct {
	Key   string `json:"key" toml:"key"`
	Value string `json:"value" toml:"value"`
}

type ReportDef struct {
	// whether its a warning or error
	Method   string         `json:"method,omitempty" toml:"method,omitempty"`
	Path     string         `json:"path,omitempty" toml:"path,omitempty"`
	Message  string         `json:"message" toml:"message"`
	Headers  []Headers      `json:"headers,omitempty" toml:"headers,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty" toml:"metadata,omitempty"`
}

type Report struct {
	Score   Score       `json:"score"`
	Reports []ReportDef `json:"reports,omitempty"`
}

type ReportManager map[string]Report

func New() ReportManager {
	return make(ReportManager)
}

func (r ReportManager) PushReport(ruleName string, data ReportDef) {
	if val, ok := r[ruleName]; ok {
		val.Reports = append(val.Reports, data)
		r[ruleName] = val
	} else {
		r[ruleName] = Report{Reports: []ReportDef{data}}
	}
}

func (r ReportManager) SetScore(ruleName string, score Score) {
	if val, ok := r[ruleName]; ok {
		val.Score = score
		r[ruleName] = val
	} else {
		r[ruleName] = Report{Score: score}
	}
}

func (r ReportManager) GetTotalScore() []Score {
	scoreN := make(map[string]int)
	scoreSum := make(map[string]float32)

	for _, report := range r {
		scoreN[report.Score.Category] += 1
		scoreSum[report.Score.Category] += report.Score.Value
	}

	var finalScore []Score
	for cat, score := range scoreSum {
		finalScore = append(finalScore, Score{Category: cat, Value: score / float32(scoreN[cat])})
	}

	return finalScore
}
