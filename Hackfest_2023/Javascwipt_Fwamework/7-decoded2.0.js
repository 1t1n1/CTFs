const UwU = '5765616c4b6579215765616c4b6579215765616c4b6579215765616c4b657921'; //Secwet:'cbb503325e40f76678068e42b8a31d971d1bb37dbbeb39f7758ebd2ee7cb0598';

function hexToUint8Array(hexString) {
    const OwO = new Uint8Array(hexString.length / 2);
    for (let i = 0; i < hexString.length; i += 2) {
        OwO[i / 2] = parseInt(hexString.substring(i, i + 2), 16);
    }
    return OwO;
}

async function decwypt(encwyptedTextHex, UwU) {
    const iv = hexToUint8Array(encwyptedTextHex.substring(0, 32));
    const encwyptedData = hexToUint8Array(encwyptedTextHex.substring(32));

    const UwUBuffer = hexToUint8Array(UwU);
    const x3 = await crypto.subtle.importKey(
        'raw',
        UwUBuffer,
        { name: 'AES-CBC', length: 256 },
        false,
        ['decrypt']
    );

    try {
        const decwyptedData = await crypto.subtle.decrypt(
            {
                name: "AES-CBC",
                iv: iv
            },
            x3,
            encwyptedData
        );

        const decoder = new TextDecoder();
        return decoder.decode(decwyptedData);
    } catch (e) {
        return 'OwO what\\'s this decwyption did a woopsie';
    }
}

(async function() {
    const ciphertext = '3877e41b75f60fe872402f6b334312f9617ca298ed5e9939a8d2e812456696b83b5435213f6715dfedbd11f92bcf2eada760e4cd9043062a189c93a655fd0e82';
    const decwyptedText = await decwypt(ciphertext, UwU);
    console.log('Fwag >_<:', decwyptedText);
})();