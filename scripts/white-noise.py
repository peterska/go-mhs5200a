#!/bin/env python

# generate white noise

import numpy as np
from scipy.signal import square
import matplotlib.pyplot as plt
import sys, getopt
from random import gauss
from random import seed

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
    t = np.linspace(0, 1, num_points)
    waveform = np.zeros_like(t)
    seed()
    for i in range(num_points):
        waveform[i] = gauss(0.0, 2.0)
    ymax = max(waveform)
    for i in range(num_points):
        waveform[i] /= ymax
    
    if showplot:
        plt.plot(t, waveform, '--')
        plt.show()
    else:
        print("# white noise with gaussian distribution")
        for y in waveform:
            print(y)

if __name__ == "__main__":
   main(sys.argv[1:])
