package reportmanager

type Score struct {
	Category string  `json:"category"`
	Value    float32 `json:"value"`
}

type Headers struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type ReportDef struct {
	Method   string         `json:"method"`
	Path     string         `json:"path"`
	Message  string         `json:"message"`
	Headers  []Headers      `json:"headers"`
	Metadata map[string]any `json:"metadata"`
}

type Report struct {
	Score   Score       `json:"score"`
	Reports []ReportDef `json:"rules"`
}

type ReportManager map[string]Report

func New() ReportManager {
	return make(ReportManager)
}

func (r ReportManager) PushReport(ruleName string, data ReportDef) {
	if val, ok := r[ruleName]; ok {
		val.Reports = append(val.Reports, data)
	} else {
		r[ruleName] = Report{Reports: []ReportDef{data}}
	}
}

func (r ReportManager) SetScore(ruleName string, score Score) {
	if val, ok := r[ruleName]; ok {
		val.Score = score
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
