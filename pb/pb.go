package pb

import (
	"fmt"
	"log"
	"time"

	"github.com/schollz/progressbar/v2"
)

type ProgressBar struct {
	*progressbar.ProgressBar
}

func New(text string, max int) *ProgressBar {
	bar := &ProgressBar{}
	if max == 0 {
		max = 1
	}
	bar.ProgressBar = progressbar.NewOptions(
		max,
		progressbar.OptionSetTheme(progressbar.Theme{Saucer: "=", SaucerHead: ">", SaucerPadding: "_", BarStart: "[", BarEnd: "]"}),
		progressbar.OptionThrottle(200*time.Millisecond),
	)
	bar.Describe(text)
	err := bar.RenderBlank()
	if err != nil {
		log.Println(err)
	}
	return bar
}

func (bar *ProgressBar) Increment() {
	err := bar.Add(1)
	if err != nil {
		log.Println(err)
	}
}

func (bar *ProgressBar) IncrementMax() {
	bar.ChangeMax(bar.GetMax() + 1)
}

func (bar *ProgressBar) Finishln() {
	err := bar.Finish()
	if err != nil {
		log.Println(err)
	}
	fmt.Println("")
}
