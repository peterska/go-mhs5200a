#!/bin/env python

# generata pwm modulated signal

import numpy as np
from scipy.signal import square
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
    
    t = np.linspace(0, 1, 2048)
    sig = np.sin(2 * np.pi * t)
    waveform = square(2 * np.pi * 25 * t, duty=(sig + 1)/2)

    if showplot:
        plt.plot(t, waveform, '--')
        plt.show()
    else:
        print("# pwm modulation using scipy.signal.square")
        for y in waveform:
            print(y)

if __name__ == "__main__":
   main(sys.argv[1:])
