// PoW WASM Wrapper
// This file provides a high-level API for the WASM PoW solver

const POW_DIFFICULTY = 5;

class PowSolver {
    constructor() {
        this.module = null;
        this.ready = false;
        this.difficulty = POW_DIFFICULTY;
        this.prefix = '0'.repeat(this.difficulty);
    }

    async init() {
        if (this.ready) return;

        try {
            // Try to load WASM module
            this.module = await createPowModule();
            this.ready = true;
            console.log('[PoW] WASM module loaded successfully');
        } catch (e) {
            console.warn('[PoW] Failed to load WASM module, falling back to JS:', e);
            this.ready = false;
        }
    }

    async solvePoW(challenge, onProgress) {
        await this.init();

        if (!this.ready || !this.module) {
            // Fallback to JavaScript implementation
            console.log('[PoW] Using JavaScript implementation');
            return this.solvePoWJS(challenge, onProgress);
        }

        console.log('[PoW] Using WASM implementation');
        return this.solvePoWWASM(challenge, onProgress);
    }

    async solvePoWWASM(challenge, onProgress) {
        const challengeLen = challenge.length;
        const batchSize = 5000000; // Very large batch to minimize JS/WASM crossing overhead
        let nonce = 0;

        // Set difficulty (number of zero hex digits, converted to bits)
        this.module._set_difficulty(this.difficulty * 4);

        // Set challenge string in WASM memory
        const challengePtr = this.module._malloc(challengeLen + 1);
        this.module.stringToUTF8(challenge, challengePtr, challengeLen + 1);
        this.module._set_challenge(challengePtr, challengeLen);
        this.module._free(challengePtr);

        while (nonce < 0x100000000) { // Keep searching until uint32 max
            const result = this.module._solve_pow(challengeLen, nonce, batchSize);

            // Check for unsigned comparison (0xFFFFFFFF becomes -1 when signed)
            if (result !== 0xFFFFFFFF && result !== -1) {
                // Found!
                console.log('[PoW WASM] Found nonce:', result);
                return result >>> 0; // Convert to unsigned
            }

            nonce += batchSize;

            if (onProgress) {
                onProgress(nonce);
            }

            // Yield to UI thread periodically
            await new Promise(resolve => setTimeout(resolve, 0));
        }

        throw new Error('PoW solution not found within uint32 range');
    }

    async solvePoWJS(challenge, onProgress) {
        let nonce = 0;
        const updateInterval = 10000;

        while (true) {
            const hash = await this.sha256(challenge + nonce);
            if (hash.startsWith(this.prefix)) {
                return nonce;
            }
            nonce++;

            if (onProgress && nonce % updateInterval === 0) {
                onProgress(nonce);
                await new Promise(resolve => setTimeout(resolve, 0));
            }
        }
    }

    async sha256(message) {
        // Use crypto.subtle if available (HTTPS/localhost)
        if (window.crypto && crypto.subtle) {
            const msgBuffer = new TextEncoder().encode(message);
            const hashBuffer = await crypto.subtle.digest('SHA-256', msgBuffer);
            const hashArray = Array.from(new Uint8Array(hashBuffer));
            const hashHex = hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
            return hashHex;
        } else if (typeof CryptoJS !== 'undefined') {
            // Fallback to CryptoJS for non-HTTPS
            return CryptoJS.SHA256(message).toString();
        } else {
            throw new Error('No SHA-256 implementation available');
        }
    }
}

// Global instance
window.powSolver = new PowSolver();
