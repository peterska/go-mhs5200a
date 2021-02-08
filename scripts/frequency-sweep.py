#!/bin/env python

# generata a frequency sweep

import numpy as np
from scipy.signal import chirp
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
    waveform = chirp(t, f0=1, f1=50, t1=1, method='linear')

    if showplot:
        plt.plot(t, waveform, '--')
        plt.show()
    else:
        print("# frequency sweep using scipy.signal.chirp")
        for y in waveform:
            print(y)

if __name__ == "__main__":
   main(sys.argv[1:])

