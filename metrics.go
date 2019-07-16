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
	"context"
	"time"

	"github.com/docker/go-metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	Percent = metrics.Unit("percent")
	Watts   = metrics.Unit("watts")
	Volts   = metrics.Unit("volts")
)

func newCollector(upss []*ups) *collector {
	ns := metrics.NewNamespace("crosbymichael", "rail", nil)
	coll := &collector{
		ns:   ns,
		upss: upss,
		metrics: []*metric{
			{
				name:   "charge",
				help:   "Battery Charge",
				unit:   Percent,
				vt:     prometheus.GaugeValue,
				labels: []string{"name"},
				values: func(s *payload) (v []value) {
					for u, d := range s.data {
						v = append(v, value{
							v: d["battery.charge"].(float64),
							l: []string{u.Name},
						})
					}
					return v
				},
			},
			{
				name:   "load",
				help:   "UPS load",
				unit:   Percent,
				vt:     prometheus.GaugeValue,
				labels: []string{"name"},
				values: func(s *payload) (v []value) {
					for u, d := range s.data {
						v = append(v, value{
							v: d["ups.load"].(float64),
							l: []string{u.Name},
						})
					}
					return v
				},
			},
			{
				name:   "output_voltage",
				help:   "UPS output voltage",
				unit:   Volts,
				vt:     prometheus.GaugeValue,
				labels: []string{"name"},
				values: func(s *payload) (v []value) {
					for u, d := range s.data {
						v = append(v, value{
							v: d["output.voltage"].(float64),
							l: []string{u.Name},
						})
					}
					return v
				},
			},
			{
				name:   "battery_runtime",
				help:   "Battery Runtime",
				unit:   metrics.Seconds,
				vt:     prometheus.GaugeValue,
				labels: []string{"name"},
				values: func(s *payload) (v []value) {
					for u, d := range s.data {
						v = append(v, value{
							v: float64(time.Duration(int64(d["battery.runtime"].(float64))) * time.Second),
							l: []string{u.Name},
						})
					}
					return v
				},
			},
			{
				name:   "status",
				help:   "UPS Status",
				unit:   metrics.Total,
				vt:     prometheus.GaugeValue,
				labels: []string{"name", "status"},
				values: func(s *payload) (v []value) {
					for u, d := range s.data {
						v = append(v, value{
							v: 1.0,
							l: []string{u.Name, d["ups.status"].(string)},
						})
					}
					return v
				},
			},
			{
				name:   "watts",
				help:   "UPS Watts",
				unit:   metrics.Total,
				vt:     prometheus.GaugeValue,
				labels: []string{"name"},
				values: func(s *payload) (v []value) {
					for u := range s.data {
						v = append(v, value{
							v: float64(u.Wattage),
							l: []string{u.Name},
						})
					}
					return v
				},
			},
		},
	}
	ns.Add(coll)
	metrics.Register(ns)
	return coll
}

type collector struct {
	ns      *metrics.Namespace
	upss    []*ups
	metrics []*metric
}

type metric struct {
	name   string
	help   string
	unit   metrics.Unit
	vt     prometheus.ValueType
	labels []string
	values func(*payload) []value
}

type value struct {
	v float64
	l []string
}

type payload struct {
	data map[*ups]map[string]interface{}
}

func (m *metric) desc(ns *metrics.Namespace) *prometheus.Desc {
	// the namespace label is for containerd namespaces
	return ns.NewDesc(m.name, m.help, m.unit, m.labels...)
}

func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	for _, m := range c.metrics {
		ch <- m.desc(c.ns)
	}
}

func (c *collector) Collect(ch chan<- prometheus.Metric) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	data := make(map[*ups]map[string]interface{}, len(c.upss))
	for _, u := range c.upss {
		d, err := u.info(ctx)
		if err != nil {
			logrus.WithError(err).Error("ups info")
			continue
		}
		data[u] = d
	}
	payload := &payload{
		data: data,
	}
	for _, m := range c.metrics {
		values := m.values(payload)
		for _, v := range values {
			ch <- prometheus.MustNewConstMetric(m.desc(c.ns), m.vt, v.v, v.l...)
		}
	}
}
