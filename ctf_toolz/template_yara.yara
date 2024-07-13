rule TemplateYara
{
    strings:
        $flag_xor = "flag" xor
	$FLAG_xor = "FLAG" xor
	$flag_base64 = "flag" base64
	$FLAG_base64 = "FLAG" base64

    condition:
        any of them
}
