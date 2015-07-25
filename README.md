# gowx
Go Weather Station Interface

A first attempt at a data logger and web interface for weather station data.

The current goal is to be able to parse the serial data from the
[WSDL WxShield for
Arduino](http://www.osengr.org/WxShield/Web/WxShield.html), store
the parsed data in a database (MongoDB, maybe?), and have a web
display interface (Using [Chart.js](http://www.chartjs.org/)).

The next goal after that is to add the ability to upload the data to Weather
Underground.
