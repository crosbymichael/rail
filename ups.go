/*
	Copyright (c) 2019 @crosbymichael

	Permission is hereby granted, free of charge, to any person
	obtaining a copy of this software and associated documentation
	files (the "Software"), to deal in the Software without
	restriction, including without limitation the rights to use, copy,
	modify, merge, publish, distribute, sublicense, and/or sell copies
	of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be
	included in all copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
	EXPRESS OR IMPLIED,
	INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
	IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
	HOLDERS BE LIABLE FOR ANY CLAIM,
	DAMAGES OR OTHER LIABILITY,
	WHETHER IN AN ACTION OF CONTRACT,
	TORT OR OTHERWISE,
	ARISING FROM, OUT OF OR IN CONNECTION WITH
	THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/docker/go-metrics"
	"github.com/getsentry/raven-go"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()
	app.Name = "rail"
	app.Version = "1"
	app.Usage = "monitor your UPS"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug output in the logs",
		},
		cli.StringFlag{
			Name:  "metrics,m",
			Usage: "prometheus metrics address",
			Value: "127.0.0.1:9930",
		},
		cli.StringFlag{
			Name:   "sentry-dsn",
			Usage:  "sentry DSN",
			EnvVar: "SENTRY_DSN",
		},
		cli.StringSliceFlag{
			Name:  "ups",
			Usage: "ups name:ip:watts",
			Value: &cli.StringSlice{},
		},
	}
	app.Before = func(clix *cli.Context) error {
		if clix.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}
		if dsn := clix.GlobalString("sentry-dsn"); dsn != "" {
			raven.SetDSN(dsn)
			raven.DefaultClient.SetRelease(app.Version)
		}
		return nil
	}
	app.Action = func(clix *cli.Context) error {
		var upss []*ups
		for _, sv := range clix.GlobalStringSlice("ups") {
			parts := strings.Split(sv, ":")
			w, err := strconv.Atoi(parts[2])
			if err != nil {
				return err
			}
			upss = append(upss, &ups{
				Name:    parts[0],
				IP:      parts[1],
				Wattage: w,
			})
		}
		newCollector(upss)
		return http.ListenAndServe(clix.GlobalString("metrics"), metrics.Handler())
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		raven.CaptureErrorAndWait(err, nil)
		os.Exit(1)
	}
}
