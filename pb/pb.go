package pb

import (
	"fmt"
	"time"

	"github.com/schollz/progressbar/v2"
)

type ProgressBar struct {
	*progressbar.ProgressBar
}

func New(text string, max int) *ProgressBar {
	bar := &ProgressBar{}
	bar.ProgressBar = progressbar.NewOptions(
		max,
		progressbar.OptionSetTheme(progressbar.Theme{Saucer: "=", SaucerHead: ">", SaucerPadding: "_", BarStart: "[", BarEnd: "]"}),
		progressbar.OptionThrottle(200*time.Millisecond),
	)
	bar.Describe(text)
	return bar
}

func (bar *ProgressBar) Increment() {
	bar.Add(1)
}

func (bar *ProgressBar) IncrementMax() {
	bar.ChangeMax(bar.GetMax() + 1)
}

func (bar *ProgressBar) Finishln() {
	bar.Finish()
	fmt.Println("")
}
