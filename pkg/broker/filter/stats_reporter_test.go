/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package filter

import (
	"net/http"
	"testing"
	"time"

	"knative.dev/pkg/metrics/metricskey"
	"knative.dev/pkg/metrics/metricstest"
)

func TestStatsReporter(t *testing.T) {
	setup()
	args := &ReportArgs{
		ns:           "testns",
		trigger:      "testtrigger",
		broker:       "testbroker",
		filterType:   "testeventtype",
		filterSource: "testeventsource",
	}

	r := NewStatsReporter()

	wantTags := map[string]string{
		metricskey.LabelNamespaceName: "testns",
		metricskey.LabelName:          "testtrigger",
		metricskey.LabelBrokerName:    "testbroker",
		metricskey.LabelFilterType:    "testeventtype",
		metricskey.LabelFilterSource:  "testeventsource",
	}

	wantAllTags := map[string]string(wantTags)
	wantAllTags[metricskey.LabelResponseCode] = "202"
	wantAllTags[metricskey.LabelResponseCodeClass] = "2xx"

	// test ReportEventCount
	expectSuccess(t, func() error {
		return r.ReportEventCount(args, http.StatusAccepted)
	})
	expectSuccess(t, func() error {
		return r.ReportEventCount(args, http.StatusAccepted)
	})
	metricstest.CheckCountData(t, "event_count", wantAllTags, 2)

	// test ReportEventDispatchTime
	expectSuccess(t, func() error {
		return r.ReportEventDispatchTime(args, http.StatusAccepted, 1100*time.Millisecond)
	})
	expectSuccess(t, func() error {
		return r.ReportEventDispatchTime(args, http.StatusAccepted, 9100*time.Millisecond)
	})
	metricstest.CheckDistributionData(t, "event_dispatch_latencies", wantAllTags, 2, 1100.0, 9100.0)

	// test ReportEventProcessingTime
	expectSuccess(t, func() error {
		return r.ReportEventProcessingTime(args, 1000*time.Millisecond)
	})
	expectSuccess(t, func() error {
		return r.ReportEventProcessingTime(args, 8000*time.Millisecond)
	})
	metricstest.CheckDistributionData(t, "event_processing_latencies", wantTags, 2, 1000.0, 8000.0)
}

func TestReporterEmptySourceAndTypeFilter(t *testing.T) {
	setup()

	args := &ReportArgs{
		ns:           "testns",
		trigger:      "testtrigger",
		broker:       "testbroker",
		filterType:   "",
		filterSource: "",
	}

	r := NewStatsReporter()

	wantTags := map[string]string{
		metricskey.LabelNamespaceName:     "testns",
		metricskey.LabelName:              "testtrigger",
		metricskey.LabelBrokerName:        "testbroker",
		metricskey.LabelFilterType:        anyValue,
		metricskey.LabelFilterSource:      anyValue,
		metricskey.LabelResponseCode:      "202",
		metricskey.LabelResponseCodeClass: "2xx",
	}

	// test ReportEventCount
	expectSuccess(t, func() error {
		return r.ReportEventCount(args, http.StatusAccepted)
	})
	expectSuccess(t, func() error {
		return r.ReportEventCount(args, http.StatusAccepted)
	})
	expectSuccess(t, func() error {
		return r.ReportEventCount(args, http.StatusAccepted)
	})
	expectSuccess(t, func() error {
		return r.ReportEventCount(args, http.StatusAccepted)
	})
	metricstest.CheckCountData(t, "event_count", wantTags, 4)
}

func expectSuccess(t *testing.T, f func() error) {
	t.Helper()
	if err := f(); err != nil {
		t.Errorf("Reporter expected success but got error: %v", err)
	}
}

func setup() {
	resetMetrics()
}

func resetMetrics() {
	// OpenCensus metrics carry global state that need to be reset between unit tests.
	metricstest.Unregister("event_count", "event_dispatch_latencies", "event_processing_latencies")
	register()
}
