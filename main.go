/*
 *
 * cmdline utility to configure and control the MHS-5200A series for function generators
 *
 * Copyright (c) 2020 - 2021 Peter Skarpetis
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 * SOFTWARE.
 *
 */

package main

import (
	"flag"
	"fmt"
	"github.com/peterska/go-utils"
	"os"
	"strconv"
	"time"
)

func main() {
	var verbose = flag.Int("v", 0, "verbose level")
	var debug = flag.Int("debug", 0, "debug level, 0=production, >0 is devmode")
	var pprof = flag.Bool("pprof", false, "enable golang profling")
	var port = flag.String("port", "/dev/ttyUSB0", "port the MHS-5200A is connected to")
	var scriptfile = flag.String("script", "", "json script file")
	flag.Parse()

	goutils.SetDebuglevel(*debug)
	goutils.SetLoglevel(*verbose)
	goutils.SetProfiling(*pprof)

	if len(*scriptfile) > 0 {
		err := playbackScript(*scriptfile, *port)
		if err != nil {
			goutils.Log.Print(err)
			os.Exit(10)
		}
	}
	if len(flag.Args()) == 0 { // nothing to do
		return
	}
	mhs5200, err := NewMHS5200A(*port)
	if err != nil {
		goutils.Log.Print(err)
		os.Exit(10)
	}
	if mhs5200 == nil {
		return
	}
	defer mhs5200.Close()
	channel := uint(1)
	needparam := false
	cmd := ""
	param := ""
	for _, argv := range flag.Args() {
		if needparam {
			param = argv
			needparam = false
		} else {
			cmd = argv
			param = ""
		}
		switch cmd {
		case "channel":
			if len(param) == 0 {
				needparam = true
				continue
			}
			v, err := strconv.ParseUint(param, 10, 32)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
				break
			}
			channel = uint(v)
			err = mhs5200.SelectChannel(channel)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
				break
			}

		case "sleep":
			if len(param) == 0 {
				needparam = true
				continue
			}
			v, err := strconv.ParseUint(param, 10, 32)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
				break
			}
			fmt.Printf("Sleeping for %v seconds\n", v)
			time.Sleep(time.Duration(v) * time.Second)

		case "delay":
			if len(param) == 0 {
				needparam = true
				continue
			}
			v, err := strconv.ParseUint(param, 10, 32)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
				break
			}
			fmt.Printf("Sleeping for %v seconds\n", v)
			time.Sleep(time.Duration(v) * time.Second)

		case "showconfig":
			var err error
			err = mhs5200.ShowChannelConfig(channel)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
			}

		case "showsweep":
			err = mhs5200.ShowSweepConfig()
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
			}

		case "frequency":
			if len(param) == 0 {
				needparam = true
				continue
			}
			v, err := strconv.ParseFloat(param, 64)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
				break
			}
			err = mhs5200.SetFrequency(channel, v)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
				break
			}

		case "waveform":
			if len(param) == 0 {
				needparam = true
				continue
			}
			err = mhs5200.SetWaveformFromString(channel, param)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
			}

		case "amplitude":
			if len(param) == 0 {
				needparam = true
				continue
			}
			v, err := strconv.ParseFloat(param, 64)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
				break
			}
			err = mhs5200.SetAmplitude(channel, v)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
			}

		case "duty":
			if len(param) == 0 {
				needparam = true
				continue
			}
			v, err := strconv.ParseFloat(param, 64)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
				break
			}
			err = mhs5200.SetDutyCycle(channel, v)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
			}

		case "offset":
			if len(param) == 0 {
				needparam = true
				continue
			}
			v, err := strconv.ParseFloat(param, 64)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
				break
			}
			err = mhs5200.SetOffset(channel, int(v))
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
			}

		case "phase":
			if len(param) == 0 {
				needparam = true
				continue
			}
			v, err := strconv.ParseFloat(param, 64)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
				break
			}
			err = mhs5200.SetPhase(channel, uint(v))
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
			}

		case "attenuation":
			if len(param) == 0 {
				needparam = true
				continue
			}
			if param == "on" {
				err = mhs5200.SetAttenuation(channel, ATTENUATION_MINUS_20DB)
			} else if param == "off" {
				err = mhs5200.SetAttenuation(channel, ATTENUATION_0DB)
			} else {
				err = fmt.Errorf("Unknown parameter %v", param)
			}
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
			}

		case "sweepstart":
			if len(param) == 0 {
				needparam = true
				continue
			}
			v, err := strconv.ParseFloat(param, 64)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
				break
			}
			err = mhs5200.SetSweepStart(v)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
				break
			}

		case "sweepend":
			if len(param) == 0 {
				needparam = true
				continue
			}
			v, err := strconv.ParseFloat(param, 64)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
				break
			}
			err = mhs5200.SetSweepEnd(v)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
				break
			}

		case "sweepduration":
			if len(param) == 0 {
				needparam = true
				continue
			}
			v, err := strconv.ParseUint(param, 10, 64)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
				break
			}
			err = mhs5200.SetSweepDuration(uint(v))
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
				break
			}

		case "sweeptype":
			if len(param) == 0 {
				needparam = true
				continue
			}
			err = mhs5200.SetSweepType(mhs5200.SweepTypeStringToInt(param))
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
				break
			}

		case "on":
			err = mhs5200.SetOnOff(true)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
			}

		case "off":
			err = mhs5200.SetOnOff(false)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
			}

		case "save":
			if len(param) == 0 {
				needparam = true
				continue
			}
			v, err := strconv.ParseUint(param, 10, 32)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
				break
			}
			err = mhs5200.Save(uint(v))
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
			}

		case "load":
			if len(param) == 0 {
				needparam = true
				continue
			}
			v, err := strconv.ParseUint(param, 10, 32)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
				break
			}
			err = mhs5200.Load(uint(v))
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
			}

		default:
			goutils.Log.Printf("%v, %v\n", goutils.Funcname(), fmt.Errorf("Unknown command %v", cmd))
		}
	}
	if err != nil {
		os.Exit(10)
	}
	if needparam {
		goutils.Log.Printf("Not enough parameters for %v command\n", cmd)
	}
}
