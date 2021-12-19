(()=>{
    const updateMetrics = async () => {
        const response = await fetch("/api/metrics");
        const metrics = await response.json();

        $("#status span").text(metrics.status);
        $("#burning span").text(metrics.patientBurning ? "ON" : "OFF");
        $("#mean_timming span").text(`${toMilli(metrics.meanTimming)}ms`);
        $("#last_timming span").text(`${toMilli(metrics.lastTimming)}ms`);
        const prog = $("#prognosis")
        prog.text(metrics.prognosis)
        prog.removeClass(prog.attr('class')).addClass(prognosisClass(metrics.prognosis));
    };

    const toMilli = (duration) => Math.round(duration/1000000.0) * 10 / 10;

    const prognosisClass = (prognosis) => prognosis.includes("good") ? "text-success" : "text-danger";

    setInterval(updateMetrics, 2000);   
})();

function startBurn(){
    fetch("/api/start");
}

function stopBurn() {
    fetch("/api/stop");
}