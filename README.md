Remote control of the MHS-5200A function generator
===========

[![Go Report Card](https://goreportcard.com/badge/github.com/peterska/go-mhs5200a)](https://goreportcard.com/report/github.com/peterska/go-mhs5200a)
[![BSD 3-Clause License](http://img.shields.io/badge/bsd-3-clause.svg)](./LICENSE)

About
-----

The [MHINSTEK_MHS-5200A](https://sigrok.org/wiki/MHINSTEK_MHS-5200A "MHINSTEK_MHS-5200A on sigrok") is a cheap, decent, dual channel function generator available from various Chinese vendors. The user interface is very clunky, virtually unusable, so I wrote this command line utility to configure and use the unit without using the controls. Most parameters and functions can be controlled by either command line parameters or a JSON file that stores a sequence of actions to perform.


Installation
-----------

[Install golang](https://golang.org/doc/install)

```
git clone https://github.com/peterska/go-mhs5200a.git
cd go-mhs5200a
make
```

The compiled binary can be found in the bin folder.

Quick start
----------

```
./bin/mhs5200a
````
will display some brief usage instructions.

Contact
-------

Please use [Github issue tracker](https://github.com/peterska/go-mhs5200a/issues) for filing bugs or feature requests.

License
-------

go-mhs5200a is licensed under the BSD 3-Clause License
