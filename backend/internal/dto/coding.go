package dto

type CodeRequest struct {
	Language string `json:"language" binding:"required,oneof=cpp java python"`
	Code     string `json:"code" binding:"required"`
}

type CodingLanguage struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type CodingStartResponse struct {
	SessionID uint `json:"sessionId"`
}

type CodingProblemSummary struct {
	Order      int    `json:"order"`
	Slug       string `json:"slug"`
	Title      string `json:"title"`
	Difficulty string `json:"difficulty"`
	Status     string `json:"status"`
	Failures   int    `json:"failures"`
	Points     int    `json:"points"`
	RuntimeMs  int    `json:"runtimeMs"`
	Language   string `json:"language"`
	TimeOpt    string `json:"timeOptimal"`
}

type CodingSessionState struct {
	SessionID       uint                   `json:"sessionId"`
	AssessmentTitle string                 `json:"assessmentTitle"`
	Status          string                 `json:"status"`
	RemainingSec    int                    `json:"remainingSec"`
	Score           int                    `json:"score"`
	MaxScore        int                    `json:"maxScore"`
	Current         int                    `json:"current"` // order of the next open problem, 0 when finished
	Problems        []CodingProblemSummary `json:"problems"`
}

type CodingExample struct {
	Input  string `json:"input"`
	Output string `json:"output"`
}

type CodingProblemView struct {
	Order       int               `json:"order"`
	Slug        string            `json:"slug"`
	Title       string            `json:"title"`
	Difficulty  string            `json:"difficulty"`
	Description string            `json:"description"`
	Examples    []CodingExample   `json:"examples"`
	Starter     map[string]string `json:"starter"`
	Hints       []string          `json:"hints"` // only the hints already unlocked
	Failures    int               `json:"failures"`
	MaxFailures int               `json:"maxFailures"`
	Status      string            `json:"status"`
}

type CodingCase struct {
	Tag      string `json:"tag"`
	Visible  bool   `json:"visible"`
	Passed   bool   `json:"passed"`
	Status   string `json:"status"`
	Input    string `json:"input,omitempty"`
	Expected string `json:"expected,omitempty"`
	Got      string `json:"got,omitempty"`
	TimeMs   int    `json:"timeMs"`
}

type CodingFeedback struct {
	TimeOptimal  string   `json:"timeOptimal"`
	SpaceOptimal string   `json:"spaceOptimal"`
	Approach     string   `json:"approach"`
	Notes        []string `json:"notes"`
	RuntimeMs    int      `json:"runtimeMs"`
}

type CodingResult struct {
	Mode          string          `json:"mode"` // run or submit
	Verdict       string          `json:"verdict"`
	Passed        int             `json:"passed"`
	Total         int             `json:"total"`
	CompileOutput string          `json:"compileOutput,omitempty"`
	Detail        string          `json:"detail,omitempty"`
	Cases         []CodingCase    `json:"cases"`
	Analysis      []string        `json:"analysis,omitempty"`
	Hint          string          `json:"hint,omitempty"`
	HintNumber    int             `json:"hintNumber,omitempty"`
	Failures      int             `json:"failures"`
	Solved        bool            `json:"solved"`
	Exhausted     bool            `json:"exhausted"` // 5 hints used: the session moves on
	Points        int             `json:"points"`
	NextOrder     int             `json:"nextOrder"` // 0 when there is no further problem
	SessionDone   bool            `json:"sessionDone"`
	Feedback      *CodingFeedback `json:"feedback,omitempty"`
}