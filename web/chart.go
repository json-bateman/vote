package web

import (
	"encoding/json"
	"time"

	"vote/sql/sqlcgen"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

const (
	bucketSecs = 5  // each point on the chart is 5 seconds of votes
	maxBuckets = 60 // rolling window: the chart shows the last 5 minutes
)

// Timeline is the votes bucketed into 5s chunks for the chart, plus all-time totals.
type Timeline struct {
	Labels                 []string
	Cheers, Boos           []int64
	TotalCheers, TotalBoos int64
}

// buildTimeline fills a fixed 5-minute window ending at now, so the chart
// scrolls left every 5s even when nobody is voting. Totals still count every
// vote, not just the window's.
func buildTimeline(rows []sqlcgen.GetVotesPerSecondRow, now time.Time) Timeline {
	const size = bucketSecs
	end := now.Unix() / size * size
	start := end - (maxBuckets-1)*size

	t := Timeline{Cheers: make([]int64, maxBuckets), Boos: make([]int64, maxBuckets)}
	for _, r := range rows {
		t.TotalCheers += r.Cheers
		t.TotalBoos += r.Boos
		i := (r.CreatedAt/size*size - start) / size
		if i < 0 || i >= maxBuckets { // older than the window, or clock skew
			continue
		}
		t.Cheers[i] += r.Cheers
		t.Boos[i] += r.Boos
	}

	t.Labels = make([]string, maxBuckets)
	for i := range maxBuckets {
		t.Labels[i] = time.Unix(start+int64(i)*size, 0).Local().Format("15:04:05")
	}
	return t
}

// EChartsOptions returns the ECharts option JSON for the timeline. Colors and
// fonts are left out on purpose: the page applies them per theme, so the server
// doesn't need to know whether the viewer is in light or dark mode.
func (t Timeline) EChartsOptions() ([]byte, error) {
	line := charts.NewLine()
	line.SetGlobalOptions(
		charts.WithTooltipOpts(opts.Tooltip{Trigger: "axis"}),
		// The totals above the chart already act as the legend
		charts.WithLegendOpts(opts.Legend{Show: opts.Bool(false)}),
		// Left padding keeps the first time label, centered on the axis start, from clipping
		charts.WithGridOpts(opts.Grid{Left: "28", Right: "64", Top: "16", Bottom: "0", ContainLabel: opts.Bool(true)}),
		charts.WithXAxisOpts(opts.XAxis{Type: "category", BoundaryGap: opts.Bool(false)}),
		charts.WithYAxisOpts(opts.YAxis{Type: "value", MinInterval: 1}),
	)

	series := []charts.SeriesOpts{
		charts.WithLineChartOpts(opts.LineChart{ShowSymbol: opts.Bool(false)}),
		charts.WithLineStyleOpts(opts.LineStyle{Width: 2}),
		charts.WithAreaStyleOpts(opts.AreaStyle{Opacity: opts.Float(0.1)}),
	}
	line.SetXAxis(t.Labels).
		AddSeries("Cheer", lineData(t.Cheers), series...).
		AddSeries("Boo", lineData(t.Boos), series...)

	line.Accept(endLabelVisitor{})
	line.Validate()
	return json.Marshal(line.JSON())
}

func lineData(vals []int64) []opts.LineData {
	data := make([]opts.LineData, len(vals))
	for i, v := range vals {
		data[i] = opts.LineData{Value: v}
	}
	return data
}

// endLabelVisitor adds ECharts' endLabel (series name at the end of each line),
// which go-echarts has no option for. Green vs red is hard to tell apart for
// red-blind viewers, so the lines carry their names as well as their colors.
type endLabelVisitor struct {
	charts.BaseConfigurationVisitor
}

func (endLabelVisitor) VisitSeriesOpt(series charts.MultiSeries) any {
	type withEndLabel struct {
		charts.SingleSeries
		EndLabel    map[string]any `json:"endLabel"`
		LabelLayout map[string]any `json:"labelLayout"`
	}
	out := make([]withEndLabel, len(series))
	for i, s := range series {
		out[i] = withEndLabel{
			SingleSeries: s,
			// "inherit" colors each label like its line
			EndLabel: map[string]any{"show": true, "formatter": "{a}", "color": "inherit", "fontWeight": "bold", "lineHeight": 20},
			// When both lines end at the same value, stack the labels instead of overlapping them
			LabelLayout: map[string]any{"moveOverlap": "shiftY"},
		}
	}
	return out
}
