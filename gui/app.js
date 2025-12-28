// ==========================================
// 1. CONFIGURATION & STATE
// ==========================================
const API_ENDPOINT = `/telemetry/read?format=json`;
const MONO_FONT = "Monaco, 'Courier New', monospace";

let CURRENT_PRN = null;
let LAST_TOW = null;
let LAST_WEEK = null;

// ==========================================
// 2. UTILITY FUNCTIONS
// ==========================================

async function remoteLog(...args) {
    const message = args.map(arg => typeof arg === "object" ? JSON.stringify(arg) : arg).join(" ");
    try {
        await fetch("/guilog", { method: "POST", body: message });
    } catch (e) {
        originalLog("Remote logging failed:", e);
    }
}
const originalLog = console.log;
console.log = function(...args) {
    originalLog.apply(console, args);
    remoteLog(...args);
};

function getSatMetadata(id) {
    let prefix = "U", displayNum = id, color = "#757575";
    if (id >= 0 && id <= 31) { prefix = "G"; displayNum = id + 1; color = "#00c853"; } 
    else if (id >= 32 && id <= 61) { prefix = "E"; displayNum = id - 31; color = "#2979ff"; } 
    else if (id >= 64 && id <= 95) { prefix = "R"; displayNum = id - 63; color = "#ff5252"; } 
    else if (id >= 96 && id <= 158) { prefix = "B"; displayNum = id - 95; color = "#ffd600"; } 
    else if (id >= 159 && id <= 163) { prefix = "J"; displayNum = id - 158; color = "#e040fb"; }
    else if (id >= 164 && id <= 170) { prefix = "I"; displayNum = id - 163; color = "#ff9100"; }
    return { label: `${prefix}${displayNum.toString().padStart(2, "0")}`, color: color };
}

function gps2utcTime(week, tow) {
    const epoch = new Date(1980, 0, 6, 0, 0, 0); // Corrected Month (0 = Jan)
    const seconds = week * 604800 + tow - 18; 
    epoch.setSeconds(epoch.getSeconds() + seconds);
    return epoch;
}

function formatDate(date) {
    const pad = (n) => n.toString().padStart(2, "0");
    return {
        date: `${pad(date.getMonth())}/${pad(date.getDate())}/${date.getFullYear()}`,
        time: `${pad(date.getHours())}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`
    };
}

// ==========================================
// 3. COMPONENT INITIALIZATION (With Guards)
// ==========================================

// --- MAP SETUP ---
let map, positionMarker;
const mapEl = document.getElementById("map-container");
if (mapEl) {
    map = L.map("map-container").setView([32.584, -85.496], 16);
    L.tileLayer("https://server.arcgisonline.com/ArcGIS/rest/services/World_Imagery/MapServer/tile/{z}/{y}/{x}", {
        attribution: "Esri"
    }).addTo(map);

    const arrowIcon = L.divIcon({
        className: "heading-arrow",
        html: "<svg width='24' height='24' viewBox='0 0 24 24' fill='red' stroke='white'><polygon points='12 2 22 22 12 18 2 22 12 2'/></svg>",
        iconSize: [24, 24], iconAnchor: [12, 12]
    });
    positionMarker = L.marker([32.584, -85.496], {icon: arrowIcon}).addTo(map);
}

// --- SHARED PLOTLY LAYOUT ---
const layoutDefaults = {
    font: { family: MONO_FONT, color: "#e0e0e0" },
    paper_bgcolor: "#3c3c3c", plot_bgcolor: "#3c3c3c",
    margin: { t: 40, b: 40, l: 50, r: 20 }
};

// --- CHART INITIALIZATION ---
if (document.getElementById("chart-bar-container")) {
    Plotly.newPlot("chart-bar-container", [{ type: "bar", x: [], y: [] }], 
        { ...layoutDefaults, 
            title: { text: "C/No [dB-Hz]", font: { family: MONO_FONT, size: 14, weight: "bold" } }, 
            xaxis: { title: "SV ID", color: "#e0e0e0", tickfont: { color: "#e0e0e0" }, gridcolor: "#555" },
            yaxis: { range: [0, 55], color: "#e0e0e0", tickfont: { color: "#e0e0e0" }, gridcolor: "#555" }
    } );
}

if (document.getElementById("chart-polar-container")) {
    Plotly.newPlot("chart-polar-container", [{
        type: "scatterpolar", mode: "markers+text", r: [], theta: [], text: [],
        marker: { size: 18, line: { color: "white", width: 1 } }, textfont: { color: "white", size: 9, family: MONO_FONT, weight: "bold" }
    }], { ...layoutDefaults, polar: { bgcolor: "#3c3c3c", 
        radialaxis: { showticklabels: true, tickmode: "array", tickvals: [90, 60, 30, 0], ticktext: ["90°", "60°", "30°", "0°"], range: [90, 0], ticks: "outside", gridcolor: "#555", angle: 180, tickangle: 180, gridcolor: "#555" },
        angularaxis: { rotation: 90, direction: "clockwise", tickvals: [0, 30, 60, 90, 120, 150, 180, 210, 240, 270, 300, 330], ticktext: ["N", "30°", "60°", "E", "120°", "150°", "S", "210°", "240°", "W", "300°", "330°"], color: "#e0e0e0", gridcolor: "#555" } 
    }});
}

// --- SATELLITE VIEW PLOTS ---
const satPlots = [
    {id: "plot-doppler", title: "Doppler History", xlabel: "ToW [s]", ylabel: "Doppler [Hz]", mode: "lines"}, 
    {id: "plot-cno", title: "C/No History", xlabel: "ToW [s]", ylabel: "C/No [dB-Hz]", mode: "lines"}, 
    {id: "plot-pseudorange", title: "Pseudorange History", xlabel: "ToW [s]", ylabel: "Pseudorange [m]", mode: "lines"}, 
    {id: "plot-correlator", title: "Correlators History", xlabel: "In-Phase", ylabel: "Quadrature-Phase", mode: "markers"}
];
satPlots.forEach(plot => {
    const el = document.getElementById(plot.id);
    if (el) {
        // Special handling for correlator which needs 3 traces (Prompt, Early, Late)
        const traces = (plot.id === "plot-correlator") 
            ? [{x:[], y:[], name: "Early", mode: plot.mode}, {x:[], y:[], name: "Prompt", mode: plot.mode}, {x:[], y:[], name: "Late", mode: plot.mode}]
            : [{x:[], y:[], mode: plot.mode}];

        Plotly.newPlot(plot.id, traces, { 
            ...layoutDefaults, 
            title: { text: plot.title, font: { family: MONO_FONT, size: 16 } },
            xaxis: { title: {text: plot.xlabel, font: {size: 12}}, color: "#e0e0e0", gridcolor: "#444", tickformat: 'd', separatethousands: false, nticks: 6},
            yaxis: { title: {text: plot.ylabel, font: {size: 12}}, color: "#e0e0e0", gridcolor: "#444", autorange: true },
            legend: { x: 1, y: 1, xanchor: 'right', yanchor: 'top', bgcolor: "#555", bordercolor: "#e0e0e0", borderwidth: 1 }
        }, {responsive: true});
    }
});

// ==========================================
// 4. DATA PROCESSING & UPDATES
// ==========================================

async function fetchAndUpdateNavigationData() {
    try {
        const response = await fetch(API_ENDPOINT);
        if (!response.ok) throw new Error(`HTTP error! status: ${response.status}`);
        const data = await response.json();
        
        // Handle if your Go server returns an array or a single object
        const payload = Array.isArray(data) ? data[0] : data;

        if (payload.navigation) {
            LAST_TOW = payload.navigation.tow;
            LAST_WEEK = payload.navigation.week;
            updateNavigationPanel(payload.navigation);
            if (map) updateNavigationMap(payload.navigation);
        }
        
        if (payload.satellites) {
            updateNavigationCharts(payload.satellites);
            const selector = document.getElementById("sat-selector");
            if (selector) refreshSatelliteList(payload.satellites);
        }

    } catch (error) {
        console.error("Fetch error:", error);
    }
}

function updateNavigationPanel(nav) {
    const timeEl = document.getElementById("val-time");
    if (!timeEl) return;

    const {date, time} = nav.tow ? formatDate(gps2utcTime(nav.week, nav.tow)) : {date: "--", time: "--"};
    timeEl.textContent = time;
    document.getElementById("val-date").textContent = date;
    document.getElementById("val-time").textContent = time;
    document.getElementById("val-lat").textContent = nav.latitude ? nav.latitude.toFixed(8) : "--";
    document.getElementById("val-lon").textContent = nav.longitude ? nav.longitude.toFixed(8) : "--";
    document.getElementById("val-alt").textContent = nav.altitude ? nav.altitude.toFixed(2) : "--";
    document.getElementById("val-speed").textContent = nav.vn ? Math.sqrt(nav.vn**2 + nav.ve**2 + nav.vd**2).toFixed(3) : "--";
    document.getElementById("val-heading").textContent = nav.yaw ? nav.yaw.toFixed(2) : "--";
    document.getElementById("val-pdop").textContent = nav.pdop ? nav.pdop.toFixed(3) : "--";
    document.getElementById("val-hdop").textContent = nav.hdop ? nav.hdop.toFixed(3) : "--";
    document.getElementById("val-vdop").textContent = nav.vdop ? nav.vdop.toFixed(3) : "--";
    document.getElementById("val-nsat").textContent = nav.n_sat ? nav.n_sat : "--";
}

function updateNavigationMap(nav) {
    if (!nav.latitude || !nav.longitude) return;
    const latlng = [nav.latitude, nav.longitude];
    positionMarker.setLatLng(latlng);
    map.panTo(latlng);

    const svg = positionMarker.getElement()?.querySelector("svg");
    if (svg && nav.yaw !== undefined) {
        svg.style.transform = `rotate(${nav.yaw}deg)`;
    }
}

function updateNavigationCharts(sats) {
    if (!document.getElementById("chart-bar-container")) return;

    const meta = sats.map(s => getSatMetadata(s.prn));
    Plotly.restyle("chart-bar-container", {
        x: [meta.map(m => m.label)],
        y: [sats.map(s => s.cno)],
        "marker.color": [meta.map(m => m.color)]
    }, [0]);

    Plotly.restyle("chart-polar-container", {
        r: [sats.map(s => s.elevation)],
        theta: [sats.map(s => s.azimuth)],
        text: [meta.map(m => m.label)],
        "marker.color": [meta.map(m => m.color)]
    }, [0]);
}

// ==========================================
// 5. SATELLITE HISTORY LOGIC
// ==========================================

function refreshSatelliteList(sv) {
    const selector = document.getElementById("sat-selector");
    sv.forEach(s => {
        const meta = getSatMetadata(s.prn);
        if (!Array.from(selector.options).some(opt => String(opt.value) == String(s.prn))) {
            const opt = document.createElement("option");
            opt.value = s.prn;
            opt.textContent = meta.label;
            selector.appendChild(opt);
        }
    });
}

const selector = document.getElementById("sat-selector");
if (selector) {
    selector.addEventListener("change", (e) => {
        CURRENT_PRN = e.target.value;
        if (LAST_TOW) {
            fetchAndPlotSatelliteHistory(CURRENT_PRN, LAST_WEEK, LAST_TOW - 100);
        }
    });
}

async function fetchAndPlotSatelliteHistory(prn, week, tow) {
    const url = `/satellite/read?format=json&prn=${prn}&week=${week}&tow=${tow}`;
    try {
        const res = await fetch(url);
        const history = await res.json();
        
        const times = history.map(h => h.tow);
        Plotly.restyle("plot-doppler", { x: [times], y: [history.map(h => h.doppler)] }, [0]);
        Plotly.restyle("plot-cno", { x: [times], y: [history.map(h => h.cno)] }, [0]);
        Plotly.restyle("plot-pseudorange", { x: [times], y: [history.map(h => h.psr)] }, [0]);
        
        Plotly.restyle("plot-correlator", {
            x: [history.map(h => h.ie), history.map(h => h.ip), history.map(h => h.il)],
            y: [history.map(h => h.qe), history.map(h => h.qp), history.map(h => h.ql)]
        }, [0, 1, 2]);
    } catch (e) { console.error(e); }
}

// ==========================================
// 6. STARTUP
// ==========================================
setInterval(fetchAndUpdateNavigationData, 500);
fetchAndUpdateNavigationData(); // The function name now matches