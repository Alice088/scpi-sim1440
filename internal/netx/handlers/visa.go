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

func (h *VISAHandler) Call(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		cmd, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		cmd = strings.TrimSpace(cmd)

		plan := h.mng.Plan()
		resp := h.dev.Handle(cmd)

		if plan.Device != nil {
			switch plan.Device.Kind {
			case noise.DevDelay:
				time.Sleep(plan.Device.Delay)
			case noise.DevGarbage:
				resp.Value = "�#@%$"
			}
		}

		if plan.Conn != nil {
			switch plan.Conn.Kind {
			case noise.ConnSilence:
				continue
			case noise.ConnBreak:
				return
			case noise.ConnTruncate:
				if len(resp.Value) > plan.Conn.CutAt { //todo CutAt лучше поменять на % чем на точное число, что отрезать 1/2, 1/3, 1/4
					resp.Value = resp.Value[:plan.Conn.CutAt]
				}
			}
		}

		if _, err := conn.Write([]byte(resp.Value + "\n")); err != nil {
			log.Printf("failed to write response: %s\n", resp.Value)
		}
	}

}
