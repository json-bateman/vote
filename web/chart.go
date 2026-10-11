package web

import (
	"encoding/json"
	"time"

	"vote/sql/sqlcgen"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
)

const (
	bucketSecs  = 5   // each point on the chart is 5 seconds of votes
	sentBuckets = 720 // always send the last hour; each viewer trims it to their window
)

type Timeline struct {
	Labels                 []string
	Cheers, Boos           []int64
	TotalCheers, TotalBoos int64
}

func windowStart(now time.Time, maxBuckets int64) int64 {
	end := now.Unix() / bucketSecs * bucketSecs // round down to the start of the current bucket (integer division drops the remainder)
	return end - (maxBuckets-1)*bucketSecs
}

func buildTimeline(rows []sqlcgen.GetVotesPerSecondRow, start, maxBuckets int64) Timeline {
	const size = bucketSecs

	t := Timeline{Cheers: make([]int64, maxBuckets), Boos: make([]int64, maxBuckets)}
	for _, r := range rows {
		t.TotalCheers += r.Cheers
		t.TotalBoos += r.Boos
		bucket := r.CreatedAt / size * size // round down to the bucket's start: :17 → :15
		i := (bucket - start) / size        // slots after the first one
		if i >= maxBuckets {
			continue
		}
		t.Cheers[i] += r.Cheers
		t.Boos[i] += r.Boos
	}

	t.Labels = make([]string, maxBuckets)
	for i := range maxBuckets {
		t.Labels[i] = time.Unix(start+int64(i)*size, 0).Local().Format("04:05")
	}

	return t
}

// EChartsOptions returns the ECharts option JSON for the timeline.
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
	l := line.JSON()
	return json.Marshal(l)
}

func lineData(vals []int64) []opts.LineData {
	data := make([]opts.LineData, len(vals))
	for i, v := range vals {
		data[i] = opts.LineData{Value: v}
	}
	return data
}

// endLabelVisitor adds ECharts' endLabel (series name at the end of each line),
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
