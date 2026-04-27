# Teknikfag Gartneri Backend CLI



## Needed APIs
- [x] POST Temperature intake API (/api/post/temperature)
- [x] POST Humidity intake API (/api/post/humidity)
- [x] POST Temperature and Humidity intake API (/api/post/temperaturehumidity) (Json example {"Temperature": 25.5, "Humidity" : 22.5, "Timestamp": "2026-04-21T12:34:56Z", "DeviceId": "device-12345"})


### GET APIs
- [x] GET get temperatures in range (/api/get/temperature?interval=5m) (interval can get minutues (5m), hours (1h) or days (1d) you can also get alltime by using interval=0)
- [x] GET get humidity in range (/api/get/humidity?interval=5m) (interval can get minutues (5m), hours (1h) or days (1d) you can also get alltime by using interval=0)
- [x] GET Greeting API (/greet)

