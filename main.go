package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

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

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Brain Gym — system design trainer")
	fmt.Println("Answer with the option number. Type 'q' to quit.")

	score, answered := 0, 0

	for i := 0; ; i = (i + 1) % len(questions) {
		q := questions[i]

		fmt.Printf("\n%s\n", q.Prompt)
		for n, c := range q.Choices {
			fmt.Printf("  %d) %s\n", n+1, c.Label)
		}
		fmt.Print("> ")

		line, err := reader.ReadString('\n')
		if err != nil { // EOF (Ctrl-D) ends the session
			break
		}
		input := strings.TrimSpace(line)

		if input == "q" || input == "quit" {
			break
		}

		choice, ok := parseChoice(input, len(q.Choices))
		if !ok {
			fmt.Printf("Please enter a number 1-%d, or 'q' to quit.\n", len(q.Choices))
			i-- // stay on the same question
			continue
		}

		answered++
		if q.Choices[choice].Correct {
			score++
			fmt.Println("✓ Correct.")
		} else {
			correct := correctLabel(q)
			fmt.Printf("✗ Not quite. Correct answer: %s\n", correct)
		}
		fmt.Printf("  %s\n", q.Explanation)
	}

	fmt.Printf("\nSession over — %d/%d correct. See you at the gym.\n", score, answered)
}

// parseChoice converts a 1-based input string into a 0-based choice index.
func parseChoice(input string, n int) (int, bool) {
	var num int
	if _, err := fmt.Sscanf(input, "%d", &num); err != nil {
		return 0, false
	}
	if num < 1 || num > n {
		return 0, false
	}
	return num - 1, true
}

// correctLabel returns the label of the correct choice for a question.
func correctLabel(q Question) string {
	for _, c := range q.Choices {
		if c.Correct {
			return c.Label
		}
	}
	return ""
}
