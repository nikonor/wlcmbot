package worker

type Worker struct {
	workDir string
}

func New(workDir string) *Worker {
	w := Worker{workDir: workDir}
	return &w
}
