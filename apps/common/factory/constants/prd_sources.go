package constants

// PRDSource identifies how a PRD was created.
type PRDSource string

const (
	SourceManual           PRDSource = "manual"
	SourceGrillMe          PRDSource = "grill_me"
	SourceFactoryGenerated PRDSource = "factory_generated"
	SourceImported         PRDSource = "imported"
)
