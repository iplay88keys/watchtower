package runners

import "fmt"

type Restart struct {
	Restart    string `json:"restart"`
	RunCleanup bool   `json:"runCleanup"`

	process Restartable
}

type Restartable interface {
	Restart(runCleanup bool) error
}

func (r *Restart) Setup(process Restartable) {
	r.process = process
}

func (r *Restart) Execute(triggeringFileName string) error {
	fmt.Println("Restarting process:", r.Restart)
	err := r.process.Restart(r.RunCleanup)
	if err != nil {
		return err
	}

	return nil
}
