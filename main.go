/*
 * cmdline utility to configure and control the MHS-5200A series for function generators
 *
 * BSD 3-Clause License
 *
 * Copyright (c) 2020 - 2021, Peter Skarpetis
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 * 1. Redistributions of source code must retain the above copyright notice, this
 *    list of conditions and the following disclaimer.
 *
 * 2. Redistributions in binary form must reproduce the above copyright notice,
 *    this list of conditions and the following disclaimer in the documentation
 *    and/or other materials provided with the distribution.
 *
 * 3. Neither the name of the copyright holder nor the names of its
 *    contributors may be used to endorse or promote products derived from
 *    this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY THE COPYRIGHT HOLDERS AND CONTRIBUTORS "AS IS"
 * AND ANY EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE
 * IMPLIED WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL THE COPYRIGHT HOLDER OR CONTRIBUTORS BE LIABLE
 * FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL
 * DAMAGES (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR
 * SERVICES; LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER
 * CAUSED AND ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY,
 * OR TORT (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE
 * OF THIS SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 *
 */

package main

import (
	"flag"
	"fmt"
	"github.com/peterska/go-utils"
	"os"
	"path"
	"strconv"
	"time"
)

func usage() {
	fmt.Printf("Usage: %v [options] [command]...\n", path.Base(os.Args[0]))

	fmt.Printf("\noptions can be zero or more of the following:\n")
	flag.PrintDefaults()
	fmt.Printf("\n")

	fmt.Printf("command can be one or more of the following:\n")
	fmt.Printf("  showconfig - show the configuration of the current channel. Use channel to 0 to show config for all channels\n")
	fmt.Printf("  on - turn output on\n")
	fmt.Printf("  off - turn output off\n")
	fmt.Printf("\n")

	fmt.Printf("  channel [1|2] - sets the channel number commands will apply to\n")
	fmt.Printf("  frequency N - set the frequency N Hz\n")
	fmt.Printf("  waveform name - set the waveform to name. Valid names are sine, square, triangle, rising sawtooth, descending sawtooth\n")
	fmt.Printf("  amplitude N - set the amplitude to N Volts\n")
	fmt.Printf("  duty N - set the duty cycle to N%%\n")
	fmt.Printf("  offset N - set the DC offset to N Volts. Valid range is -120%% to +120%% of the configured amplitude\n")
	fmt.Printf("  phase N - set the phase to N°\n")
	fmt.Printf("  attenuation [on|off] - configure -20dB channel attenuation\n")
	fmt.Printf("\n")
	
	fmt.Printf("  showsweep - show the current sweep mode configuration\n")
	fmt.Printf("  sweepstart N - set the sweep start frequenecy to N Hz\n")
	fmt.Printf("  sweepend N - set the sweep end frequenecy to N Hz\n")
	fmt.Printf("  sweepduration N - set the sweep duration to N secs\n")
	fmt.Printf("  sweeptype [log|linear] - set the sweep type to either log or linear\n")
	fmt.Printf("  sweepon - turn sweep function on\n")
	fmt.Printf("  sweepoff - turn sweep function off\n")
	fmt.Printf("\n")
	
	fmt.Printf("  measure cmd - measure values from waveform on ext-input. cmd can be one of frequency, count, period, pulsewidth, duty, negativepulsewidth, stop\n")
	fmt.Printf("\n")
	
	fmt.Printf("  sleep N - delay N seconds before executing the next command\n")
	fmt.Printf("  delay N - delay N seconds before executing the next command\n")
	fmt.Printf("\n")
	
	fmt.Printf("  save N - save current configuration to slot N\n")
	fmt.Printf("  load N - load current configuration from slot N\n")
	fmt.Printf("\n")
	
	fmt.Printf("Examples:\n")
	fmt.Printf("%v channel 2 frequency 10000 phase 180 waveform square duty 33.25 attenuation off showconfig on sleep 120 off\n", path.Base(os.Args[0]))
	fmt.Printf("%v sweepstart 10 sweepend 100000 sweepduration 60 sweetype linear showsweep sweepon delay 60 sweepoff\n", path.Base(os.Args[0]))
	fmt.Printf("%v frequency 15.503 waveform square duty 50.0 on measure frequency sleep 10 measure stop off\n", path.Base(os.Args[0]))
	fmt.Printf("%v save 10\n", path.Base(os.Args[0]))
	fmt.Printf("%v load 10\n", path.Base(os.Args[0]))
}

func main() {
	var verbose = flag.Int("v", 0, "verbose level")
	//var debug = flag.Int("debug", 0, "debug level, 0=production, >0 is devmode")
	//var pprof = flag.Bool("pprof", false, "enable golang profling")
	var port = flag.String("port", "/dev/ttyUSB0", "port the MHS-5200A is connected to")
	var scriptfile = flag.String("script", "", "json script file")
	flag.Parse()

	//goutils.SetDebuglevel(*debug)
	//goutils.SetProfiling(*pprof)
	goutils.SetLoglevel(*verbose)

	if len(*scriptfile) > 0 {
		err := playbackScript(*scriptfile, *port)
		if err != nil {
			goutils.Log.Print(err)
			os.Exit(10)
		}
	}
	if len(flag.Args()) == 0 { // nothing to do
		usage()
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
			if channel == 0 {
				err = mhs5200.ShowConfig()
			} else {
				err = mhs5200.ShowChannelConfig(channel)
			}
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
			err = mhs5200.SetOffset(channel, v)
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

		case "sweepon":
			err = mhs5200.SetSweepState(true)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
			}

		case "sweepoff":
			err = mhs5200.SetSweepState(false)
			if err != nil {
				goutils.Log.Printf("%v, %v\n", goutils.Funcname(), err)
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

		case "measure":
			if len(param) == 0 {
				needparam = true
				continue
			}
			err = mhs5200.Measure(param)
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
