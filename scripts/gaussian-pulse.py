#!/bin/env python

# generata a gaussian pulse

import numpy as np
from scipy.signal import gausspulse
import matplotlib.pyplot as plt

showplot = False
t = np.linspace(-1, 1, 2048)
waveform = gausspulse(t, fc=5)

print("# gaussian pulse using scipy.signal.gausspulse")
for y in waveform:
    print(y)

if showplot:
    plt.plot(t, waveform, '--')
    plt.show()

