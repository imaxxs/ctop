package main

import (
	types "./types"
	// "flag"
	"fmt"
	"github.com/nsf/termbox-go"
	// "time"
	// "os"
)

// Logging to a channel (from anywhere):
func log_to_channel(severity string, message string) {
	// Make a new LogMessage struct:
	logMessage := types.LogMessage{
		Severity: severity,
		Message:  message,
	}

	// Put it in the messages channel:
	select {
	case messageChannel <- logMessage:

	default:

	}
}

// Takes metrics off the channel and adds them up:
func handle_metrics() {

	var cfStats types.CFStats

	for {
		// Get a metric from the channel:
		cfMetric := <-metricsChannel
		log_to_channel("debug", fmt.Sprintf("Received a metric! %s", cfMetric.MetricName))

		// Build the key:
		statName := cfMetric.KeySpace + ":" + cfMetric.ColumnFamily

		statsMutex.Lock()
		defer statsMutex.Unlock()

		// See if we already have a stats-entry:
		if _, ok := stats[statName]; ok {
			// Use the existing stats-entry:
			log_to_channel("debug", fmt.Sprintf("Updating existing stat (%s)", statName))
			cfStats = stats[statName]
		} else {
			// Add a new entry to the map:
			log_to_channel("debug", fmt.Sprintf("Adding new stat (%s)", statName))
			cfStats = types.CFStats{
				ReadCount:    0,
				ReadCountTS:  0,
				ReadLatency:  0.0,
				ReadRate:     0.0,
				WriteCount:   0,
				WriteCountTS: 0,
				WriteLatency: 0.0,
				WriteRate:    0.0,
				KeySpace:     cfMetric.KeySpace,
				ColumnFamily: cfMetric.ColumnFamily,
			}
		}

		// Figure out which metric we need to update:
		if cfMetric.MetricName == "ReadCount" {
			// Total read count:
			interval := cfMetric.MetricTimeStamp - cfStats.ReadCountTS
			if cfStats.ReadCountTS == 0 {
				cfStats.ReadRate = 0.0
			} else {
				cfStats.ReadRate = float64(cfMetric.MetricIntValue-cfStats.ReadCount) / float64(interval)
			}
			cfStats.ReadCount = cfMetric.MetricIntValue
			cfStats.ReadCountTS = cfMetric.MetricTimeStamp
			stats[statName] = cfStats

		} else if cfMetric.MetricName == "WriteCount" {
			// Total write count:
			interval := cfMetric.MetricTimeStamp - cfStats.WriteCountTS
			if cfStats.WriteCountTS == 0 {
				cfStats.WriteRate = 0.0
			} else {
				cfStats.WriteRate = float64(cfMetric.MetricIntValue-cfStats.WriteCount) / float64(interval)
			}
			cfStats.WriteCount = cfMetric.MetricIntValue
			cfStats.WriteCountTS = cfMetric.MetricTimeStamp
			stats[statName] = cfStats

		} else if cfMetric.MetricName == "LiveDiskSpaceUsed" {
			// Total disk space used(k):
			cfStats.LiveDiskSpaceUsed = cfMetric.MetricIntValue
			stats[statName] = cfStats

		} else if cfMetric.MetricName == "RecentReadLatencyMicros" {
			// ReadLatency (MicroSeconds):
			if cfMetric.MetricFloatValue > 0 {
				cfStats.ReadLatency = cfMetric.MetricFloatValue / 1000
				stats[statName] = cfStats
			}

		} else if cfMetric.MetricName == "RecentWriteLatencyMicros" {
			// WriteLatency (MicroSeconds):
			if cfMetric.MetricFloatValue > 0 {
				cfStats.WriteLatency = cfMetric.MetricFloatValue / 1000
				stats[statName] = cfStats
			}
		}

		statsMutex.Unlock()

	}

}

// Returns the key-code:
func handle_keypress(ev *termbox.Event) {
	log_to_channel("debug", fmt.Sprintf("Key pressed: %s", ev.Ch))
}
