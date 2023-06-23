# NixMQ MQTT broker
NixMQ is an MQTT 3.1.1 broker developed as part of my final year project dissertation for my bachelor's degree at the Faculty of Computer and Information Science, University of Ljubljana.

## Getting started
These instructions will get you a copy of the project up and running on your local machine for testing purposes.

### Prerequisites
For running this project you will need [golang] version 1.20. It is also recommended to use [make] for easier compiling.

### Installing and running
if using make
```
git clone https://github.com/LanPavletic/NixMQ.git
cd NixMQ
make
./bin/mqtt-broker
```

After executing this, the broker should be running on your machine.

### testing with clients
For testing the broker I recommend using [MQTTX] application. It is made for testing MQTT brokers with powerfull and well designed GUI.

[goang]: https://go.dev/
[make]: https://www.gnu.org/software/make/
[MQTTX]: https://mqttx.app/
