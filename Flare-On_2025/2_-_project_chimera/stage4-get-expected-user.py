signature = b'm\x1b@I\x1dAoe@\x07ZF[BL\rN\n\x0cS'

# expected_user = [chr((c - 42) ^ i) for i, c in enumerate(signature)]
expected_user = "".join([chr(c ^ (i + 42)) for i, c in enumerate(signature)])
print(expected_user)