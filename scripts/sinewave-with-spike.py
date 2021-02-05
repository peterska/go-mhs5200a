#!/bin/env python

# example from https://www.bkprecision.com/support/downloads/function-and-arbitrary-waveform-generator-guidebook.html
from math import pi
from math import sin

num_points = 2048
amplitude = 4000
peak_amplitude = 8000
print("# generated  by scripts/sinewave-with-spike.py")

# Flags to allow peaks only one point wide
positive_done = False
negative_done = False
threshold = 1e-4

for i in range(num_points):
    x = i/num_points # Fraction along X axis
    y = int(amplitude*sin(2*pi*x))
    if not positive_done and abs(x - 1/4) < threshold:
        positive_done = True
        y = peak_amplitude
    if not negative_done and abs(x - 3/4) < threshold:
        negative_done = True
        y = -peak_amplitude
    print(y / peak_amplitude)
