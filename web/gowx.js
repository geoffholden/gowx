function getQueryVariable(variable) {
    var query = window.location.search.substring(1);
    var vars = query.split("&");
    for (var i=0;i<vars.length;i++) {
        var pair = vars[i].split("=");
        if(pair[0] == variable){return pair[1];}
    }
    return(false);
}

function populateCurrentData() {
    $.getJSON("/currentdata.json", function(data) {
        $('#current_temp').html(data['Temperature'].toFixed(1));
        $('#current_humidity').html(data['Humidity'].toFixed(0));
        $('#current_pressure').html(data['Pressure'].toFixed(1));
        $('#current_wind').html(data['Wind'].toFixed(1));
        $('#current_wind_dir').html(degreesToCardinal(data['WindDir']));
        $('#current_wind_angle').css("transform", "rotate(" + (data['WindDir'] + 90) + "deg)");
        $('#current_rain').html(data['RainRate'].toFixed(2));
    });
    $.getJSON("/change.json?key=Pressure&type=pressure&time=3h", function(data) {
        if (data.Change[0] >= 0.1) {
            result = "Rising";
        } else if (data.Change[0] <= -0.1) {
            result = "Falling";
        } else {
            result = "Steady";
        }

        abs = Math.abs(data.Change[0]);
        if (abs >= 6) {
            result += "<br />(Very Rapidly)";
        } else if (abs >= 3.6) {
            result += "<br />(Quickly)";
        } else if (abs >= 1.6) {
        } else if (abs >= 0.1) {
            result += "<br />(Slowly)";
        } else {
            result += "<br />&nbsp;";
        }
        $('#pressure_tendency').html(result);
    });
    $.getJSON("/change.json?key=RainTotal&type=rain&time=24h", function(data) {
        $('#rain_total').html(data.Change[0].toFixed(2));
    });
    setTimeout(populateCurrentData, 30000);
}

function degreesToCardinal(angle) {
    switch (angle) {
        case 0:
            return "N";
        case 22.5:
            return "NNE";
        case 45:
            return "NE";
        case 67.5:
            return "ENE";
        case 90:
            return "E";
        case 112.5:
            return "ESE";
        case 135:
            return "SE";
        case 157.5:
            return "SSE";
        case 180:
            return "S";
        case 202.5:
            return "SSW";
        case 225:
            return "SW";
        case 247.5:
            return "WSW";
        case 270:
            return "W";
        case 292.5:
            return "WNW";
        case 315:
            return "NW";
        case 337.5:
            return "NNW";
        case 360:
            return "N";
        default:
            return angle;
    }
}

function makeChart(query, container, name, units, showRange) {
    showRange = typeof showRange !== 'undefined' ? showRange : true;
    var options = {
        chart: {
            renderTo: container,
            zoomType: 'x',
        },
        title: {
            text: name,
        },
        xAxis: {
            type: 'datetime',
            events: {
                setExtremes: syncExtremes,
            },
        },
        yAxis: {
            title: {
                text: name + " (" + units + ")",
            },
            minPadding: 0,
            maxPadding: 0,
        },
        tooltip: {
            valueDecimals: 1,
            valueSuffix: " " + units,
        },
    };
    $.getJSON(query, function(data) {
        options.series = [];
        for (var i = 0; i < data.Data.length; i++) {
            title = name
            options.series.push({
                type: 'spline',
                data: data.Data[i],
                name: data.Label[i],
                marker: {enabled: false,},
            });
            if (showRange) {
                options.series.push({
                    type: 'areasplinerange',
                    data: data.Errorbars[i],
                    enableMouseTracking: false,
                    color: Highcharts.getOptions().colors[i],
                    fillOpacity: 0.3,
                    zIndex: 0,
                    linkedTo: ":previous",
                    visible: showRange,
                });
            }
        }
        var chart = new Highcharts.Chart(options);

        setInterval(function() {
            $.getJSON(query, function(data) {
                for (var i = 0; i < data.Data.length; i++) {
                    if (showRange) {
                        chart.series[i * 2].setData(data.Data[i]);
                        chart.series[i * 2 + 1].setData(data.Errorbars[i]);
                    } else {
                        chart.series[i].setData(data.Data[i]);
                    }
                }
            });
        },300000);
    });
}

function circularMean(data) {
    var sums = 0;
    var sumc = 0;
    for (var i = 0; i < data.length; i++) {
        var radians = data[i] * Math.PI / 180.0;
        sums += Math.sin(radians);
        sumc += Math.cos(radians);
    }

    return Math.atan2(sums/data.length, sumc/data.length) * 180.0 / Math.PI;
}

function generateBarbData(mag, dir, num) {
    function reducer(map, entry, index) {
        map[entry[0]] = index;
        return map;
    }
    var mag_map = mag.reduce(reducer, {});
    var dir_map = dir.reduce(reducer, {});

    for (var i = mag.length - 1; i >= 0; i--) {
        if (!(mag[i][0] in dir_map)) {
            mag.splice(i, 1);
        }
    }
    for (var i = dir.length - 1; i >= 0; i--) {
        if (!(dir[i][0] in mag_map)) {
            dir.splice(i, 1);
        }
    }

    var data = mag.map(function(val, i) { return [val[0], val[1], dir[i][1]]; });

    // Need to average the barbdata
    var starttime = data[0][0];
    var endtime = data[data.length - 1][0];
    var binsize = (endtime - starttime) / num;

    var t = starttime + binsize;
    var bin_time = [];
    var bin_speed = [];
    var bin_direction = [];
    var barbData = [];
    for (var i = 0; i < data.length; i++) {
        bin_time.push(data[i][0]);
        bin_speed.push(data[i][1]);
        bin_direction.push(data[i][2]);

        if (i+1 == data.length || data[i+1][0] >= t) {
            // compute averages
            var time_avg = bin_time.reduce(function(a, b) { return a + b;}) / bin_time.length;
            var speed_avg = bin_speed.reduce(function(a, b) { return a + b;}) / bin_speed.length;
            speed_avg = speed_avg * 1000 / 3600.0;
            var direction_avg = circularMean(bin_direction);
            barbData.push([time_avg, speed_avg, direction_avg]);
            t += binsize;
            bin_time = [];
            bin_speed = [];
            bin_direction = [];
        }
    }

    return barbData;
}

function syncExtremes(e) {
    var thisChart = this.chart;

    if (e.trigger !== 'syncExtremes') { // Prevent feedback loop
        Highcharts.each(Highcharts.charts, function (chart) {
            if (chart !== thisChart && chart.polar != true) {
                if (chart.xAxis[0].setExtremes) { // It is null while updating
                    chart.xAxis[0].setExtremes(e.min, e.max, undefined, false, { trigger: 'syncExtremes' });
                }
            }
        });
    }
}

$(document).ready(populateCurrentData);

