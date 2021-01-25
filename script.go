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
	"encoding/json"
	"fmt"
	"github.com/peterska/go-utils"
	"io/ioutil"
	"math"
	"time"
)

const (
	defaultPollingIntervalSeconds = 60
)

type CMDPARAMS struct {
	Channel     *uint    `json:"channel,omitempty"`
	Frequency   *float64 `json:"frequency,omitempty"`
	Waveform    *string  `json:"waveform,omitempty"`
	Amplitude   *float64 `json:"amplitude,omitempty"`
	Phase       *float64 `json:"phase,omitempty"`
	Duty        *float64 `json:"duty,omitempty"`
	Offset      *float64 `json:"offset,omitempty"`
	Attenuation *bool    `json:"attenuation,omitempty"`
	Seconds     *uint    `json:"seconds,omitempty"`
	Slot        *uint    `json:"slot,omitempty"`
	Startf      *float64 `json:"startf,omitempty"`
	Endf        *float64 `json:"endf,omitempty"`
	Type        *string  `json:"type,omitempty"`
}

type CMD struct {
	Cmd  string      `json:"cmd,omitempty"`
	Data []CMDPARAMS `json:"data,omitempty"`
}

type SCRIPT struct {
	Port string `json:"port,omitempty"`
	Cmds []CMD  `json:"cmds,omitempty"`
}

func (params *CMDPARAMS) convertToChannelVals(mhs5200 *MHS5200A) *CHANNELVALS {
	v := CHANNELVALS{
		Channel:     math.MaxUint32,
		Frequency:   math.NaN(),
		Waveform:    "",
		Amplitude:   math.NaN(),
		Phase:       math.NaN(),
		Duty:        math.NaN(),
		Offset:      math.NaN(),
		Attenuation: math.MaxUint32,
	}
	if params.Channel != nil {
		v.Channel = *params.Channel
	}
	if params.Frequency != nil {
		v.Frequency = *params.Frequency
	}
	if params.Waveform != nil {
		v.Waveform = *params.Waveform
	}
	if params.Amplitude != nil {
		v.Amplitude = *params.Amplitude
	}
	if params.Phase != nil {
		v.Phase = *params.Phase
	}
	if params.Duty != nil {
		v.Duty = *params.Duty
	}
	if params.Offset != nil {
		v.Offset = *params.Offset
	}
	if params.Attenuation != nil {
		if *params.Attenuation {
			v.Attenuation = ATTENUATION_MINUS_20DB
		} else {
			v.Attenuation = ATTENUATION_0DB
		}
	}
	if goutils.Loglevel() > 0 {
		goutils.Log.Printf("%v: %+v", goutils.Funcname(), v)
	}
	return &v
}

func (params *CMDPARAMS) convertToSweepVals(mhs5200 *MHS5200A) *SWEEPVALS {
	v := SWEEPVALS{
		Startf:   math.NaN(),
		Endf:     math.NaN(),
		Duration: math.MaxUint32,
		Type:     math.MaxUint32,
		Waveform: "",
		Duty:     math.NaN(),
	}
	if params.Startf != nil {
		v.Startf = *params.Startf
	}
	if params.Endf != nil {
		v.Endf = *params.Endf
	}
	if params.Seconds != nil {
		v.Duration = *params.Seconds
	}
	if params.Type != nil {
		v.Type = mhs5200.SweepTypeStringToInt(*params.Type)
	}
	if params.Waveform != nil {
		v.Waveform = *params.Waveform
	}
	if params.Duty != nil {
		v.Duty = *params.Duty
	}
	if goutils.Loglevel() > 0 {
		goutils.Log.Printf("%v: %+v", goutils.Funcname(), v)
	}
	return &v
}

func script(scriptfile string, port string) (*SCRIPT, error) {
	if len(scriptfile) == 0 {
		return nil, fmt.Errorf("Cannot find configuration file")
	}
	jsn, err := ioutil.ReadFile(scriptfile)
	if err != nil {
		goutils.Log.Printf("%v", err)
		return nil, err
	}
	script := SCRIPT{
		Port: port,
	}
	err = json.Unmarshal(jsn, &script)
	if err != nil {
		goutils.Log.Printf("%v", err)
		return nil, err
	}
	if goutils.Loglevel() > 1 {
		goutils.Log.Printf("Loaded config from %v, %+v", scriptfile, script)
	}
	return &script, nil
}

func timestampString() string {
	return time.Now().Format(time.Stamp)
}

func playbackScript(scriptfile string, port string) error {
	script, err := script(scriptfile, port)
	if err != nil {
		return err
	}
	if len(script.Port) == 0 {
		goutils.Log.Printf("%v", fmt.Errorf("Port was not specified"))
		return fmt.Errorf("Port was not specified")
	}
	mhs5200, err := NewMHS5200A(script.Port)
	if err != nil {
		return err
	}
	defer mhs5200.Close()
	for _, cmd := range script.Cmds {
		switch cmd.Cmd {
		case "config":
			for _, data := range cmd.Data {
				fmt.Printf("%v: Configuring channel %v\n", timestampString(), *data.Channel)
				err = mhs5200.ApplyChannelConfig(data.convertToChannelVals(mhs5200))
				if err != nil {
					return err
				}
			}

		case "showconfig":
			if cmd.Data != nil {
				for _, data := range cmd.Data {
					if data.Channel != nil {
						err = mhs5200.ShowChannelConfig(*data.Channel)
					} else {
						err = mhs5200.ShowConfig()
					}
					if err != nil {
						return err
					}
				}
			} else {
				err = mhs5200.ShowConfig()
				if err != nil {
					return err
				}
			}

		case "delay":
			if cmd.Data != nil {
				for _, data := range cmd.Data {
					if data.Seconds != nil {
						fmt.Printf("%v: Sleeping %v seconds\n", timestampString(), *data.Seconds)
						time.Sleep(time.Second * time.Duration(*data.Seconds))
					}
				}
			} else {
				time.Sleep(time.Second * 1)
			}

		case "sleep":
			if cmd.Data != nil {
				for _, data := range cmd.Data {
					if data.Seconds != nil {
						fmt.Printf("%v: Sleeping %v seconds\n", timestampString(), *data.Seconds)
						time.Sleep(time.Second * time.Duration(*data.Seconds))
					}
				}
			} else {
				time.Sleep(time.Second * 1)
			}

		case "on":
			fmt.Printf("%v: Output on\n", timestampString())
			err = mhs5200.SetOnOff(true)
			if err != nil {
				return err
			}

		case "off":
			fmt.Printf("%v: Output off\n", timestampString())
			err = mhs5200.SetOnOff(false)
			if err != nil {
				return err
			}

		case "save":
			if cmd.Data != nil {
				for _, data := range cmd.Data {
					if data.Slot != nil {
						fmt.Printf("%v: Saving to slot %v\n", timestampString(), *data.Slot)
						err = mhs5200.Save(*data.Slot)
					} else {
						err = mhs5200.Save(0)
					}
					if err != nil {
						return err
					}
				}
			} else {
				err = mhs5200.Save(0)
				if err != nil {
					return err
				}
			}

		case "load":
			if cmd.Data != nil {
				for _, data := range cmd.Data {
					if data.Slot != nil {
						fmt.Printf("%v: Loading from slot %v\n", timestampString(), *data.Slot)
						err = mhs5200.Load(*data.Slot)
					} else {
						err = mhs5200.Load(0)
					}
					if err != nil {
						return err
					}
				}
			} else {
				err = mhs5200.Load(0)
				if err != nil {
					return err
				}
			}

		case "showsweep":
			err = mhs5200.ShowSweepConfig()
			if err != nil {
				return err
			}

		case "configsweep":
			// sweeps are only valid on channel 1
			for _, data := range cmd.Data {
				fmt.Printf("%v: Configuring sweep\n", timestampString())
				err = mhs5200.SetSweep(data.convertToSweepVals(mhs5200))
				if err != nil {
					return err
				}
			}

		case "sweepon":
			fmt.Printf("%v: Sweep on\n", timestampString())
			err = mhs5200.SetSweepState(true)
			if err != nil {
				return err
			}

		case "sweepoff":
			fmt.Printf("%v: Sweep off\n", timestampString())
			err = mhs5200.SetSweepState(false)
			if err != nil {
				return err
			}

		default:
			return fmt.Errorf("Unknown command %s", cmd.Cmd)
		}
	}
	return nil
}
