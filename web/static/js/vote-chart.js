// Draws the results chart with ECharts. The server streams option JSON in the
// $_chartJson signal; this applies it with the colors for the current theme.
// Called from a data-effect, so it reruns whenever the JSON or $theme changes.

// Keep in sync with --vg-cheer / --vg-boo in app.css
const votePalette = {
    light: ["#008300", "#e34948"],
    dark: ["#008300", "#e66767"],
};

window.renderVoteChart = (el, json, theme) => {
    if (!json) return;

    // ECharts themes are fixed at init, so a theme toggle means a fresh instance
    if (el._chart && el._theme !== theme) {
        el._chart.dispose();
        el._chart = null;
    }
    if (!el._chart) {
        el._chart = echarts.init(el, theme === "dark" ? "dark" : null);
        el._theme = theme;
    }
    if (!el._resizeObserver) {
        el._resizeObserver = new ResizeObserver(() => el._chart?.resize());
        el._resizeObserver.observe(el);
    }

    el._chart.setOption({
        ...JSON.parse(json),
        color: votePalette[theme] ?? votePalette.light,
        backgroundColor: "transparent",
        textStyle: { fontFamily: getComputedStyle(document.body).fontFamily },
    });
};
