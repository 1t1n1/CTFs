// ---------------------------------------------------------------------------------------------------------------------
// Imports
// ---------------------------------------------------------------------------------------------------------------------

const express = require('express');
const crypto = require('crypto');
const fs = require('fs');
const path = require('path');
const {exec} = require('child_process');
const {promisify} = require('util');

// ---------------------------------------------------------------------------------------------------------------------
// Server initialization
// ---------------------------------------------------------------------------------------------------------------------

const app = express();

const STORAGE_PATH = process.env.STORAGE_PATH || path.join(__dirname, "../../storage");

const CARD_ID_LENGTH = 32;

const config = {shell: '/bin/sh'};

// ---------------------------------------------------------------------------------------------------------------------
// Utils
// ---------------------------------------------------------------------------------------------------------------------
function resolveCardFilename(cardId, ext = "html") {
    return path.join(STORAGE_PATH, `${cardId}.${ext}`);
}

function isValidCardId(cardId) {
    return cardId !== undefined &&
        typeof cardId === 'string' &&
        cardId.length === CARD_ID_LENGTH &&
        /^[0-9a-f]+$/.test(cardId) &&
        fs.existsSync(resolveCardFilename(cardId));
}

// ---------------------------------------------------------------------------------------------------------------------
// Routes
// ---------------------------------------------------------------------------------------------------------------------
app.get('/', (req, res) => {
    res.send('The save2pdf server app is running...');
});

app.get('/make-card-pdf', async (req, res) => {

    const data = req.query.data;

    if (!data || typeof data !== 'string') {
        console.warn("Wrong arguments")
        return res.status(401).send('Wrong arguments');
    }

    const parts = data.split(' ');
    if (parts.length !== 2) {
        console.warn('Malformed data')
        return res.status(401).send('Malformed data');
    }

    const [email, cardId] = [parts[0], parts[1]];

    if (!isValidCardId(cardId)) {
        console.warn('Invalid data')
        return res.status(401).send('Invalid data');
    }

    const htmlFile = path.join(STORAGE_PATH, `${cardId}.html`);
    const pdfFile = path.join('/tmp/', `${cardId}-${crypto.randomBytes(8).toString('hex')}.pdf`);
    const cmdWkhtmltopdfCmd = `wkhtmltopdf --title '${email}' ${htmlFile} ${pdfFile}`;
    try {
        const execAsync = promisify(exec);
        await execAsync(cmdWkhtmltopdfCmd, {shell: config.shell});
        res.sendFile(pdfFile);
    } catch (pdfErr) {
        console.error(`Command failed: ${cmdWkhtmltopdfCmd}`);
        res.status(501).send(`Command failed: ${cmdWkhtmltopdfCmd}`);
    }
});

// ---------------------------------------------------------------------------------------------------------------------
// Starting the server!
// ---------------------------------------------------------------------------------------------------------------------

const PORT = 1984;

app.listen(PORT, () => {
    console.log(`save2pdf is listening at http://localhost:${PORT}`);
});
