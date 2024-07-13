## Connect to the badge
Using a cable that is able to transmit data, connect to the badge using the following command:

`[sudo] screen /dev/ttyUSB0 115800`

To quit, press the following keys:

`Ctrl+a, k`

You should now have access to a buggy CLI. Use the `restart` command to access the fully functional CLI.

---

## Reverse 101
Use the ghidra processor module for XTensa that someone made online: https://github.com/yath/ghidra-xtensa
When looking at verify, reconstruct the stack. It should give you `11a922186c46496eb1317c128b361720` which 
yields the flag `flag-REverSINg_xteNSA_is_NOt_that_HArd`

---

## Reverse 102
Same as Reverse 101, but the conditionals were different. Still had to transform from hex to ascii and 
rearrange the offset. I am pretty certain now that user_code corresponds to what we send. The correct code
is `f219e6cdb1fa4a48b160d00d61X18f93` and the flag we receive is `flag-this_is_a_big_huge_enormous_condition`


---

## Firmware

Espressif gives access to a free tool, `esptool`, which can be used to dump the firmware using the following command:

`python3 ./esptool.py --chip esp32 --port /dev/ttyUSB0 --baud 460800 read_flash 0x0 0x1000000 ./data/firmware.bin`

Fun fact: I found my WiFi BSSID and password in plain text in the firmware XD

You can now extract the partitions from the firmware using a tool called [esp32_image_parser.py](https://github.com/tenable/esp32_image_parser)

_Show partitions_:

`python3 ./esp32_image_parser/esp32_image_parser.py show_partitions data/firmware.bin`

_Extract partition_:

`python3 ./esp32_image_parser/esp32_image_parser.py -output data/nvs.bin -partition nvs dump_partition data/firmware.bin`

_Extract NVS partition as JSON_:

`python3 ./esp32_image_parser/esp32_image_parser.py -nvs_output_type json -partition nvs dump_nvs data/firmware.bin | grep -v "Dumping partition 'nvs' to nvs_out.bin" >  data/nvs.json`

The `factory` partition follows the SPIFFS Filesystem, as described in the [Espressif documentation](https://docs.espressif.com/projects/esp-idf/en/latest/esp32/api-reference/storage/spiffs.html). From there, you can use a tool such as [mkspiffs](https://github.com/igrr/mkspiffs) to extract the filesystem from the dump and get access to the assets and images that the game uses.

To extract the ELF, this command seemed to work:

`python3 ./esp32_image_parser.py create_elf ../data/factory.bin -appimage -output ../data/factory.elf`

Note that I had to apply the changes that correspond to this pull request: https://github.com/tenable/esp32_image_parser/pull/3 (Add option to command line to allow creating elf file from a single app image instead of the full flash dump). I was at commit 0c9ff48 when this happened.

After getting the ELF, loading it into ghidra should not give you much except the main function. You have to manually decompile the functions yourself. When a call is made to, let's say, PTR_DAT_400d0084. This means you have to get to the location pointed and manually
decompile (D key) and set as function (F key) at the given location.

### **Last Flag**

By opening the NVS partition (`python3 ./esp32_image_parser.py -partition nvs -nvs_output_type json dump_nvs ../data/complete_dump.bin > ../data/nvs.json`), we could find the `save` entry that corresponds to where the number of found flags are located. I read from someone writeup that you could create an empty NVS partition to reset the game to zero, and dump the NVS parititon each time you get the flag. You could then compare the dumps to find which address corresponds to each flag. You could also look at the `Erased` NVS entries and compare them together with the `Written` one to find which bit you need to flip to get all flags.

Finally, to flash the NVS partition back into the badge, use the following command: `python3 ./esptool/esptool.py --chip esp32 --port /dev/ttyUSB0 --baud 460800 write_flash 0x9000 ./data/nvs_end_game.bin`
The offset of the partition is given by the badge cli when you enter `restart`

// TODO : 
- Rewrite the player's sayings
- Make me do impossible things
