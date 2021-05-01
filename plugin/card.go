package plugin

type MessageCard struct {
	Type       string
	Context    string
	ThemeColor string
	Summary    string
	Sections   []MessageCardSection
}

type MessageCardSection struct {
	ActivityImage    string
	ActivityTitle    string
	ActivitySubtitle string
	ActivityText     string
	Facts            []MessageCardSectionFact
	PotentialAction  []OpenURIAction
	Markdown         bool
}

type MessageCardSectionFact struct {
	Name  string
	Value string
}

type OpenURIAction struct {
	Name    string
	Targets []OpenURITarget
}

type OpenURITarget struct {
	OS  string
	URI string
}
