#!/bin/env python

# generata a frequency sweep

import numpy as np
from scipy.signal import chirp
import matplotlib.pyplot as plt

showplot = False
t = np.linspace(0, 1, 2048)
waveform = chirp(t, f0=1, f1=50, t1=1, method='linear')

print("# frequency sweep using scipy.signal.chirp")
for y in waveform:
    print(y)

if showplot:
    plt.plot(t, waveform, '--')
    plt.show()

