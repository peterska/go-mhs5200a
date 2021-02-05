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
	"bufio"
	"fmt"
	"github.com/peterska/go-utils"
	"github.com/tarm/serial"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	ATTENUATION_MINUS_20DB = 0
	ATTENUATION_0DB        = 1

	SWEEP_LINEAR = 0
	SWEEP_LOG    = 1

	SWEEP_START = 0
	SWEEP_STOP  = 1

	MHS5200A_CMD_TIMEOUT = 500 * time.Millisecond
)

const (
	ARB_WAVEFORM_NUM_POINTS        = 2048
	ARB_WAVEFORM_MAX_AMPLITUDE     = 4095
	ARB_WAVEFORM_NUM_SLICES        = 16
	ARB_WAVEFORM_SAMPLES_PER_SLICE = 128

	// range of input values for arbitrary waveform definition
	ARB_WAVEFORM_INPUT_MIN  = -1.0
	ARB_WAVEFORM_INPUT_MAX  = 1.0
	ARB_WAVEFORM_OUTPUT_MIN = 0.0
	ARB_WAVEFORM_OUTPUT_MAX = 4095.0
)

const (
	WAVEFORM_SINE = iota
	WAVEFORM_SQUARE
	WAVEFORM_TRIANGLE
	WAVEFORM_RISING_SAWTOOTH
	WAVEFORM_DESCENDING_SAWTOOTH
	WAVEFORM_SINC
	WAVEFORM_NORM_SINC
)

const (
	WAVEFORM_ARB_0 = iota + 100
	WAVEFORM_ARB_1
	WAVEFORM_ARB_2
	WAVEFORM_ARB_3
	WAVEFORM_ARB_4
	WAVEFORM_ARB_5
	WAVEFORM_ARB_6
	WAVEFORM_ARB_7
	WAVEFORM_ARB_8
	WAVEFORM_ARB_9
	WAVEFORM_ARB_10
	WAVEFORM_ARB_11
	WAVEFORM_ARB_12
	WAVEFORM_ARB_13
	WAVEFORM_ARB_14
	WAVEFORM_ARB_15
)

const (
	WAVEFORM_ARB_ALT_0 = iota + 10
	WAVEFORM_ARB_ALT_1
	WAVEFORM_ARB_ALT_2
	WAVEFORM_ARB_ALT_3
	WAVEFORM_ARB_ALT_4
	WAVEFORM_ARB_ALT_5
	WAVEFORM_ARB_ALT_6
	WAVEFORM_ARB_ALT_7
	WAVEFORM_ARB_ALT_8
	WAVEFORM_ARB_ALT_9
	WAVEFORM_ARB_ALT_10
	WAVEFORM_ARB_ALT_11
	WAVEFORM_ARB_ALT_12
	WAVEFORM_ARB_ALT_13
	WAVEFORM_ARB_ALT_14
	WAVEFORM_ARB_ALT_15
)

const (
	WAVEFORM_SINE_STR                = "sine"
	WAVEFORM_SQUARE_STR              = "square"
	WAVEFORM_TRIANGLE_STR            = "triangle"
	WAVEFORM_RISING_SAWTOOTH_STR     = "rising sawtooth"
	WAVEFORM_DESCENDING_SAWTOOTH_STR = "descending sawtooth"
	WAVEFORM_SINC_STR                = "sinc"
	WAVEFORM_NORM_SINC_STR           = "normsinc"
	WAVEFORM_ARB_STR                 = "arbitrary"
	WAVEFORM_ARB_STR_SHORT           = "arb"
)

const (
	COUNTER_MEASURE_FREQUENCY = iota
	COUNTER_MEASURE_COUNT
	COUNTER_MEASURE_PERIOD
	COUNTER_MEASURE_PULSE_WIDTH
	COUNTER_MEASURE_NEGATIVE_PULSE_WIDTH
	COUNTER_MEASURE_DUTY_CYCLE
)

type SWEEPVALS struct {
	Startf   float64
	Endf     float64
	Duration uint
	Type     uint // log or linear
	Waveform string
	Duty     float64
}

type CHANNELVALS struct {
	Channel     uint    `json:"channel,omitempty"`
	Frequency   float64 `json:"frequency,omitempty"`
	Waveform    string  `json:"waveform,omitempty"`
	Amplitude   float64 `json:"amplitude,omitempty"`
	Phase       float64 `json:"phase,omitempty"`
	Duty        float64 `json:"duty,omitempty"`
	Offset      float64 `json:"offset,omitempty"`
	Attenuation uint    `json:"attenuation,omitempty"`
}

type MHS5200A struct {
	stream      *serial.Port
	quit        chan struct{}
	wg          sync.WaitGroup
	mutex       sync.Mutex
	port        string
	measure     bool // whether we are reading measurements from the instrument
	measuretype int  // type of measurement
}

// normalise values to the requested range
func normalise(data []float64, inmin float64, inmax float64, outmin float64, outmax float64) {
	fmt.Printf("# Normalise: %v, %v -> %v, %v\n", inmin, inmax, outmin, outmax) //XXXXXXXXXXXXX
	for i, _ := range data {
		data[i] = outmin + (data[i]-inmin)*(outmax-outmin)/(inmax-inmin)
	}
}

// normalise values to the requested range
func autoNormalise(data []float64, outmin float64, outmax float64) {
	minval := math.NaN()
	maxval := math.NaN()
	for _, v := range data {
		if math.IsNaN(minval) || v < minval {
			minval = v
		}
		if math.IsNaN(maxval) || v > maxval {
			maxval = v
		}
	}
	if minval >= 0 { // must be arbitrary waveform encoded with an offset
		minval = 0
		if maxval < 256 {
			maxval = 255
		} else if maxval < 512 {
			maxval = 511
		} else if maxval < 1024 {
			maxval = 1023
		} else if maxval < 2048 {
			maxval = 2047
		} else if maxval < 4096 {
			maxval = 4095
		} else if maxval < 8192 {
			maxval = 8191
		}
	}
	normalise(data, minval, maxval, outmin, outmax)
}

func normalisedSinc(x float64) float64 {
	x *= math.Pi
	if x != 0.0 {
		return math.Sin(x) / x
	}
	return 1.0
}

func generateNormalisedSinc() []float64 {
	waveform := make([]float64, ARB_WAVEFORM_NUM_POINTS)
	xstart := -4.0 * math.Pi
	xend := 3.0 * math.Pi
	xstep := (xend - xstart) / float64(ARB_WAVEFORM_NUM_POINTS)
	x := xstart
	for i := 0; i < ARB_WAVEFORM_NUM_POINTS; i++ {
		waveform[i] = normalisedSinc(x)
		x += xstep
	}
	return waveform
}

func sinc(x float64) float64 {
	if x != 0.0 {
		return math.Sin(x) / x
	}
	return 1.0
}

func generateSinc() []float64 {
	waveform := make([]float64, ARB_WAVEFORM_NUM_POINTS)
	xstart := -6.0 * math.Pi
	xend := 5.0 * math.Pi
	xstep := (xend - xstart) / float64(ARB_WAVEFORM_NUM_POINTS)
	x := xstart
	for i := 0; i < ARB_WAVEFORM_NUM_POINTS; i++ {
		waveform[i] = sinc(x)
		x += xstep
	}
	return waveform
}

func SiUnitsPrefix(exponent int) string {
	switch exponent {
	case 0:
		return ""
	case 3:
		return "K"
	case 6:
		return "M"
	case 9:
		return "G"
	case 12:
		return "T"
	case 15:
		return "P"
	case 18:
		return "E"
	case 21:
		return "Z"
	case 24:
		return "Y"
	case -3:
		return "m"
	case -6:
		return "u"
	case -9:
		return "n"
	case -12:
		return "p"
	case -15:
		return "f"
	case -18:
		return "a"
	case -21:
		return "z"
	case -24:
		return "y"
	}
	return ""
}

func (mhs5200 *MHS5200A) MeasuretypeString(v int) string {
	switch v {
	case COUNTER_MEASURE_FREQUENCY:
		return "frequency"

	case COUNTER_MEASURE_COUNT:
		return "count"

	case COUNTER_MEASURE_PERIOD:
		return "period"

	case COUNTER_MEASURE_PULSE_WIDTH:
		return "pulse width"

	case COUNTER_MEASURE_NEGATIVE_PULSE_WIDTH:
		return "negative pulse width"

	case COUNTER_MEASURE_DUTY_CYCLE:
		return "duty cycle"
	}
	return "unknown"
}

func (mhs5200 *MHS5200A) OnOffString(v uint) string {
	if v == 0 {
		return "Off"
	} else {
		return "On"
	}
}

func (mhs5200 *MHS5200A) BooleanString(v bool) string {
	if v {
		return "True"
	} else {
		return "False"
	}
}

func (mhs5200 *MHS5200A) UnitsString(v float64, units string, engmode bool) string {
	if engmode {
		exponent := 0
		for math.Abs(v) >= 1.0e3 {
			exponent += 3
			v *= 1.0e-3
			if exponent > 9 {
				break
			}
		}
		for math.Abs(v) > 0.0 && math.Abs(v) < 1.0 {
			exponent -= 3
			v *= 1.0e3
			if exponent < -9 {
				break
			}
		}
		return fmt.Sprintf("%.3g %s%s", v, SiUnitsPrefix(exponent), units)
	} else {
		return fmt.Sprintf("%.3g %s", v, units)
	}
}

func (mhs5200 *MHS5200A) sendCommand(cmd []byte) ([]byte, error) {
	mhs5200.mutex.Lock()
	defer mhs5200.mutex.Unlock()
	if goutils.Loglevel() > 1 {
		goutils.Log.Printf("%v:\tsend:\t%s\n", goutils.Callername(), string(cmd))
	}
	_, err := mhs5200.stream.Write(append(cmd, '\n'))
	if err != nil {
		goutils.Log.Print(err)
		return nil, err
	}

	response := []byte{}
	start := time.Now()
	for time.Now().Sub(start) < MHS5200A_CMD_TIMEOUT {
		b, err := ioutil.ReadAll(mhs5200.stream)
		if err != nil {
			return nil, err
		}
		if len(b) > 0 {
			response = append(response, b...)
			if response[len(response)-1] == '\n' {
				break
			}
		}
	}
	s := []byte(strings.TrimRight(string(response), " \n\r"))
	if goutils.Loglevel() > 1 {
		goutils.Log.Printf("%v:\treceive: %s", goutils.Callername(), s)
	}
	return s, nil
}

func (mhs5200 *MHS5200A) sendCommandAndExpect(cmd []byte, expect string) error {
	data, err := mhs5200.sendCommand([]byte(cmd))
	if err != nil {
		return err
	}
	if string(data) != expect {
		return fmt.Errorf("Expected %v, got %v, %v instead", expect, string(data), data)
	}
	return nil
}

func (mhs5200 *MHS5200A) sendCommandAndExpectUint(cmd []byte) (uint, error) {
	data, err := mhs5200.sendCommand([]byte(cmd))
	if err != nil {
		return 0, err
	}
	if len(data) < 4 {
		return 0, fmt.Errorf("data underlow")
	}
	data = data[4:]
	v, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

func (mhs5200 *MHS5200A) GetCounterValue() (uint, error) {
	return mhs5200.sendCommandAndExpectUint([]byte(":r0e"))
}

func (mhs5200 *MHS5200A) GetFrequencyMeasurement() (float64, error) {
	u, err := mhs5200.GetCounterValue()
	if err != nil {
		return math.NaN(), err
	}
	return float64(u), nil
}

func (mhs5200 *MHS5200A) GetPeriodMeasurement() (float64, error) {
	u, err := mhs5200.GetCounterValue()
	if err != nil {
		return math.NaN(), err
	}
	return float64(u) * 1.0e-9, nil
}

func (mhs5200 *MHS5200A) GetDutyCycleMeasurement() (float64, error) {
	u, err := mhs5200.GetCounterValue()
	if err != nil {
		return math.NaN(), err
	}
	return float64(u) / 10.0, nil
}

func (mhs5200 *MHS5200A) GetMeasurement() (float64, error) {
	switch mhs5200.measuretype {
	case COUNTER_MEASURE_FREQUENCY:
		return mhs5200.GetFrequencyMeasurement()

	case COUNTER_MEASURE_COUNT:
		v, err := mhs5200.GetCounterValue()
		return float64(v), err

	case COUNTER_MEASURE_PERIOD:
		return mhs5200.GetPeriodMeasurement()

	case COUNTER_MEASURE_PULSE_WIDTH:
		return mhs5200.GetPeriodMeasurement()

	case COUNTER_MEASURE_NEGATIVE_PULSE_WIDTH:
		return mhs5200.GetPeriodMeasurement()

	case COUNTER_MEASURE_DUTY_CYCLE:
		return mhs5200.GetDutyCycleMeasurement()
	}
	return math.NaN(), fmt.Errorf("Unknown measurement type %v", mhs5200.measuretype)
}

func (mhs5200 *MHS5200A) GetMeasurementAsString() (string, error) {
	v, err := mhs5200.GetMeasurement()
	if err != nil {
		return "", err
	}
	switch mhs5200.measuretype {
	case COUNTER_MEASURE_FREQUENCY:
		return mhs5200.FrequencyString(v), nil

	case COUNTER_MEASURE_COUNT:
		return fmt.Sprintf("%v", uint(v)), nil

	case COUNTER_MEASURE_PERIOD:
		return mhs5200.UnitsString(v, "s", true), nil

	case COUNTER_MEASURE_PULSE_WIDTH:
		return mhs5200.UnitsString(v, "s", true), nil

	case COUNTER_MEASURE_NEGATIVE_PULSE_WIDTH:
		return mhs5200.UnitsString(v, "s", true), nil

	case COUNTER_MEASURE_DUTY_CYCLE:
		return mhs5200.DutyCycleString(v), nil
	}
	return "", fmt.Errorf("Unknown measurement type %v", mhs5200.measuretype)
}

func (mhs5200 *MHS5200A) Measure(cmd string) error {
	var err error = nil
	switch cmd {
	case "stop":
		mhs5200.measure = false
		err = mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s6b%d", 0)), "ok")

	case "frequency":
		mhs5200.measuretype = COUNTER_MEASURE_FREQUENCY
		err = mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s%dm", mhs5200.measuretype)), "ok")
		mhs5200.measure = true

	case "count":
		mhs5200.measuretype = COUNTER_MEASURE_COUNT
		err = mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s%dm", mhs5200.measuretype)), "ok")
		mhs5200.measure = true

	case "period":
		mhs5200.measuretype = COUNTER_MEASURE_PERIOD
		err = mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s%dm", mhs5200.measuretype)), "ok")
		mhs5200.measure = true

	case "pulsewidth":
		mhs5200.measuretype = COUNTER_MEASURE_PULSE_WIDTH
		err = mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s%dm", mhs5200.measuretype)), "ok")
		mhs5200.measure = true

	case "negativepulsewidth":
		mhs5200.measuretype = COUNTER_MEASURE_NEGATIVE_PULSE_WIDTH
		err = mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s%dm", mhs5200.measuretype)), "ok")
		mhs5200.measure = true

	case "duty":
		mhs5200.measuretype = COUNTER_MEASURE_DUTY_CYCLE
		err = mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s%dm", mhs5200.measuretype)), "ok")
		mhs5200.measure = true

	default:
		err = fmt.Errorf("unknown measure paramter %v", cmd)
	}
	return err
}

func (mhs5200 *MHS5200A) FrequencyString(v float64) string {
	return mhs5200.UnitsString(v, "Hz", true)
}

func (mhs5200 *MHS5200A) SetFrequency(ch uint, v float64) error {
	if math.IsNaN(v) {
		return nil
	}
	if v < 0.0 || v > 25.0e6 {
		return fmt.Errorf("%v is not a valid frequency", v)
	}
	return mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s%df%d", ch, int(v*100.0))), "ok")
}

func (mhs5200 *MHS5200A) GetFrequency(ch uint) (float64, error) {
	u, err := mhs5200.sendCommandAndExpectUint([]byte(fmt.Sprintf(":r%df", ch)))
	if err != nil {
		return 0.0, err
	}
	return float64(u) / 100.0, nil
}

func (mhs5200 *MHS5200A) WaveformString(v uint) string {
	switch v {
	case WAVEFORM_SINE:
		return WAVEFORM_SINE_STR

	case WAVEFORM_SQUARE:
		return WAVEFORM_SQUARE_STR

	case WAVEFORM_TRIANGLE:
		return WAVEFORM_TRIANGLE_STR

	case WAVEFORM_RISING_SAWTOOTH:
		return WAVEFORM_RISING_SAWTOOTH_STR

	case WAVEFORM_DESCENDING_SAWTOOTH:
		return WAVEFORM_DESCENDING_SAWTOOTH_STR

	case WAVEFORM_SINC:
		return WAVEFORM_SINC_STR

	case WAVEFORM_NORM_SINC:
		return WAVEFORM_NORM_SINC_STR

	}
	if v >= WAVEFORM_ARB_0 && v <= WAVEFORM_ARB_15 {
		return fmt.Sprintf("%s%d", WAVEFORM_ARB_STR, v-WAVEFORM_ARB_0)
	} else if v >= WAVEFORM_ARB_ALT_0 && v <= WAVEFORM_ARB_ALT_15 {
		return fmt.Sprintf("%s%d", WAVEFORM_ARB_STR, v-WAVEFORM_ARB_ALT_0)
	}
	return "unknown"
}

func (mhs5200 *MHS5200A) WaveformStringToInt(s string) uint {
	switch s {
	case WAVEFORM_SINE_STR:
		return WAVEFORM_SINE

	case WAVEFORM_SQUARE_STR:
		return WAVEFORM_SQUARE

	case WAVEFORM_TRIANGLE_STR:
		return WAVEFORM_TRIANGLE

	case WAVEFORM_RISING_SAWTOOTH_STR:
		return WAVEFORM_RISING_SAWTOOTH

	case WAVEFORM_DESCENDING_SAWTOOTH_STR:
		return WAVEFORM_DESCENDING_SAWTOOTH

	case WAVEFORM_SINC_STR:
		return WAVEFORM_SINC

	case WAVEFORM_NORM_SINC_STR:
		return WAVEFORM_NORM_SINC

	default:
		if strings.HasPrefix(s, WAVEFORM_ARB_STR) {
			s = strings.TrimPrefix(s, WAVEFORM_ARB_STR)
			u, err := strconv.ParseUint(s, 10, 32)
			if err != nil {
				goutils.Log.Printf("%s, %v", goutils.Funcname(), err)
			}
			return uint(u + WAVEFORM_ARB_0)
		} else if strings.HasPrefix(s, WAVEFORM_ARB_STR_SHORT) {
			s = strings.TrimPrefix(s, WAVEFORM_ARB_STR_SHORT)
			u, err := strconv.ParseUint(s, 10, 32)
			if err != nil {
				goutils.Log.Printf("%s, %v", goutils.Funcname(), err)
			}
			return uint(u + WAVEFORM_ARB_0)
		}
		return math.MaxUint32
	}
}

// NormalisedToArbitraryWaveform convert an amplitude in the range -1.0 - 1.0 to a value in the range of the arbitrary waveform
func (mhs5200 *MHS5200A) NormalisedToArbitraryWaveform(v float64) int {
	if v > 1.0 {
		goutils.Log.Printf("adjusting bad value %v", v)
		v = 1.0
	} else if v < -1.0 {
		goutils.Log.Printf("adjusting bad value %v", v)
		v = -1.0
	}
	v = (v - ARB_WAVEFORM_INPUT_MIN) * (ARB_WAVEFORM_OUTPUT_MAX - ARB_WAVEFORM_OUTPUT_MIN) / (ARB_WAVEFORM_INPUT_MAX - ARB_WAVEFORM_INPUT_MIN)
	return int(math.Round(v))
}

/* Aribtrary waveform format:
*
* Waveform Length 2048 point
* Waveform amplitude resolution 12 bit
* Frequency Range 0 - 6MHz
*
* 2048 samples in 16 slices, 128 samples per slice. N=0...F for each slice. Each sample is 0 to 4095 with 2048 as the nominal center
*
 */

// SetArbitrayWaveform send an arbitrary waveform to the generator
func (mhs5200 *MHS5200A) SetArbitraryWaveform(slot uint, data []float64) error {
	if len(data) != ARB_WAVEFORM_NUM_POINTS {
		return fmt.Errorf("An abrbitrary waveform must contain exactly %v samples", ARB_WAVEFORM_NUM_POINTS)
	}
	for slice := 0; slice < ARB_WAVEFORM_NUM_SLICES; slice++ {
		cmd := fmt.Sprintf(":a%x%x", slot, slice)
		for sample := 0; sample < ARB_WAVEFORM_SAMPLES_PER_SLICE; sample++ {
			cmd += fmt.Sprintf("%d", mhs5200.NormalisedToArbitraryWaveform(data[slice*ARB_WAVEFORM_SAMPLES_PER_SLICE+sample]))
			if sample != (ARB_WAVEFORM_SAMPLES_PER_SLICE - 1) {
				cmd += ","
			}
		}
		err := mhs5200.sendCommandAndExpect([]byte(cmd), "ok")
		if err != nil {
			goutils.Log.Printf("%v failed to send arbitrary waveform slice %v", goutils.Funcname(), slice)
			return err
		}
	}
	return nil
}

func (mhs5200 *MHS5200A) SetArbitrayWaveformFromFile(slot uint, filename string) error {
	if len(filename) == 0 {
		return fmt.Errorf("filename is empty")
	}
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	data := make([]float64, ARB_WAVEFORM_NUM_POINTS)
	sample := 0
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		s := scanner.Text()
		if s[0] == '#' { // comment
			continue
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		if v > 1.0 || v < -1.0 {
			return fmt.Errorf("Arbitrary waveform sample values must be between -1.0 and 1.0")
		}
		data[sample] = v
		sample++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if sample != ARB_WAVEFORM_NUM_POINTS {
		return fmt.Errorf("An abrbitrary waveform must contain exactly %v samples, only read %v samples", ARB_WAVEFORM_NUM_POINTS, sample)
	}
	err = mhs5200.SetArbitraryWaveform(slot, data)
	if err != nil {
		return err
	}
	return mhs5200.SetWaveform(1, WAVEFORM_ARB_0+slot)
}

func (mhs5200 *MHS5200A) IsArbirtraryWaveform(v uint) bool {
	if v >= WAVEFORM_ARB_0 && v <= WAVEFORM_ARB_15 {
		return true
	} else if v >= WAVEFORM_ARB_ALT_0 && v <= WAVEFORM_ARB_ALT_15 {
		return true
	}
	return false
}

func (mhs5200 *MHS5200A) SetWaveform(ch uint, v uint) error {
	switch v { // handle our custom waveforms
	case WAVEFORM_SINC:
		data := generateSinc()
		err := mhs5200.SetArbitraryWaveform(WAVEFORM_ARB_15-WAVEFORM_ARB_0, data)
		if err != nil {
			return err
		}
		return mhs5200.SetWaveform(ch, WAVEFORM_ARB_15)

	case WAVEFORM_NORM_SINC:
		data := generateNormalisedSinc()
		err := mhs5200.SetArbitraryWaveform(WAVEFORM_ARB_15-WAVEFORM_ARB_0, data)
		if err != nil {
			return err
		}
		return mhs5200.SetWaveform(ch, WAVEFORM_ARB_15)
	}
	if (v > WAVEFORM_DESCENDING_SAWTOOTH && v < WAVEFORM_ARB_0) || v > WAVEFORM_ARB_15 {
		return fmt.Errorf("%v is not a valid waveform", v)
	}
	err := mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s%dw%d", ch, v)), "ok")
	if err != nil {
		return err
	}
	if mhs5200.IsArbirtraryWaveform(v) {
		// We need to delay before sending the next command
		time.Sleep(2 * time.Second)
	}
	return nil
}

func (mhs5200 *MHS5200A) SetWaveformFromString(ch uint, s string) error {
	if len(s) == 0 {
		return nil
	}
	return mhs5200.SetWaveform(ch, mhs5200.WaveformStringToInt(s))
}

func (mhs5200 *MHS5200A) GetWaveform(ch uint) (uint, error) {
	data, err := mhs5200.sendCommand([]byte(fmt.Sprintf(":r%dw", ch)))
	if err != nil {
		return 0, err
	}
	if len(data) < 4 {
		return 0, fmt.Errorf("data underlow")
	}
	data = data[4:]
	w, err := strconv.ParseUint(string(data), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(w), nil
}

func (mhs5200 *MHS5200A) AmplitudeString(v float64) string {
	return mhs5200.UnitsString(v, "V", true)
}

func (mhs5200 *MHS5200A) SetAmplitude(ch uint, v float64) error {
	if math.IsNaN(v) {
		return nil
	}
	attenuation, err := mhs5200.GetAttenuation(ch)
	if err != nil {
		return err
	}
	if attenuation == ATTENUATION_MINUS_20DB {
		if v < 5e-3 || v > 2.0 {
			return fmt.Errorf("%v is not a valid amplitude", v)
		}
		v *= 1000.0
	} else {
		if v < 5e-3 || v > 20.0 {
			return fmt.Errorf("%v is not a valid amplitude", v)
		}
		v *= 100.0
	}
	return mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s%da%d", ch, int(v))), "ok")
}

func (mhs5200 *MHS5200A) GetAmplitude(ch uint) (float64, error) {
	data, err := mhs5200.sendCommand([]byte(fmt.Sprintf(":r%da", ch)))
	if err != nil {
		return 0.0, err
	}
	if len(data) < 4 {
		return 0.0, fmt.Errorf("data underlow")
	}
	data = data[4:]
	u, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return 0.0, err
	}
	attenuation, err := mhs5200.GetAttenuation(ch)
	if err != nil {
		return 0.0, err
	}
	v := float64(u) / 100.0
	if attenuation == ATTENUATION_MINUS_20DB {
		v /= 10.0
	}
	return v, nil
}

func (mhs5200 *MHS5200A) DutyCycleString(v float64) string {
	return fmt.Sprintf("%.1f%%", v)
}

func (mhs5200 *MHS5200A) SetDutyCycle(ch uint, v float64) error {
	if math.IsNaN(v) {
		return nil
	}
	if v < 0.0 || v > 99.9 {
		return fmt.Errorf("%v is not a valid duty cycle", v)
	}
	return mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s%dd%d", ch, int(v*10.0))), "ok")
}

func (mhs5200 *MHS5200A) GetDutyCycle(ch uint) (float64, error) {
	data, err := mhs5200.sendCommand([]byte(fmt.Sprintf(":r%dd", ch)))
	if err != nil {
		return 0.0, err
	}
	if len(data) < 4 {
		return 0.0, fmt.Errorf("data underlow")
	}
	data = data[4:]
	u, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return 0.0, err
	}
	v := float64(u) / 10.0
	return v, nil
}

func (mhs5200 *MHS5200A) OffsetString(v float64) string {
	return mhs5200.AmplitudeString(v)
}

func (mhs5200 *MHS5200A) SetOffset(ch uint, v float64) error {
	if v == math.MaxUint32 {
		return nil
	}
	ampl, err := mhs5200.GetAmplitude(ch)
	if err != nil {
		return err
	}
	v = v / ampl * 100.0
	if v < -120 || v > 120 {
		return fmt.Errorf("%v is not a valid offset. Supported values are between -120%% and 120%% of the amplitude value", v)
	}
	return mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s%do%d", ch, int(math.Round(v))+120)), "ok")
}

func (mhs5200 *MHS5200A) GetOffset(ch uint) (float64, error) {
	ampl, err := mhs5200.GetAmplitude(ch)
	if err != nil {
		return math.NaN(), err
	}
	data, err := mhs5200.sendCommand([]byte(fmt.Sprintf(":r%do", ch)))
	if err != nil {
		return math.NaN(), err
	}
	if len(data) < 4 {
		return math.NaN(), fmt.Errorf("data underlow")
	}
	data = data[4:]
	v, err := strconv.ParseFloat(string(data), 64)
	if err != nil {
		return math.NaN(), err
	}
	v = ampl * (v - 120.0) / 100.0
	return v, nil
}

func (mhs5200 *MHS5200A) PhaseString(v uint) string {
	return fmt.Sprintf("%d°", v)
}

func (mhs5200 *MHS5200A) SetPhase(ch uint, v uint) error {
	if v == math.MaxUint32 {
		return nil
	}
	if v > 360 {
		return fmt.Errorf("%v is not a valid phase", v)
	}
	return mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s%dp%d", ch, v)), "ok")
}

func (mhs5200 *MHS5200A) GetPhase(ch uint) (uint, error) {
	data, err := mhs5200.sendCommand([]byte(fmt.Sprintf(":r%dp", ch)))
	if err != nil {
		return 0, err
	}
	if len(data) < 4 {
		return 0, fmt.Errorf("data underlow")
	}
	data = data[4:]
	v, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

func (mhs5200 *MHS5200A) AttenuationString(v uint) string {
	if v == ATTENUATION_MINUS_20DB {
		return "-20dB"
	}
	return "0dB"
}

func (mhs5200 *MHS5200A) SetAttenuation(ch uint, v uint) error {
	if v == math.MaxUint32 {
		return nil
	}
	return mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s%dy%d", ch, v)), "ok")
}

func (mhs5200 *MHS5200A) GetAttenuation(ch uint) (uint, error) {
	// attenuation 0 = -20dB, 1 = 0dB
	data, err := mhs5200.sendCommand([]byte(fmt.Sprintf(":r%dy", ch)))
	if err != nil {
		return 0, err
	}
	if len(data) < 4 {
		return 0, fmt.Errorf("data underlow")
	}
	data = data[4:]
	v, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(v), nil
}

func (mhs5200 *MHS5200A) SetSweepState(v bool) error {
	state := 0
	if v {
		state = 1
	}
	return mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s8b%d", state)), "ok")
}

func (mhs5200 *MHS5200A) GetSweepState() (bool, error) {
	v, err := mhs5200.sendCommandAndExpectUint([]byte(fmt.Sprintf(":r8b")))
	if err != nil {
		return false, err
	}
	return v != 0, err
}

func (mhs5200 *MHS5200A) SweepTypeString(v uint) string {
	switch v {
	case SWEEP_LINEAR:
		return "linear"

	case SWEEP_LOG:
		return "logarithmic"
	}
	return "unknown"
}

func (mhs5200 *MHS5200A) SweepTypeStringToInt(s string) uint {
	switch s {
	case "linear":
		return SWEEP_LINEAR

	case "logarithmic":
		return SWEEP_LOG

	case "log":
		return SWEEP_LOG
	default:
		return math.MaxUint32
	}
}

func (mhs5200 *MHS5200A) SetSweepStart(v float64) error {
	if math.IsNaN(v) {
		return nil
	}
	if v < 0.0 || v > 25.0e6 {
		return fmt.Errorf("%v is not a valid frequency", v)
	}
	return mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s3f%d", int(v*100.0))), "ok")
}

func (mhs5200 *MHS5200A) GetSweepStart() (float64, error) {
	u, err := mhs5200.sendCommandAndExpectUint([]byte(fmt.Sprintf(":r3f")))
	if err != nil {
		return 0.0, err
	}
	return float64(u) / 100.0, nil
}

func (mhs5200 *MHS5200A) SetSweepEnd(v float64) error {
	if math.IsNaN(v) {
		return nil
	}
	if v < 0.0 || v > 25.0e6 {
		return fmt.Errorf("%v is not a valid frequency", v)
	}
	return mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s4f%d", int(v*100.0))), "ok")
}

func (mhs5200 *MHS5200A) GetSweepEnd() (float64, error) {
	u, err := mhs5200.sendCommandAndExpectUint([]byte(fmt.Sprintf(":r4f")))
	if err != nil {
		return 0.0, err
	}
	return float64(u) / 100.0, nil
}

func (mhs5200 *MHS5200A) SetSweepDuration(v uint) error {
	if v == math.MaxUint32 {
		return nil
	}
	if v > 999 {
		return fmt.Errorf("%v is not a valid duration", v)
	}
	return mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s1t%d", int(v))), "ok")
}

func (mhs5200 *MHS5200A) GetSweepDuration() (uint, error) {
	u, err := mhs5200.sendCommandAndExpectUint([]byte(fmt.Sprintf(":r1t")))
	if err != nil {
		return 0, err
	}
	return uint(u), nil
}

func (mhs5200 *MHS5200A) SetSweepType(v uint) error {
	if v == math.MaxUint32 {
		return nil
	}
	return mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s7b%d", int(v))), "ok")
}

func (mhs5200 *MHS5200A) GetSweepType() (uint, error) {
	u, err := mhs5200.sendCommandAndExpectUint([]byte(fmt.Sprintf(":r7b")))
	if err != nil {
		return 0, err
	}
	return uint(u), nil
}

func (mhs5200 *MHS5200A) SetSweep(v *SWEEPVALS) error {
	if v == nil {
		return fmt.Errorf("null parameters")
	}
	var err error
	err = mhs5200.SetDutyCycle(1, v.Duty)
	if err != nil {
		return err
	}
	err = mhs5200.SetWaveformFromString(1, v.Waveform)
	if err != nil {
		return err
	}
	err = mhs5200.SetSweepStart(v.Startf)
	if err != nil {
		return err
	}
	err = mhs5200.SetSweepEnd(v.Endf)
	if err != nil {
		return err
	}
	err = mhs5200.SetSweepDuration(v.Duration)
	if err != nil {
		return err
	}
	err = mhs5200.SetSweepType(v.Type)
	if err != nil {
		return err
	}
	return nil
}

func (mhs5200 *MHS5200A) GetSweep() (*SWEEPVALS, error) {
	var err error
	v := SWEEPVALS{}
	v.Duty, err = mhs5200.GetDutyCycle(1)
	if err != nil {
		return nil, err
	}
	w, err := mhs5200.GetWaveform(1)
	if err != nil {
		return nil, err
	}
	v.Waveform = mhs5200.WaveformString(w)

	v.Startf, err = mhs5200.GetSweepStart()
	if err != nil {
		return nil, err
	}

	v.Endf, err = mhs5200.GetSweepEnd()
	if err != nil {
		return nil, err
	}

	v.Duration, err = mhs5200.GetSweepDuration()
	if err != nil {
		return nil, err
	}

	v.Type, err = mhs5200.GetSweepType()
	if err != nil {
		return nil, err
	}

	return &v, err
}

func (mhs5200 *MHS5200A) SetOnOff(v bool) error {
	state := 0
	if v {
		state = 1
	}
	return mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s1b%d", state)), "ok")
}

func (mhs5200 *MHS5200A) SelectChannel(ch uint) error {
	if ch == 0 {
		return nil
	}
	if ch > 2 {
		return fmt.Errorf("%v is not a valid channel", ch)
	}
	return mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":s2b%d", ch)), "ok")
}

func (mhs5200 *MHS5200A) Save(v uint) error {
	if v > 15 {
		return fmt.Errorf("%v is not a valid save position", v)
	}
	return mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":su%02d", v)), "ok")
}

func (mhs5200 *MHS5200A) Load(v uint) error {
	if v > 15 {
		return fmt.Errorf("%v is not a valid load position", v)
	}
	return mhs5200.sendCommandAndExpect([]byte(fmt.Sprintf(":sv%02d", v)), "ok")
}

func (mhs5200 *MHS5200A) GetModel() (string, error) {
	data, err := mhs5200.sendCommand([]byte(fmt.Sprintf(":r0c")))
	if err != nil {
		return "", err
	}
	if len(data) < 4 {
		return "", fmt.Errorf("data underlow")
	}
	data = data[4:]
	return "MHS-" + string(data[:5]), nil
}

func (mhs5200 *MHS5200A) GetFirmwareVersion() (float64, error) {
	data, err := mhs5200.sendCommand([]byte(fmt.Sprintf(":r1c")))
	if err != nil {
		return 0.0, err
	}
	if len(data) < 12 {
		return 0.0, fmt.Errorf("data underlow")
	}
	data = data[9:12]
	u, err := strconv.ParseUint(string(data), 10, 64)
	if err != nil {
		return 0.0, err
	}
	v := float64(u) / 100.0
	return v, nil
}

func (mhs5200 *MHS5200A) GetSerial() (string, error) {
	data, err := mhs5200.sendCommand([]byte(fmt.Sprintf(":r2c")))
	if err != nil {
		return "", err
	}
	if len(data) < 16 {
		return "", fmt.Errorf("data underlow")
	}
	data = data[12:16]
	return string(data), nil
}

func (mhs5200 *MHS5200A) ShowSweepConfig() error {
	// sweep is only output on channel 1
	sweep, err := mhs5200.GetSweepState()
	if err != nil {
		return err
	}
	fmt.Printf("Sweep config\n")
	fmt.Printf("\tActive:\t\t%v\n", mhs5200.BooleanString(sweep))
	sweepvals, err := mhs5200.GetSweep()
	if err != nil {
		return err
	}
	fmt.Printf("\tWaveform:\t%v\n", sweepvals.Waveform)
	if sweepvals.Waveform == WAVEFORM_SQUARE_STR {
		fmt.Printf("\tDutyCycle:\t%v\n", mhs5200.DutyCycleString(sweepvals.Duty))
	}
	fmt.Printf("\tStart:\t\t%v\n", mhs5200.FrequencyString(sweepvals.Startf))
	fmt.Printf("\tEnd:\t\t%v\n", mhs5200.FrequencyString(sweepvals.Endf))
	fmt.Printf("\tDuration:\t%v seconds\n", sweepvals.Duration)
	fmt.Printf("\tType:\t\t%v\n", mhs5200.SweepTypeString(sweepvals.Type))

	return nil
}

func (mhs5200 *MHS5200A) ShowChannelConfig(ch uint) error {
	fmt.Printf("Channel %d config\n", ch)
	err := mhs5200.SelectChannel(ch)
	if err != nil {
		return err
	}
	freq, err := mhs5200.GetFrequency(ch)
	if err != nil {
		return err
	}
	w, err := mhs5200.GetWaveform(ch)
	if err != nil {
		return err
	}
	ampl, err := mhs5200.GetAmplitude(ch)
	if err != nil {
		return err
	}
	duty, err := mhs5200.GetDutyCycle(ch)
	if err != nil {
		return err
	}
	offset, err := mhs5200.GetOffset(ch)
	if err != nil {
		return err
	}
	phase, err := mhs5200.GetPhase(ch)
	if err != nil {
		return err
	}
	attenuation, err := mhs5200.GetAttenuation(ch)
	if err != nil {
		return err
	}

	fmt.Printf("\tFrequency:\t%v\n", mhs5200.FrequencyString(freq))
	fmt.Printf("\tWaveform:\t%v\n", mhs5200.WaveformString(w))
	fmt.Printf("\tAmplitude:\t%v\n", mhs5200.AmplitudeString(ampl))
	fmt.Printf("\tDutyCycle:\t%v\n", mhs5200.DutyCycleString(duty))
	fmt.Printf("\tOffset:\t\t%v\n", mhs5200.OffsetString(offset))
	fmt.Printf("\tPhase:\t\t%v\n", mhs5200.PhaseString(phase))
	fmt.Printf("\tAttenuation:\t%v\n", mhs5200.AttenuationString(attenuation))

	return nil
}

func (mhs5200 *MHS5200A) ShowConfig() error {
	model, err := mhs5200.GetModel()
	if err != nil {
		return err
	}
	version, err := mhs5200.GetFirmwareVersion()
	if err != nil {
		return err
	}
	serial, err := mhs5200.GetSerial()
	if err != nil {
		return err
	}
	fmt.Printf("Model:\t\t%v\n", model)
	fmt.Printf("Serial:\t\t%v\n", serial)
	fmt.Printf("Firmware:\t%v\n", version)
	fmt.Println("")
	err = mhs5200.ShowChannelConfig(1)
	if err != nil {
		return err
	}
	fmt.Println("")
	mhs5200.ShowChannelConfig(2)
	if err != nil {
		return err
	}
	return nil
}

func (mhs5200 *MHS5200A) ApplyChannelConfig(v *CHANNELVALS) error {
	if v == nil {
		return fmt.Errorf("null data")
	}
	var err error
	if v.Channel != math.MaxUint32 {
		err = mhs5200.SelectChannel(v.Channel)
		if err != nil {
			return err
		}
	}
	if v.Attenuation != math.MaxUint32 {
		err = mhs5200.SetAttenuation(v.Channel, v.Attenuation)
		if err != nil {
			return err
		}
	}
	if !math.IsNaN(v.Frequency) {
		err = mhs5200.SetFrequency(v.Channel, v.Frequency)
		if err != nil {
			return err
		}
	}
	if len(v.Waveform) > 0 {
		err = mhs5200.SetWaveformFromString(v.Channel, v.Waveform)
		if err != nil {
			return err
		}
	}
	if !math.IsNaN(v.Amplitude) {
		err = mhs5200.SetAmplitude(v.Channel, v.Amplitude)
		if err != nil {
			return err
		}
	}
	if !math.IsNaN(v.Phase) {
		err = mhs5200.SetPhase(v.Channel, uint(math.Round(v.Phase)))
		if err != nil {
			return err
		}
	}
	if !math.IsNaN(v.Duty) {
		err = mhs5200.SetDutyCycle(v.Channel, v.Duty)
		if err != nil {
			return err
		}
	}
	if !math.IsNaN(v.Offset) {
		err = mhs5200.SetOffset(v.Channel, v.Offset)
		if err != nil {
			return err
		}
	}
	return nil
}

func (mhs5200 *MHS5200A) mhs5200() {
	measure_ticker := time.NewTicker(1 * time.Second)
	defer measure_ticker.Stop()
	for {
		select {
		case <-mhs5200.quit:
			mhs5200.wg.Done()
			return

		case <-measure_ticker.C:
			if mhs5200.measure {
				s, err := mhs5200.GetMeasurementAsString()
				if err == nil {
					fmt.Println(s)
				} else {
					goutils.Log.Print(err)
				}
			}
		}
	}
}

func NewMHS5200A(port string) (*MHS5200A, error) {
	if len(port) == 0 {
		err := fmt.Errorf("%s: no port specified", goutils.Funcname())
		return nil, err
	}
	config := &serial.Config{
		Name:        port,
		Baud:        57600,
		Size:        8,
		Parity:      serial.ParityNone,
		StopBits:    serial.Stop1,
		ReadTimeout: time.Millisecond * 5,
	}
	stream, err := serial.OpenPort(config)
	if err != nil {
		return nil, err
	}
	stream.Flush()
	mhs5200 := &MHS5200A{
		stream: stream,
		quit:   make(chan struct{}),
	}
	mhs5200.wg.Add(1)
	go mhs5200.mhs5200()
	return mhs5200, nil
}

func (mhs5200 *MHS5200A) Close() {
	close(mhs5200.quit)
	mhs5200.wg.Wait()
	if mhs5200.stream != nil {
		mhs5200.stream.Close()
	}
}

func convertWaveFile(filename string) error {
	if len(filename) == 0 {
		return fmt.Errorf("filename is empty")
	}
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	data := make([]float64, 0)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		v, err := strconv.ParseFloat(scanner.Text(), 64)
		if err != nil {
			return err
		}
		data = append(data, v)
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	autoNormalise(data, -1.0, 1.0)
	if len(data) == 1024 { // old style 1024 point file convert it
		waveform := make([]float64, ARB_WAVEFORM_NUM_POINTS)
		last := len(data) - 1
		for i, _ := range data {
			waveform[i*2] = data[i]
			if i == last { // wraparound
				waveform[i*2+1] = (data[i] + data[0]) * 0.5
			} else {
				waveform[i*2+1] = (data[i] + data[i+1]) * 0.5
			}
		}
		data = waveform
	}
	for i, _ := range data {
		fmt.Println(data[i])
	}
	return nil

}
