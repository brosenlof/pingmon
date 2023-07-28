package main

import (
	"os"
	"time"

	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/go-ping/ping"
	"github.com/xeonx/timeago"
)

type PingState struct {
	LastSendTime time.Time
	LastRecvTime time.Time

	pinger *ping.Pinger
	index  int
}

var state map[string]*PingState

func main() {
	addresses := os.Args[1:]

	state = make(map[string]*PingState)

	// setup pingers
	for i, address := range addresses {
		s := PingState{
			index: i,
		}

		if pinger, err := ping.NewPinger(address); err != nil {
			s.pinger = nil
		} else {
			pinger.OnSend = func(*ping.Packet) {
				s.LastSendTime = time.Now()
			}

			pinger.OnRecv = func(*ping.Packet) {
				s.LastRecvTime = time.Now()
			}

			// TODO
			// OnDuplicateRecv func(*Packet)
			// OnSendError func(*Packet, error)
			// OnRecvError func(error)

			s.pinger = pinger
		}

		state[address] = &s
	}

	// start pinging
	for _, s := range state {
		if s.pinger != nil {
			go s.pinger.Run()
		}
	}

	// setup ui
	if err := ui.Init(); err != nil {
		panic(err)
	}
	defer ui.Close()

	go func() {
		table := widgets.NewTable()

		table.SetRect(0, 0, 80, 2*len(state)+3)

		table.Rows = make([][]string, len(state)+1)

		table.Title = "pingmon"
		table.TitleStyle = ui.NewStyle(ui.ColorWhite, ui.ColorClear, ui.ModifierBold)

		table.RowSeparator = true
		table.FillRow = true

		table.BorderStyle = ui.NewStyle(ui.ColorGreen)

		// header row
		table.Rows[0] = []string{"address", "last-send", "last-recv"}

		// status row update loop
		for {
			for address, s := range state {

				table.Rows[s.index+1] = []string{
					address,
					timeago.English.Format(s.LastSendTime),
					timeago.English.Format(s.LastRecvTime),
				}

				if s.LastSendTime.IsZero() {
					table.Rows[s.index+1][1] = "never"
				}

				if s.LastRecvTime.IsZero() {
					table.Rows[s.index+1][2] = "never"
				}

				delta := s.LastRecvTime.Sub(s.LastSendTime)

				// styles
				table.RowStyles[s.index+1] = ui.NewStyle(ui.ColorGreen)

				if delta > 3*time.Second {
					table.RowStyles[s.index+1] = ui.NewStyle(ui.ColorYellow)
				}

				if delta > 5*time.Second || s.LastRecvTime.IsZero() {
					table.RowStyles[s.index+1] = ui.NewStyle(ui.ColorRed)
				}
			}

			ui.Render(table)
			time.Sleep(time.Second)
		}
	}()

	uiEvents := ui.PollEvents()
	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		}
	}
}
