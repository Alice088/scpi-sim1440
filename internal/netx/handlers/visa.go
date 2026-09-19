package handlers

import (
	"bufio"
	"log"
	"net"
	"scpi-sim1440/internal/core"
	"scpi-sim1440/internal/noise"
	"strings"
	"time"
)

type VISAHandler struct {
	mng *noise.Manager
	dev core.Device
}

func NewVISAHandler(mng *noise.Manager, dev core.Device) VISAHandler {
	return VISAHandler{
		mng: mng,
		dev: dev,
	}
}

func (h *VISAHandler) Handle(conn net.Conn) {
	defer conn.Close()

	if err := conn.SetReadDeadline(time.Now().Add(5 * time.Minute)); err != nil {
		return
	}

	reader := bufio.NewReader(conn)

	for {
		cmd, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		_ = conn.SetReadDeadline(time.Now().Add(5 * time.Minute))
		cmd = strings.TrimSpace(cmd)

		plan := h.mng.Plan()
		resp := h.dev.Handle(cmd)

		suppress := false

		for _, e := range plan.Device {
			switch e.Kind {
			case noise.DevDelay:
				time.Sleep(e.Delay)
			case noise.DevGarbage:
				resp.Value = "\ufffd#@%$"
			}
		}

		closeAfterWrite := false
		for _, e := range plan.Conn {
			switch e.Kind {
			case noise.ConnTruncate:
				if len(resp.Value) > e.CutAt {
					resp.Value = resp.Value[:e.CutAt]
				}
			case noise.ConnBreak:
				closeAfterWrite = true
			case noise.ConnSilence:
				suppress = true
			}
		}

		if !suppress {
			if _, err := conn.Write([]byte(resp.Value + "\n")); err != nil {
				log.Printf("failed to write response: %s\n", resp.Value)
			}
		}

		if closeAfterWrite {
			return
		}
	}

}
