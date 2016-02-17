gowx
====

Go Weather Station Interface

[![Build Status](https://travis-ci.org/geoffholden/gowx.svg?branch=master)](https://travis-ci.org/geoffholden/gowx)

A first attempt at a data logger and web interface for weather station data.

The current goal is to be able to parse the serial data from the [WSDL WxShield for Arduino](http://www.osengr.org/WxShield/Web/WxShield.html), store the parsed data in a database, and have a web display interface (Using [Highcharts](http://http://www.highcharts.com/)).

The next goal after that is to add the ability to upload the data to Weather Underground.

Installation
------------

`go get github.com/geoffholden/gowx`
