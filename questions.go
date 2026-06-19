package main

// Choice is one option in a multiple-choice question.
type Choice struct {
	Label   string
	Correct bool
}

// Question is a single multiple-choice question.
type Question struct {
	Prompt      string
	Choices     []Choice
	Explanation string
}

// correctIndex returns the index of the correct choice, or -1 if none.
func (q Question) correctIndex() int {
	for i, c := range q.Choices {
		if c.Correct {
			return i
		}
	}
	return -1
}

// questions is the static seed pool. The generator (web search + LLM) comes later;
// for now we train against a hand-written set.
var questions = []Question{
	{
		Prompt: "A service needs to stay available even if a whole datacenter goes down. Which approach fits best?",
		Choices: []Choice{
			{Label: "Run a single large instance with a fast restart script"},
			{Label: "Deploy redundant instances across multiple availability zones", Correct: true},
			{Label: "Vertically scale the database to handle all load"},
			{Label: "Cache everything in the application's memory"},
		},
		Explanation: "Surviving a datacenter/AZ failure requires redundancy spread across independent failure domains, not a bigger single box.",
	},
	{
		Prompt: "You need to add a read-heavy feature without overloading the primary database. First thing to reach for?",
		Choices: []Choice{
			{Label: "Add read replicas and route reads to them", Correct: true},
			{Label: "Increase the connection pool size on the primary"},
			{Label: "Switch the primary to a faster disk"},
			{Label: "Rewrite all queries as stored procedures"},
		},
		Explanation: "Read replicas offload read traffic from the primary, which is the standard first move for read-heavy scaling.",
	},
	{
		Prompt: "Two services must communicate, but the consumer can tolerate processing messages a few seconds late. What decouples them best?",
		Choices: []Choice{
			{Label: "A synchronous REST call with retries"},
			{Label: "A shared database table both poll"},
			{Label: "A message queue between producer and consumer", Correct: true},
			{Label: "A direct gRPC stream"},
		},
		Explanation: "A message queue decouples producer and consumer in time and load, which suits tolerable-latency async work.",
	},
}
