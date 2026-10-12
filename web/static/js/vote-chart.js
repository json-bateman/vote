// <vote-chart>: draws the results chart with ECharts. The server streams option JSON
// in the $_chartJson signal, which p_results.templ binds to the `option` attribute.
// echarts is the global from the vendored echarts.min.js.
import { rocket } from 'datastar'

// Keep in sync with --vg-cheer / --vg-boo in app.css
const votePalette = {
    default: ['#008300', '#e34948'],
    dark: ['#008300', '#e66767'],
}

rocket('vote-chart', {
    renderOnPropChange: false,
    props: ({ json, string, number }) => ({
        option: json,
        theme: string.default('default'),
        // Points to show; the server always sends the last hour
        buckets: number.min(1).default(60),
        resizeDelay: number.min(0).default(100),
    }),
    onFirstRender: ({ refs, cleanup, host, observeProps, props }) => {
        let chart
        let observer
        let resizeTimer = 0

        const update = () => {
            if (!props.option?.series) return
            const option = structuredClone(props.option)
            for (const axis of [].concat(option.xAxis ?? [])) axis.data = axis.data?.slice(-props.buckets)
            for (const s of option.series) s.data = s.data?.slice(-props.buckets)
            chart?.setOption({
                ...option,
                color: votePalette[props.theme] ?? votePalette.default,
                backgroundColor: 'transparent',
                textStyle: { fontFamily: getComputedStyle(host).fontFamily },
            })
        }
        const resize = () => {
            clearTimeout(resizeTimer)
            resizeTimer = setTimeout(() => chart?.resize(), props.resizeDelay)
        }
        // ECharts themes are fixed at init, so a theme change means a fresh instance
        const init = () => {
            if (!(refs.container instanceof HTMLElement)) return
            chart?.dispose()
            chart = echarts.init(
                refs.container,
                props.theme === 'default' ? undefined : props.theme,
            )
            update()
            observer?.disconnect()
            observer = new ResizeObserver(resize)
            observer.observe(refs.container)
        }

        init()
        observeProps(update, 'option', 'buckets')
        observeProps(init, 'theme')
        observeProps(resize, 'resizeDelay')

        cleanup(() => {
            observer?.disconnect()
            clearTimeout(resizeTimer)
            chart?.dispose()
        })
    },
    // The host's size comes from the page (flex-grow-1, min-height); the container fills it
    render: ({ html }) => html`
    <style>
      :host {
        display: block;
        position: relative;
      }

      .chart-container {
        position: absolute;
        inset: 0;
      }
    </style>
    <div class="chart-container" data-ref:container></div>
  `,
})
