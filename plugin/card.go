package plugin

type MessageCard struct {
	Type            string
	Context         string
	ThemeColor      string
	Summary         string
	Sections        []MessageCardSection
	PotentialAction []GenericAction
}

type MessageCardSection struct {
	ActivityImage    string
	ActivityTitle    string
	ActivitySubtitle string
	ActivityText     string
	Facts            []MessageCardSectionFact
	Markdown         bool
}

type MessageCardSectionFact struct {
	Name  string
	Value string
}

type GenericAction struct {
	Type string
	Name string
	// For OpenURI
	Targets *[]OpenURITarget `json:"targets,omitempty"`
	// For HttpPOST
	Target          *string     `json:"target,omitempty"`
	Headers         *[]KeyValue `json:"headers,omitempty"`
	Body            *string     `json:"body,omitempty"`
	BodyContentType *string     `json:"bodyContentType,omitempty"`
	// For ActionCard
	Inputs  *[]GenericInput  `json:"inputs,omitempty"`
	Actions *[]GenericAction `json:"actions,omitempty"`
}

type GenericInput struct {
	Type       string
	ID         string
	IsRequired bool
	Title      string
	Value      string
	// For TextInput
	IsMultiline *bool `json:"isMultiline,omitempty"`
	MaxLength   *int  `json:"maxLength,omitempty"`
	// For DateInput
	IncludeTime *bool `json:"includeTime,omitempty"`
	// For MultichoiceInput
	Choices       *[]KeyValue `json:"choices,omitempty"`
	IsMultiSelect *bool       `json:"isMultiSelect,omitempty"`
	Style         *string     `json:"style,omitempty"`
}

type OpenURITarget struct {
	OS  string
	URI string
}

type KeyValue struct {
	Name  string
	Value string
}
