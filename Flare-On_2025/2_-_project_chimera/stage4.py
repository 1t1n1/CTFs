import os
import sys
import emoji
import random
import asyncio
import cowsay
import pyjokes
import art
from arc4 import ARC4

async def activate_catalyst():
    LEAD_RESEARCHER_SIGNATURE = b'm\x1b@I\x1dAoe@\x07ZF[BL\rN\n\x0cS'
    ENCRYPTED_CHIMERA_FORMULA = b'r2b-\r\x9e\xf2\x1fp\x185\x82\xcf\xfc\x90\x14\xf1O\xad#]\xf3\xe2\xc0L\xd0\xc1e\x0c\xea\xec\xae\x11b\xa7\x8c\xaa!\xa1\x9d\xc2\x90'

    print("--- Catalyst Serum Injected ---")
    print("Verifying Lead Researcher's credentials via biometric scan...")

    # current_user = os.getlogin().encode()
    current_user = 'G0ld3n_Tr4nsmut4t10n'.encode()
    user_signature = bytes(c ^ i + 42 for i, c in enumerate(current_user))

    await asyncio.sleep(0.01)

    status = 'pending'

    if status == 'pending':
        if user_signature == LEAD_RESEARCHER_SIGNATURE:
            art.tprint("AUTHENTICATION   SUCCESS", font="small")
            print("Biometric scan MATCH. Identity confirmed as Lead Researcher.")
            print("Finalizing Project Chimera...")

            arc4_decipher = ARC4(current_user)
            decrypted_formula = arc4_decipher.decrypt(ENCRYPTED_CHIMERA_FORMULA).decode()

            cowsay.cow("I am alive! The secret formula is:\n" + decrypted_formula)
        else:
            art.tprint("AUTHENTICATION   FAILED", font="small")
            print("Impostor detected, my genius cannot be replicated!")
            print("The resulting specimen has developed an unexpected, and frankly useless, sense of humor.")

            joke = pyjokes.get_joke(language='en', category='all')
            animals = cowsay.char_names[1:]
            print(cowsay.get_output_string(random.choice(animals), joke))

            sys.exit(1)
    else:
        print("System error: Unknown experimental state.")

asyncio.run(activate_catalyst())