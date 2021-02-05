#!/bin/env python

# Generate an AM modulated signal

import numpy as np

print("# generated  by scripts/am-modulation.py")

time = np.linspace(0.0, 1.0, 2048)
carrier_amplitude = 1.0
modulation_amplitude = 0.5
carrier_frequency = 10.0
modulation_frequency = 1.0
modulation_index = 1.0


carrier = carrier_amplitude * np.cos(2 * np.pi * carrier_frequency * time)
modulator = modulation_amplitude * np.cos(2 * np.pi * modulation_frequency * time)
product = carrier_amplitude * (1.0 + modulation_index * np.cos(2.0 * np.pi * modulation_frequency * time)) * np.cos(2.0 * np.pi * carrier_frequency * time)

for y in product:
    print(y / (carrier_amplitude * (1.0 + modulation_index)))
