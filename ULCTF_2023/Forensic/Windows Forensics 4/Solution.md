When looking at the Microsoft-Windows-DeviceSetupManager%4Admin logs, you could find that around 10:13:39 (one minute before the virus was detected) a new DataTraveler 3.0 device was registered in a log. 

The flag doesn't consider the version numbers, so you get:

ULCTF-Kingston_DataTraveler

You could also find the flag by looking at the Microsoft-Windows-Kernel-PnP%4Configuration logs at 10:13:39. A log contained the following information: "Device USBSTOR\Disk&Ven_Kingston&Prod_DataTraveler_3.0&Rev_\20CF30E117EBF190561E39B0&0 was started." from which you can extract the flag