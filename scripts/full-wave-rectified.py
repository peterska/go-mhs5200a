#!/bin/env python

# Generate a full wave rectified waveform to use with the MHS-5200A function generator
from math import pi
from math import sin

num_points = 2048
amplitude = 1.0
print("# generated  by scripts/full-wave-rectified.py")

# Flags to allow peaks only one point wide
for i in range(num_points):
    x = i/num_points # Fraction along X axis
    y = amplitude*sin(2*pi*x)
    if y < 0.0:
        y *= -1.0
    print(y)
