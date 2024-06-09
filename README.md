# LeafMQ MQTT broker
LeafMQ is a fast, lightweight and MQTT compliant broker / server designed for many different IoT tasks. As of now the server fully supports MQTT version 3.1.1.

## Getting started
These instructions will get you a copy of the project up and running on your local machine.

### Prerequisites
If you wish to manualy build the project you are going to need [golang]. It is also recommended to use [make] for easier compiling.

### Installing and running
if using make
```
git clone https://github.com/lawnp/LeafMQ.git
cd LeafMQ
make
./bin/mqtt-broker
```

After executing this, the broker should be running on your machine.

[golang]: https://go.dev/
[make]: https://www.gnu.org/software/make/
