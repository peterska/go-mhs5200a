#!/bin/env python

# Generate an AM modulated signal

import numpy as np
import matplotlib.pyplot as plt
import sys, getopt

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
    
    time = np.linspace(0.0, 1.0, 2048)
    carrier_amplitude = 1.0
    modulation_amplitude = 0.5
    carrier_frequency = 10.0
    modulation_frequency = 1.0
    modulation_index = 0.5

    carrier = carrier_amplitude * np.cos(2 * np.pi * carrier_frequency * time)
    modulator = modulation_amplitude * np.cos(2 * np.pi * modulation_frequency * time)
    product = carrier_amplitude * (1.0 + modulation_index * np.cos(2.0 * np.pi * modulation_frequency * time)) * np.cos(2.0 * np.pi * carrier_frequency * time)

    normalise(product, -1.0, 1.0)
    #product /= (carrier_amplitude * (1.0 + modulation_index))

    if showplot:
        plt.plot(time, product, '--')
        plt.show()
    else:
        print("# generated  by scripts/am-modulation.py")
        for y in product:
            print(y)

if __name__ == "__main__":
   main(sys.argv[1:])

