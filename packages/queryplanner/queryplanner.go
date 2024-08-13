package queryplanner

type QueryNode struct {
	ID       int    `json:"id" jsonschema_description:"unique query id - incrementing integer"`
	Question string `json:"question" jsonschema:"required" jsonschema_description:"Question we are asking using a question answer system, if we are asking multiple questions, this question is asked by also providing the answers to the sub questions"`
	// NodeType string      `json:"node_type" jsonschema:"required,enum=single_question,enum=merge_responses" jsonschema_description:"type of question. Either a single question or a multi question merge when there are multiple questions."`
	DependencyIDs []int `json:"dependency_ids,omitempty" jsonschema_description:"list of sub-question ID's that need to be answered before this question can be answered. Use a subquery when anything may be unknown, and we need to ask multiple questions to get the answer. Dependencies must only be other query IDs."`
}

type QueryPlan struct {
	Nodes []QueryNode `json:"nodes" jsonschema_description:"list of questions to ask"`
}
