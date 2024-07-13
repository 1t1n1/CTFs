Il faut dump les différents segments de la ram avec mem_dump. Exemple de commande:

`python3 ./esptool/esptool.py --chip esp32-S3 --port /dev/serial/by-id/usb-Espressif_USB_JTAG_serial_debug_unit_30:30:F9:3D:C6:18-if00 --baud 460800 dump_mem 0x3fc88000 0x77fff ./data/ram_segments/sram1.bin`

Pour dump la mémoire  de la première section de la RAM en mémoire (memory mapping selon documentation espressif).
