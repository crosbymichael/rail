package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

/*
Init SSL without certificate database
battery.charge: 100
battery.charge.low: 10
battery.charge.warning: 20
battery.mfr.date: CPS
battery.runtime: 1560
battery.runtime.low: 300
battery.type: PbAcid
battery.voltage: 14.2
battery.voltage.nominal: 12
device.mfr: CPS
device.model: UPS OR500
device.type: ups
driver.name: usbhid-ups
driver.parameter.pollfreq: 30
driver.parameter.pollinterval: 5
driver.parameter.port: auto
driver.version: DSM6-2-2-24922-broadwell-fmp-repack-24922-190507
driver.version.data: CyberPower HID 0.3
driver.version.internal: 0.38
input.transfer.high: 140
input.transfer.low: 90
input.voltage: 123.0
input.voltage.nominal: 120
output.voltage: 123.0
ups.beeper.status: enabled
ups.delay.shutdown: 20
ups.delay.start: 30
ups.load: 35
ups.mfr: CPS
ups.model: UPS OR500
ups.productid: 0601
ups.realpower.nominal: 300
ups.status: OL
ups.test.result: Done and passed
ups.timer.shutdown: -60
ups.timer.start: -60
ups.vendorid: 0764
*/
func parseInput(r io.Reader) (map[string]interface{}, error) {
	var (
		s   = bufio.NewScanner(r)
		out = make(map[string]interface{})
	)
	for s.Scan() {
		if err := s.Err(); err != nil {
			return nil, errors.Wrap(err, "scan raw data")
		}
		parts := strings.SplitN(s.Text(), ":", 2)
		if len(parts) == 1 {
			continue
		}
		v := strings.TrimSpace(parts[1])
		k := strings.TrimRight(parts[0], ":")
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			out[k] = v
			continue
		}
		out[k] = f
	}
	return out, nil
}

type ups struct {
	Name    string
	IP      string
	Wattage int
}

func (u *ups) info(ctx context.Context) (map[string]interface{}, error) {
	cmd := exec.CommandContext(ctx, "upsc", fmt.Sprintf("ups@%s", u.IP))
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}
	go cmd.Wait()
	return parseInput(stdout)
}
