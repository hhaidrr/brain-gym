# Brain Gym — Architecture

A terminal-based "mental gym" that trains a user's system-design judgment by
quizzing them on generated questions, tracking what they get wrong, and
resurfacing weak areas over time.

See `thing.md` for the product motivation.

## v1 scope

- **Interface:** terminal TUI (Bubble Tea), single binary, local SQLite.
- **Question type:** multiple-choice only for now. The schema and interfaces
  support a mixed pool (`multiple_choice`, `open_ended`) so an LLM-judged
  open-ended pool can be added later without migration.
- **Generation:** Claude + server-side web search produces grounded MC
  questions on a topic.
- **Voting:** user marks a question keep/discard; quality score drives pruning
  and scheduling weight.
- **Scheduling:** spaced repetition (SM-2-ish) weighted by topic coverage and
  question quality.

## Components

```
            ┌────────────────────────────────────────────┐
            │              TUI (Bubble Tea)               │
            │   session loop, question view, voting       │
            └───────────────┬────────────────────────────┘
                            │ calls services
        ┌───────────────────┼───────────────────────────┐
        │                   │                           │
 ┌──────▼──────┐    ┌───────▼────────┐         ┌────────▼────────┐
 │  Generator  │    │    Trainer     │         │     Grader      │
 │ web search →│    │  selection /   │         │  MC: key match  │
 │ MC question │    │  session +     │         │  (open_ended:   │
 │ + options   │    │  review update │         │   LLM, later)   │
 └──────┬──────┘    └───────┬────────┘         └────────┬────────┘
        │                   │                           │
        └─────────┬─────────┴───────────┬───────────────┘
                  │                      │
          ┌───────▼────────┐    ┌────────▼─────────┐
          │  Store (repo)  │    │   LLM client     │
          │   SQLite       │    │  Claude + web    │
          └────────────────┘    │  search tool     │
                                └──────────────────┘
```

`generator`, `trainer`, and `grader` depend only on `domain` plus the `store`
and `llm` interfaces — never on each other's internals or on the TUI. This
keeps storage and the LLM swappable and lets every service be tested with fakes.

## Package layout

```
brain-gym/
  cmd/braingym/main.go        # config, wiring, flags
  internal/
    domain/                   # Question, Choice, Attempt, Review, Topic — pure types
    store/                    # SQLite repo (interface + impl)
    llm/                      # Claude client wrapper (generate; grade later)
    generator/                # web search → synthesize MC + validate → persist
    grader/                   # answer evaluation; MC = deterministic key match
    trainer/                  # scheduling (SM-2) + session orchestration
    tui/                      # Bubble Tea models/views
  docs/
```

## Data model (SQLite)

- **questions** — `id, topic, difficulty, type ('multiple_choice'|'open_ended'),
  prompt, explanation, source_refs (JSON), status ('active'|'discarded'),
  quality_score, created_at`
- **choices** — `id, question_id, label, is_correct` (MC only; `open_ended`
  questions carry a rubric in a future column instead)
- **attempts** — `id, question_id, selected_choice_id, correct (bool), created_at`
- **reviews** — `question_id, ease, interval_days, due_at` (spaced-repetition state)
- **votes** — folded into `questions` via `quality_score` + counts; auto-discard
  past a downvote threshold
- **topics** — coverage tracking so the generator fills gaps

## Key flows

### Generate
`topic → Claude (web search tool) → draft MC question + options + correct key +
explanation → self-check pass (is this a real fixed-answer scenario?, exactly one
correct option?) → persist as active`. Voting prunes what the self-check misses.

### Train (session loop)
`trainer.NextQuestion()` picks a due/weak question → TUI renders prompt +
shuffled options → user selects → `grader.Grade()` (key match) → record attempt
→ `trainer` updates review state (SM-2) → user votes keep/discard.

### Grade
MC grading is deterministic key matching in `grader`, no LLM. The `Grader`
interface returns a structured result (correct, correct choice, explanation) so
an `open_ended` LLM-judge implementation can slot in behind the same interface.

## Technical choices

- **LLM:** Claude via the official Go SDK with the server-side **web search
  tool** for grounded generation; `claude-opus-4-8` for generation quality.
  (Confirm exact SDK + model details against the claude-api reference before
  writing LLM code.)
- **Store:** `modernc.org/sqlite` (pure Go, no cgo). Single-user, local,
  zero-ops. Repo interface keeps Postgres a drop-in if it ever goes multi-user.
- **Scheduling:** SM-2 weighted by topic coverage and question quality. Simple,
  well-understood, no ML.

## Deferred (post-v1)

- `open_ended` pool with LLM-as-judge rubric grading.
- HTTP API behind the same services for multi-device / multi-user.
- Batch/background generation to keep the pool ahead of the user.
