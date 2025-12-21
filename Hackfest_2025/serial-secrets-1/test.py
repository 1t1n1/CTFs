from vcdvcd import VCDVCD

# Load VCD file
vcd = VCDVCD("uart_log.vcd")

# Get TX signal data
# tx_signal = vcd.signals["top.uart.tx"].tv  # list of (timestamp, value)
tx_signal = vcd['logic.UART_TX'].tv # list of (timestamp, value)
for i in range ()
bit_period_ns = 1e9 / 57600 # ≈ 8680 ns per bit

def decode_uart(tx_data, bit_period):
    bits = []
    bytes_out = []
    i = 0
    while i < len(tx_data) - 1:
        t, val = tx_data[i]
        # Detect start bit (falling edge)
        try:
            if val == '0':
                start_time = t + bit_period * 1.5
                byte_val = 0
                for bit in range(8):
                    sample_time = start_time + bit * bit_period
                    # find closest timestamp in tx_data
                    v = next(val for (time, val) in tx_data if time >= sample_time)
                    byte_val |= (int(v) << bit)
                bytes_out.append(byte_val)
                i += 10  # skip roughly one frame
            i += 1
        except:
            return bytes_out
    return bytes_out

decoded = decode_uart(tx_signal, bit_period_ns)
print("Decoded bytes:", decoded)
print("As text:", ''.join(chr(b) for b in decoded))
