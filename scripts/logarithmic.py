#!/bin/env python

# generate logarithmic waveform

import numpy as np
import matplotlib.pyplot as plt
import sys, getopt
from math import log

def normalise(w, minval, maxval):
    ymin = min(w)
    ymax = max(w)
    for i in range(len(w)):
        w[i] = minval + (w[i] - ymin) / (ymax - ymin) * (maxval - minval)
    
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
    t = np.linspace(0.05, 1, num_points)
    waveform = np.zeros_like(t)
    for i in range(num_points):
        waveform[i] = log(t[i])

    normalise(waveform, -1.0, 1.0)
    
    if showplot:
        plt.plot(t, waveform, '--')
        plt.show()
    else:
        print("# logarithmic waveform")
        for y in waveform:
            print(y)

if __name__ == "__main__":
   main(sys.argv[1:])
