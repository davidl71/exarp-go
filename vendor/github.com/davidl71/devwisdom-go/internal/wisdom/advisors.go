package wisdom

import "fmt"

// AdvisorRegistry manages advisor mappings for metrics, tools, and stages
type AdvisorRegistry struct {
	metricAdvisors map[string]*AdvisorInfo
	toolAdvisors   map[string]*AdvisorInfo
	stageAdvisors  map[string]*AdvisorInfo
	initialized    bool
}

// NewAdvisorRegistry creates a new advisor registry
func NewAdvisorRegistry() *AdvisorRegistry {
	return &AdvisorRegistry{
		metricAdvisors: make(map[string]*AdvisorInfo),
		toolAdvisors:   make(map[string]*AdvisorInfo),
		stageAdvisors:  make(map[string]*AdvisorInfo),
	}
}

// Initialize loads advisor mappings
func (r *AdvisorRegistry) Initialize() {
	if r.initialized {
		return
	}

	// Load metric advisors
	r.loadMetricAdvisors()

	// Load tool advisors
	r.loadToolAdvisors()

	// Load stage advisors
	r.loadStageAdvisors()

	r.initialized = true
}

// loadMetricAdvisors populates metric → advisor mappings
func (r *AdvisorRegistry) loadMetricAdvisors() {
	r.metricAdvisors["security"] = &AdvisorInfo{
		Advisor:   "bofh",
		Icon:      "😈",
		Rationale: "BOFH is paranoid about security, expects users to break everything",
		HelpsWith: "Finding vulnerabilities, defensive thinking, access control",
	}

	r.metricAdvisors["testing"] = &AdvisorInfo{
		Advisor:   "stoic",
		Icon:      "🏛️",
		Rationale: "Stoics teach discipline through adversity - tests reveal truth",
		HelpsWith: "Persistence through failures, accepting harsh feedback",
	}

	r.metricAdvisors["documentation"] = &AdvisorInfo{
		Advisor:   "confucius",
		Icon:      "🎓",
		Rationale: "Confucius emphasized teaching and transmitting wisdom",
		HelpsWith: "Clear explanations, teaching future maintainers",
	}

	r.metricAdvisors["completion"] = &AdvisorInfo{
		Advisor:   "art_of_war",
		Icon:      "⚔️",
		Rationale: "Sun Tzu teaches strategy and decisive execution",
		HelpsWith: "Prioritization, knowing when to attack vs wait",
	}

	r.metricAdvisors["alignment"] = &AdvisorInfo{
		Advisor:   "tao",
		Icon:      "☯️",
		Rationale: "Tao emphasizes balance, flow, and purpose",
		HelpsWith: "Ensuring work serves project goals, finding harmony",
	}

	r.metricAdvisors["clarity"] = &AdvisorInfo{
		Advisor:   "gracian",
		Icon:      "🎭",
		Rationale: "Gracián's maxims are models of clarity and pragmatism",
		HelpsWith: "Simplifying complexity, clear communication",
	}

	r.metricAdvisors["ci_cd"] = &AdvisorInfo{
		Advisor:   "kybalion",
		Icon:      "⚗️",
		Rationale: "Kybalion teaches cause and effect - CI/CD is pure causation",
		HelpsWith: "Understanding pipelines, automation philosophy",
	}

	r.metricAdvisors["dogfooding"] = &AdvisorInfo{
		Advisor:   "murphy",
		Icon:      "🔧",
		Rationale: "Murphy's Law: if it can break, it will - use your own tools!",
		HelpsWith: "Finding edge cases, eating your own cooking",
	}

	r.metricAdvisors["uniqueness"] = &AdvisorInfo{
		Advisor:   "shakespeare",
		Icon:      "🎭",
		Rationale: "Shakespeare created unique works that transcended his time",
		HelpsWith: "Creative differentiation, memorable design",
	}

	r.metricAdvisors["codebase"] = &AdvisorInfo{
		Advisor:   "enochian",
		Icon:      "🔮",
		Rationale: "Enochian mysticism reveals hidden structure and patterns",
		HelpsWith: "Architecture, finding hidden connections",
	}

	r.metricAdvisors["parallelizable"] = &AdvisorInfo{
		Advisor:   "tao_of_programming",
		Icon:      "💻",
		Rationale: "The Tao of Programming teaches elegant parallel design",
		HelpsWith: "Decomposition, independent task design",
	}

	// Hebrew Advisors - Jewish wisdom traditions
	r.metricAdvisors["ethics"] = &AdvisorInfo{
		Advisor:   "rebbe",
		Icon:      "🕎",
		Rationale: "The Rebbe teaches ethical conduct and righteous behavior (מוסר)",
		HelpsWith: "Code ethics, proper conduct, doing the right thing",
		Language:  "hebrew",
	}

	r.metricAdvisors["perseverance"] = &AdvisorInfo{
		Advisor:   "tzaddik",
		Icon:      "✡️",
		Rationale: "The Tzaddik (righteous one) demonstrates steadfast commitment",
		HelpsWith: "Persistence, staying on the righteous path, not giving up",
		Language:  "hebrew",
	}

	r.metricAdvisors["wisdom"] = &AdvisorInfo{
		Advisor:   "chacham",
		Icon:      "📜",
		Rationale: "The Chacham (sage) seeks deep understanding through Torah",
		HelpsWith: "Deep analysis, seeking understanding, learning from tradition",
		Language:  "hebrew",
	}
}

// loadToolAdvisors populates tool → advisor mappings
func (r *AdvisorRegistry) loadToolAdvisors() {
	r.toolAdvisors["project_scorecard"] = &AdvisorInfo{
		Advisor:   "pistis_sophia",
		Rationale: "Journey through aeons mirrors project health stages",
	}

	r.toolAdvisors["project_overview"] = &AdvisorInfo{
		Advisor:   "kybalion",
		Rationale: "Hermetic principles for holistic understanding",
	}

	r.toolAdvisors["sprint_automation"] = &AdvisorInfo{
		Advisor:   "art_of_war",
		Rationale: "Sprint is a campaign requiring strategy",
	}

	r.toolAdvisors["check_documentation_health"] = &AdvisorInfo{
		Advisor:   "confucius",
		Rationale: "Teaching requires good documentation",
	}

	r.toolAdvisors["analyze_todo2_alignment"] = &AdvisorInfo{
		Advisor:   "tao",
		Rationale: "Alignment is balance and flow",
	}

	r.toolAdvisors["detect_duplicate_tasks"] = &AdvisorInfo{
		Advisor:   "bofh",
		Rationale: "Duplicates are user error manifested",
	}

	r.toolAdvisors["scan_dependency_security"] = &AdvisorInfo{
		Advisor:   "bofh",
		Rationale: "Security paranoia is a feature",
	}

	r.toolAdvisors["run_tests"] = &AdvisorInfo{
		Advisor:   "stoic",
		Rationale: "Tests teach through failure",
	}

	r.toolAdvisors["validate_ci_cd_workflow"] = &AdvisorInfo{
		Advisor:   "kybalion",
		Rationale: "CI/CD is cause and effect",
	}

	r.toolAdvisors["dev_reload"] = &AdvisorInfo{
		Advisor:   "murphy",
		Rationale: "Hot reload because Murphy says restarts will fail at the worst time",
	}

	// Hebrew advisor tools - for ethical and wisdom-focused operations
	r.toolAdvisors["ethics_check"] = &AdvisorInfo{
		Advisor:   "rebbe",
		Rationale: "Rebbe guides ethical code review and conduct",
		Language:  "hebrew",
	}

	r.toolAdvisors["wisdom_reflection"] = &AdvisorInfo{
		Advisor:   "chacham",
		Rationale: "Chacham provides deep wisdom for retrospectives",
		Language:  "hebrew",
	}
}

// loadStageAdvisors populates stage → advisor mappings
func (r *AdvisorRegistry) loadStageAdvisors() {
	r.stageAdvisors["daily_checkin"] = &AdvisorInfo{
		Advisor:   "pistis_sophia",
		Icon:      "📜",
		Rationale: "Start each day with enlightenment journey wisdom",
	}

	r.stageAdvisors["planning"] = &AdvisorInfo{
		Advisor:   "art_of_war",
		Icon:      "⚔️",
		Rationale: "Planning is strategy - Sun Tzu is the master",
	}

	r.stageAdvisors["implementation"] = &AdvisorInfo{
		Advisor:   "tao_of_programming",
		Icon:      "💻",
		Rationale: "During coding, let the code flow naturally",
	}

	r.stageAdvisors["debugging"] = &AdvisorInfo{
		Advisor:   "bofh",
		Icon:      "😈",
		Rationale: "BOFH knows all the ways things break",
	}

	r.stageAdvisors["review"] = &AdvisorInfo{
		Advisor:   "stoic",
		Icon:      "🏛️",
		Rationale: "Review requires accepting harsh truths with equanimity",
	}

	r.stageAdvisors["retrospective"] = &AdvisorInfo{
		Advisor:   "confucius",
		Icon:      "🎓",
		Rationale: "Retrospectives are about learning and teaching",
	}

	r.stageAdvisors["celebration"] = &AdvisorInfo{
		Advisor:   "shakespeare",
		Icon:      "🎭",
		Rationale: "Celebrate with drama and poetry!",
	}

	// Hebrew advisor stages
	r.stageAdvisors["shabbat"] = &AdvisorInfo{
		Advisor:   "rebbe",
		Icon:      "🕎",
		Rationale: "Shabbat is for reflection and spiritual renewal (מנוחה)",
		Language:  "hebrew",
	}

	r.stageAdvisors["teshuvah"] = &AdvisorInfo{
		Advisor:   "tzaddik",
		Icon:      "✡️",
		Rationale: "Teshuvah (repentance) is for fixing past mistakes and returning to the right path",
		Language:  "hebrew",
	}

	r.stageAdvisors["learning"] = &AdvisorInfo{
		Advisor:   "chacham",
		Icon:      "📜",
		Rationale: "Torah study and continuous learning (לימוד)",
		Language:  "hebrew",
	}
}

// GetAdvisorForMetric returns the advisor for a given metric
func (r *AdvisorRegistry) GetAdvisorForMetric(metric string) (*AdvisorInfo, error) {
	advisor, exists := r.metricAdvisors[metric]
	if !exists {
		availableMetrics := make([]string, 0, len(r.metricAdvisors))
		for m := range r.metricAdvisors {
			availableMetrics = append(availableMetrics, m)
		}
		return nil, fmt.Errorf("no advisor found for metric %q (available metrics: %v). Check metric name spelling", metric, availableMetrics)
	}
	return advisor, nil
}

// GetAdvisorForTool returns the advisor for a given tool
func (r *AdvisorRegistry) GetAdvisorForTool(tool string) (*AdvisorInfo, error) {
	advisor, exists := r.toolAdvisors[tool]
	if !exists {
		availableTools := make([]string, 0, len(r.toolAdvisors))
		for t := range r.toolAdvisors {
			availableTools = append(availableTools, t)
		}
		return nil, fmt.Errorf("no advisor found for tool %q (available tools: %v). Check tool name spelling", tool, availableTools)
	}
	return advisor, nil
}

// GetAdvisorForStage returns the advisor for a given stage
func (r *AdvisorRegistry) GetAdvisorForStage(stage string) (*AdvisorInfo, error) {
	advisor, exists := r.stageAdvisors[stage]
	if !exists {
		availableStages := make([]string, 0, len(r.stageAdvisors))
		for s := range r.stageAdvisors {
			availableStages = append(availableStages, s)
		}
		return nil, fmt.Errorf("no advisor found for stage %q (available stages: %v). Check stage name spelling", stage, availableStages)
	}
	return advisor, nil
}

// GetConsultationMode returns the consultation mode based on score
func GetConsultationMode(score float64) ConsultationModeConfig {
	modes := []ConsultationModeConfig{
		{
			Name:        string(ModeChaos),
			MinScore:    0,
			MaxScore:    30,
			Frequency:   "every_action",
			Description: "Chaos mode: Consult advisor before every significant action",
			Icon:        "🔥",
		},
		{
			Name:        string(ModeBuilding),
			MinScore:    30,
			MaxScore:    60,
			Frequency:   "start_and_review",
			Description: "Building mode: Consult at start of work and during review",
			Icon:        "🏗️",
		},
		{
			Name:        string(ModeMaturing),
			MinScore:    60,
			MaxScore:    80,
			Frequency:   "milestones",
			Description: "Maturing mode: Consult at planning and major milestones",
			Icon:        "🌱",
		},
		{
			Name:        string(ModeMastery),
			MinScore:    80,
			MaxScore:    100,
			Frequency:   "weekly",
			Description: "Mastery mode: Weekly reflection with advisor",
			Icon:        "🎯",
		},
	}

	for _, mode := range modes {
		if mode.MinScore <= score && score < mode.MaxScore {
			return mode
		}
	}

	// Handle edge cases
	if score < 0 {
		// Negative scores default to chaos
		return modes[0] // ModeChaos
	}

	// Default to mastery for scores >= 100
	return modes[3] // ModeMastery
}

// GetModeConfig returns configuration for a session mode
func GetModeConfig(mode SessionMode) *ModeConfig {
	configs := map[SessionMode]*ModeConfig{
		SessionModeAgent: {
			PreferredAdvisors: []string{"art_of_war", "tao_of_programming", "kybalion"},
			Tone:              "strategic",
			Focus:             "progress tracking and checkpoints",
		},
		SessionModeAsk: {
			PreferredAdvisors: []string{"confucius", "gracian", "stoic"},
			Tone:              "direct",
			Focus:             "quick answers and focused explanations",
		},
		SessionModeManual: {
			PreferredAdvisors: []string{"tao_of_programming", "bible", "pistis_sophia"},
			Tone:              "observational",
			Focus:             "encouragement and reflection",
		},
	}

	if config, exists := configs[mode]; exists {
		return config
	}

	// Default config for unknown modes
	return nil
}

// AdjustAdvisorForMode adjusts advisor selection based on session mode for random consultations
// Returns the adjusted advisor ID and rationale if adjustment is made, otherwise returns empty strings
func AdjustAdvisorForMode(sessionMode SessionMode, consultationType string, availableSources []string) (string, string) {
	// Only adjust for random consultations
	if consultationType != "random" {
		return "", ""
	}

	modeConfig := GetModeConfig(sessionMode)
	if modeConfig == nil {
		return "", ""
	}

	// Find preferred advisors that are available
	availablePreferred := []string{}
	for _, preferred := range modeConfig.PreferredAdvisors {
		for _, available := range availableSources {
			if preferred == available {
				availablePreferred = append(availablePreferred, preferred)
				break
			}
		}
	}

	if len(availablePreferred) > 0 {
		// Use first available preferred advisor (could be randomized later)
		selected := availablePreferred[0]
		rationale := fmt.Sprintf("Mode-aware selection for %s", sessionMode)
		return selected, rationale
	}

	return "", ""
}

// GetAllMetricAdvisors returns all metric → advisor mappings
func (r *AdvisorRegistry) GetAllMetricAdvisors() map[string]*AdvisorInfo {
	if !r.initialized {
		r.Initialize()
	}
	result := make(map[string]*AdvisorInfo)
	for k, v := range r.metricAdvisors {
		result[k] = v
	}
	return result
}

// GetAllToolAdvisors returns all tool → advisor mappings
func (r *AdvisorRegistry) GetAllToolAdvisors() map[string]*AdvisorInfo {
	if !r.initialized {
		r.Initialize()
	}
	result := make(map[string]*AdvisorInfo)
	for k, v := range r.toolAdvisors {
		result[k] = v
	}
	return result
}

// GetAllStageAdvisors returns all stage → advisor mappings
func (r *AdvisorRegistry) GetAllStageAdvisors() map[string]*AdvisorInfo {
	if !r.initialized {
		r.Initialize()
	}
	result := make(map[string]*AdvisorInfo)
	for k, v := range r.stageAdvisors {
		result[k] = v
	}
	return result
}
