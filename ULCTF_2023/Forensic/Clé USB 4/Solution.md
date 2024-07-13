This one was interesting. You had to take a look at the macro of the excel sheet and get the flag from it. 

First, I extracted the value of the `zf3dede9f2b604c66b0343ed1d25408da` variable by adding an instruction after its last assignation and placing a breakpoint there. I then ran the macro (I had to disable macro security to low, add the file location as a trusted source and reopen it) and extracted the value from the variable. The rest of the solution can be found in the file solution.py
