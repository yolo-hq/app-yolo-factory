package constants

// SuggestionCategory identifies the category of a suggestion.
type SuggestionCategory string

const (
	CategoryOptimization      SuggestionCategory = "optimization"
	CategoryRefactoring       SuggestionCategory = "refactoring"
	CategoryTechDebt          SuggestionCategory = "tech_debt"
	CategorySecurity          SuggestionCategory = "security"
	CategoryNewFeature        SuggestionCategory = "new_feature"
	CategoryPatternExtraction SuggestionCategory = "pattern_extraction"
	CategoryBugFix            SuggestionCategory = "bug_fix"
)
