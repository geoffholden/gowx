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
        $('#current_wind_angle').css("transform", "rotate(" + (data['WindDir'] - 90) + "deg)");
        $('#current_rain').html(data['RainRate'].toFixed(1));
    });
    $.getJSON("/change.json?type=Pressure&time=3h", function(data) {
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
    $.getJSON("/change.json?type=RainTotal&time=24h", function(data) {
        $('#rain_total').html(data.Change[0].toFixed(1));
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
    var title;
    if (name instanceof Array) {
        title = name[0];
    } else {
        title = name;
    }
    var options = {
        chart: {
            renderTo: container,
            zoomType: 'x',
        },
        title: {
            text: title,
        },
        xAxis: {
            type: 'datetime',
        },
        yAxis: {
            title: {
                text: title + " (" + units + ")",
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
            if (name instanceof Array) {
                title = name[i+1];
            } else {
                title = name;
            }
            options.series.push({
                type: 'spline',
                data: data.Data[i],
                name: title,
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

$(document).ready(populateCurrentData);

