Remote control of the MHINSTEK MHS-5200A function generator
===========

[![BSD 3-Clause License](http://img.shields.io/badge/bsd-3-clause.svg)](./LICENSE)

About
-----

The [MHINSTEK MHS-5200A](https://sigrok.org/wiki/MHINSTEK_MHS-5200A "MHINSTEK MHS-5200A on sigrok") is a cheap, decent, dual channel function generator available from various Chinese vendors. The user interface is very clunky, virtually unusable, so I wrote this command line utility to configure and use the unit without using the controls. Most parameters and functions can be controlled by either command line parameters or a JSON file that stores a sequence of actions to perform.


Installation
------------

[Install golang](https://golang.org/doc/install)

```
git clone https://github.com/peterska/go-mhs5200a.git
cd go-mhs5200a
make
```

The compiled binary can be found in the bin folder.

Quick start
-----------

```
./bin/mhs5200a

Usage: mhs5200a [options] [command]...

options can be zero or more of the following:
  -port string
    	port the MHS-5200A is connected to (default "/dev/ttyUSB0")
  -script string
    	json script file
  -v int
    	verbose level

command can be one or more of the following:
  showconfig - show the configuration of the current channel
  on - turn output on
  off - turn output off

  channel [1|2] - sets the channel number commands will apply to
  frequency N - set the frequency N Hz
  waveform name - set the waveform to name. Valid names are sine, square, triangle, rising sawtooth, descending sawtooth
  amplitude N - set the amplitude to N Volts
  duty N - set the duty cycle to N%
  offset N - set the DC offset to N Volts. Valid range is -120% to +120% of the configured amplitude
  phase N - set the phase to N°
  attenuation [on|off] - configure -20dB channel attenuation

  showsweep - show the current sweep mode configuration
  sweepstart N - set the sweep start frequenecy to N Hz
  sweepend N - set the sweep end frequenecy to N Hz
  sweepduration N - set the sweep duration to N secs
  sweeptype [log|linear] - set the sweep type to either log or linear
  sweepon - turn sweep function on
  sweepoff - turn sweep function off

  measure cmd - measure values from waveform on ext-input. cmd can be one of frequency, count, period, pulsewidth, duty, negativepulsewidth, stop

  sleep N - delay N seconds before executing the next command
  delay N - delay N seconds before executing the next command

  save N - save current configuration to slot N
  load N - load current configuration from slot N

Examples:
mhs5200a channel 2 frequency 10000 phase 180 waveform square duty 33.25 attenuation off showconfig on sleep 120 off
mhs5200a sweepstart 10 sweepend 100000 sweepduration 60 sweetype linear showsweep sweepon delay 60 sweepoff
mhs5200a frequency 15.503 waveform square duty 50.0 on measure frequency sleep 10 measure stop off
mhs5200a save 10
mhs5200a load 10
````

Scripting
---------

mhs5200a supports rudimentary scripting by leveraging the JSON format. This allows storing frequently used setups on the host computer and not relying on the instrument's internal save/load slots.

In essence the json format encapsulates a sequence of commands into a JSON array called cmds and plays them back in sequence. In turn each cmd block contains a data array which encapsulates the configuration for that command.
A list of available commands is show below:
````
config
showconfig
delay
sleep
on
off
save
load
showsweep
configsweep
sweepon
sweepoff
````
A list of available parameters that can be specified in the data array are show below:
````
channel
frequency
waveform
amplitude
phase
duty
offset
attenuation
seconds
slot
startf
endf
type
````
A list of supported values for the waveform parameter are shown below:
````
sine
square
triangle
rising sawtooth
descending sawtooth
````
Let's look at a brief example
````JSON
{
    "cmds" : [
        { "cmd" : "config", "data" : [ { "channel" : 1, "frequency" : 1.0e03, "waveform" : "square", "amplitude" : 5.0, "phase" : 0.0, "duty" : 50.0, "attenuation" : false } ] },
        { "cmd" : "showconfig", "data" : [ { "channel" : 1 } ] }
    ]
}
````
The first line, cmd config, configures channel 1 for a frequency of 1KHz, a square waveform, an amplitude for 5.0V, a phase angle of 0 degrees, a 50% duty cycle and turns off the -20dB attenuation.
The second line simply dumps the configuration for channel 1.

Here is an example that show how to use the measurement fucntions
````JSON
{
    "cmds" : [
        { "cmd" : "config", "data" : [ { "channel" : 1, "frequency" : 25e03, "waveform" : "square", "amplitude" : 5.0, "phase" : 180.0, "duty" : 55.0, "attenuation" : false } ] },
        { "cmd" : "on" },
        
        { "cmd" : "measure", "data" : [ { "type" : "frequency" } ] },
        { "cmd" : "delay", "data" : [ { "seconds" : 5 } ] },
        
        { "cmd" : "measure", "data" : [ { "type" : "period" } ] },
        { "cmd" : "delay", "data" : [ { "seconds" : 5 } ] },
        
        { "cmd" : "measure", "data" : [ { "type" : "duty" } ] },
        { "cmd" : "delay", "data" : [ { "seconds" : 5 } ] },
        
        { "cmd" : "measure", "data" : [ { "type" : "off" } ] },
        { "cmd" : "off" }
    ]
}
````

Here is another more complicated example, that shows most of the commands available
````JSON
{
    "cmds" : [
        { "cmd" : "config", "data" : [ { "channel" : 1, "frequency" : 25e03, "waveform" : "square", "amplitude" : 5.0, "phase" : 90.0, "duty" : 33.3, "attenuation" : false } ] },
        { "cmd" : "on", "data" : [ { "channel" : 1 } ] },
        { "cmd" : "showconfig", "data" : [ { "channel" : 1 } ] },
        { "cmd" : "delay", "data" : [ { "seconds" : 15 } ] },
        
        { "cmd" : "config", "data" : [ { "channel" : 1, "frequency" : 50e03, "waveform" : "sine", "amplitude" : 3.3, "phase" : 180.0, "duty" : 50.0, "attenuation" : false } ] },
        { "cmd" : "showconfig", "data" : [ { "channel" : 1 } ] },
        { "cmd" : "delay", "data" : [ { "seconds" : 15 } ] },
        
        { "cmd" : "config", "data" : [ { "channel" : 1, "frequency" : 55e03, "waveform" : "triangle", "amplitude" : 5.0, "phase" : 0.0, "duty" : 75.0, "attenuation" : false } ] },
        { "cmd" : "showconfig", "data" : [ { "channel" : 1 } ] },
        { "cmd" : "delay", "data" : [ { "seconds" : 15 } ] },
        
        { "cmd" : "config", "data" : [ { "channel" : 1, "frequency" : 60e03, "waveform" : "rising sawtooth", "amplitude" : 5.0, "phase" : 0.0, "duty" : 50.0, "attenuation" : false } ] },
        { "cmd" : "showconfig", "data" : [ { "channel" : 1 } ] },
        { "cmd" : "delay", "data" : [ { "seconds" : 15 } ] },
        
        { "cmd" : "config", "data" : [ { "channel" : 1, "frequency" : 65e03, "waveform" : "descending sawtooth", "amplitude" : 5.0, "phase" : 0.0, "duty" : 50.0, "attenuation" : false } ] },
        { "cmd" : "showconfig", "data" : [ { "channel" : 1 } ] },
        { "cmd" : "delay", "data" : [ { "seconds" : 15 } ] },
        
        { "cmd" : "configsweep", "data" : [ { "startf" : 1e03, "endf" : 100e03, "seconds" : 10, "type" : "log", "waveform": "square", "duty" : 50.0 } ] },
        { "cmd" : "showsweep" },
        { "cmd" : "sweepon" },
        { "cmd" : "delay", "data" : [ { "seconds" : 10 } ] },
        { "cmd" : "sweepoff" },
        
        { "cmd" : "off", "data" : [ { "channel" : 1 } ] }
    ]
}
````


Contact
-------

Please use [Github issue tracker](https://github.com/peterska/go-mhs5200a/issues) for filing bugs or feature requests.

License
-------

go-mhs5200a is licensed under the BSD 3-Clause License
