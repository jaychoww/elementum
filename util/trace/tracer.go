package trace

import (
	"bytes"
	"fmt"
	"time"
)

type Tracer struct {
	current  time.Time
	create   time.Time
	complete time.Time

	currentStage int
	stages       []string
	timers       []time.Duration
}

func (t *Tracer) String() string {
	if t.complete.IsZero() {
		t.Complete()
	}

	stages := bytes.Buffer{}
	for stage := 0; stage < t.currentStage; stage++ {
		stages.WriteString(fmt.Sprintf("\n%18s: %s", t.stages[stage], t.timers[stage]))
	}

	return fmt.Sprintf(`
           Created: %s%s
          Complete: %s`,
		t.create.Format("2006-01-02 15:04:05"), stages.String(), t.complete.Sub(t.create))
}

func (t *Tracer) Reset() {
	t.current = time.Time{}
	t.create = time.Time{}
	t.complete = time.Time{}

	t.currentStage = 0
	t.stages = []string{}
	t.timers = []time.Duration{}
}

func (t *Tracer) Current() {
	t.current = time.Now()
}

func (t *Tracer) Create() {
	t.Reset()
	t.create = time.Now()
	t.Current()
}

func (t *Tracer) Stage(name string) {
	t.timers = append(t.timers, time.Since(t.current))
	t.stages = append(t.stages, name)
	t.currentStage++
	t.Current()
}

func (t *Tracer) Complete() {
	t.complete = time.Now()
	t.Current()
}

func (t *Tracer) IsComplete() bool {
	return !t.complete.IsZero()
}

func (t *Tracer) GetComplete() time.Time {
	return t.complete
}
