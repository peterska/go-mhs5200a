#!/bin/env python

# Generate a FM modulated signal

import numpy as np

print("# generated  by scripts/fm-modulation.py")

time = np.linspace(0, 1, 2048)
modulator_frequency = 1.0
carrier_frequency = 10.0
modulation_index = 1.0

modulator = np.sin(2.0 * np.pi * modulator_frequency * time) * modulation_index
carrier = np.sin(2.0 * np.pi * carrier_frequency * time)
product = np.zeros_like(modulator)

for i, t in enumerate(time):
    product[i] = np.sin(2. * np.pi * (carrier_frequency * t + modulator[i]))

for y in product:
    print(y)
