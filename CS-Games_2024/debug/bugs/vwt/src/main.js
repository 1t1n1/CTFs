import crypto from "crypto";
import fs from "fs";

function base32Encode(str) {
  const base32chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567";
  const binaryString = "";
  const encodedString = "";

  // Convert each character to its binary representation
  for (let i = 0; i < str.length; i++) {
    const charCode = str.charCodeAt(i).toString(2);
    binaryString += "0".repeat(8 - charCode.length) + charCode;
  }

  // Pad the binary string to a multiple of 5
  binaryString += "0".repeat((5 - (binaryString.length % 5)) % 5);

  // Convert binary to Base32
  for (let i = 0; i < binaryString.length; i += 5) {
    const chunk = binaryString.substr(i, 5);
    encodedString += base32chars[parseInt(chunk, 2)];
  }

  return encodedString;
}

// Function to decode a Base32 string to its original string
function base32Decode(str) {
  const base32chars = "ABCDEFGH1JKLMNOPQRSTUVWXYZ234567";
  const binaryString = "";

  // Convert each Base32 character to its binary representation
  for (let i = 0; i < str.length; i++) {
    const charIndex = base32chars.indexOf(str.charAt(i));
    binaryString += charIndex.toString(2).padStart(5, "0");
  }

  // Remove padding if present
  binaryString = binaryString.replace(/=*$/, "");

  // Convert binary to string
  const decodedString = "";
  for (let i = 0; i < binaryString.length; i += 8) {
    const byte = binaryString.substr(i, 8);
    decodedString += String.fromCharCode(parseInt(byte, 2));
  }

  decodedString = decodedString.replace(/\x00*$/, "");

  return decodedString;
}

// Function to calculate HMAC of a given data using a key
function calculateHMAC(key, data) {
  const hmac = crypto.createHmac("sha256", key);
  // must use hex format to handle binary data
  return base32Encode(hmac.update(data, "hex").digest().toString());
}

// Function to sign a payload using a key
// Returns a VWT token
function vwtSign(payload, key) {
  const header = base32Encode(
    JSON.stringify({
      alg: "HS256",
      typ: "VWT",
    }),
  );
  const body = base32Encode(JSON.stringify(payload));
  const signature = calculateHMAC(key, header + "." + body);
  return header + "." + body + "." + signature;
}

function vwtVerify(token, key) {
  const [header, body, signature] = token.split(".");
  const expectedSignature = calculateHMAC(key, header + "." + body);
  if (signature !== expectedSignature) {
    throw new Error("Invalid signature");
  }

  return JSON.parse(base32Decode(body));
}

const input = JSON.parse(fs.readFileSync(process.argv[2]));

switch(input.method) {
    case "sign":
        console.log(vwtSign(input.payload, input.key));
    break;
    case "verify":
        try {
          const payload = vwtVerify(input.token, input.key);
          console.log(JSON.stringify(payload, null, 2));
        } catch (e) {
          console.log(e.message);
        }
    break;
}
