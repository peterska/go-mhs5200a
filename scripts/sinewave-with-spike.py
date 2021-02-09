#!/bin/env python

# example from https://www.bkprecision.com/support/downloads/function-and-arbitrary-waveform-generator-guidebook.html
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
    amplitude = 4000
    peak_amplitude = 8000

    # Flags to allow peaks only one point wide
    positive_done = False
    negative_done = False
    threshold = 1e-4

    t = np.linspace(0, 1, 2048)
    waveform = np.zeros_like(t)

    for i in range(num_points):
        x = i/num_points # Fraction along X axis
        y = int(amplitude*sin(2*pi*x))
        if not positive_done and abs(x - 1/4) < threshold:
            positive_done = True
            y = peak_amplitude
        if not negative_done and abs(x - 3/4) < threshold:
            negative_done = True
            y = -peak_amplitude
        waveform[i] = y / peak_amplitude

    if showplot:
        plt.plot(t, waveform, '--')
        plt.show()
    else:
        print("# generated  by scripts/sinewave-with-spike.py")
        for y in waveform:
            print(y)

if __name__ == "__main__":
   main(sys.argv[1:])
