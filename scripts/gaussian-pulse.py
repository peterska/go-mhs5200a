#!/bin/env python

# generate a gaussian pulse

import numpy as np
from scipy.signal import gausspulse
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
    
    t = np.linspace(-1, 1, 2048)
    waveform = gausspulse(t, fc=5)

    if showplot:
        plt.plot(t, waveform, '--')
        plt.show()
    else:
        print("# gaussian pulse using scipy.signal.gausspulse")
        for y in waveform:
            print(y)

if __name__ == "__main__":
   main(sys.argv[1:])
