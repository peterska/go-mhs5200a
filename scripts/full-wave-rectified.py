#!/bin/env python

# Generate a half wave rectified waveform to use with the MHS-5200A function generator
from math import pi
from math import sin
import numpy as np
import matplotlib.pyplot as plt
import sys, getopt

def main(argv):
    showplot = False
    try:
        opts, args = getopt.getopt(argv,"p")
    except getopt.GetoptError:
        sys.exit(2)
    for opt, arg in opts:
        if opt == '-p':
            showplot = True
    
    num_points = 2048
    amplitude = 1.0
    t = np.linspace(0, 1, 2048)
    waveform = np.zeros_like(t)
    # Flags to allow peaks only one point wide
    for i in range(num_points):
        x = i/num_points # Fraction along X axis
        y = amplitude*sin(2*pi*x)
        if y < 0.0:
            y *= -1.0
        waveform[i] = y
    if showplot:
        plt.plot(t, waveform, '--')
        plt.show()
    else:
        print("# generated  by scripts/full-wave-rectified.py")
        for y in waveform:
            print(y)

if __name__ == "__main__":
   main(sys.argv[1:])
