package thyme

import (
	"fmt"
	"regexp"
	"strconv"
	"time"
	"io/ioutil"
)

func init() {
	RegisterTracker("linux", NewLinuxTracker)
}

// LinuxTracker tracks application usage on Linux via a few standard command-line utilities.
type LinuxTracker struct{}

var _ Tracker = (*LinuxTracker)(nil)

func NewLinuxTracker() Tracker {
	return &LinuxTracker{}
}

func (t *LinuxTracker) Deps() string {
	return `
Install the following command-line utilities via your package manager of choice:
* xdpyinfo
* xwininfo
* xdotool
* wmctrl

For example:
* Debian: apt-get install x11-utils xdotool wmctrl

Note: this command prints out this message regardless of whether the dependencies are already installed.
`
}

func (t *LinuxTracker) Snap() (*Snapshot, error) {
	var windows []*Window
	{
		out, err := ioutil.ReadFile("/tmp/emacs-active-window")
		if err != nil {
			return nil, fmt.Errorf("reading /tmp/emacs-active-window failed with error: %s.", err)
		}

		var window = Window{ID: 0, Desktop: -1, Name: string(out)}
		windows = append(windows, &window)
	}

	return &Snapshot{Windows: windows, Active: 0, Visible: nil, Time: time.Now()}, nil
}

// isVisible checks if the window is visible in the current viewport.
// x and y are assumed to be relative to the current viewport (i.e.,
// (0, 0) is the coordinate of the top-left corner of the current
// viewport.
func isVisible(x, y, w, h, viewHeight, viewWidth int) bool {
	return (0 <= x && x < viewWidth && 0 <= y && y < viewHeight) ||
		(0 <= x+w && x+w < viewWidth && 0 <= y && y < viewHeight) ||
		(0 <= x && x < viewWidth && 0 <= y+h && y+h < viewHeight) ||
		(0 <= x+w && x+w < viewWidth && 0 <= y+h && y+h < viewHeight)
}

var (
	dimRx = regexp.MustCompile(`dimensions:\s+([0-9]+)x([0-9]+)\s+pixels`)
	xRx   = regexp.MustCompile(`Absolute upper\-left X:\s+(\-?[0-9]+)`)
	yRx   = regexp.MustCompile(`Absolute upper\-left Y:\s+(\-?[0-9]+)`)
	wRx   = regexp.MustCompile(`Width:\s+([0-9]+)`)
	hRx   = regexp.MustCompile(`Height:\s+([0-9]+)`)
)

// parseWinDim parses window dimension info from the output of `xwininfo`
func parseWinDim(rx *regexp.Regexp, out string, varname string) (int, error) {
	if matches := rx.FindStringSubmatch(out); len(matches) == 2 {
		n, err := strconv.Atoi(matches[1])
		if err != nil {
			return 0, err
		}
		return n, nil
	} else {
		return 0, fmt.Errorf("could not parse window %s from output %s", varname, out)
	}

}
