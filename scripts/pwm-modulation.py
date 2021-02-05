#!/bin/env python

# generata pwm modulated signal

import numpy as np
from scipy.signal import square
import matplotlib.pyplot as plt

showplot = False
t = np.linspace(0, 1, 2048)
sig = np.sin(2 * np.pi * t)
waveform = square(2 * np.pi * 25 * t, duty=(sig + 1)/2)

print("# pwm modulation using scipy.signal.square")
for y in waveform:
    print(y)

if showplot:
    plt.plot(t, waveform, '--')
    plt.show()

