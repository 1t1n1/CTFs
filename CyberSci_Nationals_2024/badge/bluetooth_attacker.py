import bluetooth

# Replace with your Bluetooth device's address
target_address = 'E9:B1:C3:DB:67:C6'

# Discover nearby Bluetooth devices
nearby_devices = bluetooth.discover_devices()
print(nearby_devices)

# Find the device with the specified address
for addr in nearby_devices:
    if addr == target_address:
        target_device = addr
        break

# Connect to the target Bluetooth device
port = 1  # RFCOMM port number for serial communication
sock = bluetooth.BluetoothSocket(bluetooth.RFCOMM)
sock.connect((target_device, port))

## Winner solution: use bleak
