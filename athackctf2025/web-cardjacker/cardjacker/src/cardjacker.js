// ---------------------------------------------------------------------------------------------------------------------
// Imports
// ---------------------------------------------------------------------------------------------------------------------
const express = require('express');
const nunjucks = require('nunjucks');
const {query, body, validationResult} = require('express-validator');
const path = require('path');
const crypto = require('crypto');
const fs = require('fs');
const axios = require('axios');

// ---------------------------------------------------------------------------------------------------------------------
// Express app initialization
// ---------------------------------------------------------------------------------------------------------------------

const app = express();

// Serve static files from the 'public' directory
app.use(express.static(path.join(__dirname, '../public')));

app.use(express.urlencoded({extended: true}));

const config = {prefs: {mode: 'dark'}, saveToPdfPort: 1984};

// ---------------------------------------------------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------------------------------------------------

const STORAGE_PATH = process.env.STORAGE_PATH || path.join(__dirname, '../../storage');
const CARD_ID_LENGTH = 32;

// ---------------------------------------------------------------------------------------------------------------------
// Template engine configs
// ---------------------------------------------------------------------------------------------------------------------

const env = nunjucks.configure(path.join(__dirname, 'views'), {
    autoescape: true,
    express: app
});

function renderUrlAsAvatar(url) {
    const imgUrl = isValidHttpUrl(url) ? url : `https://i.pravatar.cc/400?u=${Math.floor(Math.random() * 1000) + 1}`;
    return `<div style="border-radius: 50%;overflow: hidden;width: 100%;"><img src="${imgUrl}" style="width: 100%; height: 100%; object-fit: cover;"/></div>`
}

env.addFilter('renderUrlAsAvatar', renderUrlAsAvatar);

app.set('view engine', 'html');

// ---------------------------------------------------------------------------------------------------------------------
// Utilities
// ---------------------------------------------------------------------------------------------------------------------

function resolveCardFilename(cardId, ext = "html") {
    return path.join(STORAGE_PATH, `${cardId}.${ext}`);
}

function deriveCardId(params) {
    const concatenated = params.join('');
    const hash = crypto.createHash('sha256').update(concatenated).digest('hex');
    return hash.substring(0, CARD_ID_LENGTH);
}

// ---------------------------------------------------------------------------------------------------------------------
// Validators
// ---------------------------------------------------------------------------------------------------------------------

function isValidHttpUrl(url) {
    if (typeof url !== "string") return false;
    try {
        const parsedUrl = new URL(url);
        return parsedUrl.protocol === "http:" || parsedUrl.protocol === "https:";
    } catch (e) {
        return false;
    }
}


function isValidCardId(cardId) {
    if (typeof cardId !== "string") return false;
    return cardId.length === CARD_ID_LENGTH &&
        /^[0-9a-f]+$/.test(cardId) &&
        fs.existsSync(resolveCardFilename(cardId, "html"));
}

const isTextContainingOnlyAlphabeticAndSpaces = (text) => {
    if (typeof text !== "string") return false;
    if (/^[a-zA-Z\s]+$/.test(text) === false) {
        throw new Error('');
    }
    return true;
}

function isLikelyValidGitHubHandle(username) {
    if (typeof username !== "string") return false;
    const regex = /^[a-zA-Z0-9]([a-zA-Z0-9-]{0,37}[a-zA-Z0-9])?$/;
    return regex.test(username);
}

// ---------------------------------------------------------------------------------------------------------------------
// Routes
// ---------------------------------------------------------------------------------------------------------------------

app.get('/', (req, res) => {
    res.redirect(`/create-card`);
});

app.get('/', (req, res) => {
    res.redirect(`/create-card`);
});

app.get('/create-card', (req, res) => {
    res.render('create-card.twig', {config: config});
});

app.post('/set-config', [
        body('config')
            .trim()
            .isString()
            .notEmpty()
            .withMessage('Config must be a non-empty string.'),
        body('key')
            .trim()
            .isString()
            .notEmpty()
            .withMessage('Config must be a non-empty string.'),
        body('val')
            .trim()
            .isString()
            .notEmpty()
            .withMessage('Config must be a non-empty string.'),
    ],
    async (req, res) => {
        const errors = validationResult(req);
        if (!errors.isEmpty()) {
            return res.status(400).json({errors: errors.array()});
        }
        const cfg = req.body.config;
        if (cfg in config) {
            const key = req.body.key;
            const val = req.body.val;
            config[cfg][key] = val;
            res.json({message: `Mode set to ${val}`});
        } else {
            return res.status(400).json({errors: [`Invalid configuration ${val}`]});
        }
    });

app.post('/create-card', [
        body('firstName')
            .notEmpty()
            .withMessage('First name is required')
            .custom(isTextContainingOnlyAlphabeticAndSpaces)
            .withMessage('First name is invalid'),
        body('lastName').notEmpty()
            .withMessage('Last name is required')
            .custom(isTextContainingOnlyAlphabeticAndSpaces)
            .withMessage('Last name is invalid'),
        body('email')
            .isEmail()
            .withMessage('Invalid email address'),
        body('github').notEmpty()
            .withMessage('Github handle is required')
            .custom(isLikelyValidGitHubHandle)
            .withMessage('Invalid Github handle'),
        body('company').notEmpty()
            .withMessage('Company is required')
            .custom(isTextContainingOnlyAlphabeticAndSpaces)
            .withMessage('Company is invalid'),
    ],
    async (req, res) => {
        const errors = validationResult(req);
        if (!errors.isEmpty()) {
            return res.status(400).render('error.twig', {
                config: config,
                success: false,
                message: "Validation failed",
                errors: errors.array()
            });
        }

        const userData = {};
        const userAttribs = ['firstName', 'lastName', 'email', 'github']
        for (const attr of userAttribs) {
            userData[attr] = req.body[attr];
        }
        const companyAttribs = ['company']
        const companyData = {};
        for (const attr of companyAttribs) {
            companyData[attr] = req.body[attr];
        }

        userData.getAvatarUrl = async function () {
            // Lazy loading pattern
            if (typeof this.avatarUrl == 'undefined') {
                // Trick: checking if the Github handle is valid by verifying the existence of a profile pic
                let isValidGithubHandle;
                try {
                    const githubProfilePicUrl = `https://github.com/${this.github}.png`;
                    isValidGithubHandle = (await axios.get(githubProfilePicUrl, {maxRedirects: 2})).status === 200;
                } catch {
                    isValidGithubHandle = false;
                }
                this.avatarUrl = isValidGithubHandle ? `https://github.com/${this.github}.png` : 'undefined';
            }
            return this.avatarUrl;
        };

        const cardId = deriveCardId([
            userData.firstName,
            userData.lastName,
            userData.email,
            await userData.getAvatarUrl(),
        ]);

        let html;
        try {
            html = env.render('card.twig', {
                userData: userData,
                userAvatarUrl: await userData.getAvatarUrl(),
                companyData: companyData
            });
        } catch (err) {
            console.error('Error rendering card:', err);
            return res.status(500).render('error.twig', {
                config: config,
                success: false,
                message: 'Error rendering card',
                errors: []
            });
        }

        // Save rendered HTML to a file
        const filePath = resolveCardFilename(cardId, "html");
        try {
            fs.writeFileSync(filePath, html);
            res.redirect(`/view-card?cardId=${cardId}&email=${userData.email}`);
        } catch (err) {
            console.error('Error writing file:', err);
            return res.status(500).render('error.twig', {
                config: config,
                success: false,
                message: 'Error saving card',
                errors: []
            });
        }

    }
);

app.get('/view-card', [
        query('cardId')
            .notEmpty()
            .withMessage('Query param cardId is required')
            .custom(isValidCardId)
            .withMessage('Invalid cardId'),
        query('email')
            .notEmpty()
            .withMessage('Email address is required')
            .isEmail()
            .withMessage('Invalid email address')
    ],
    (req, res) => {

        const errors = validationResult(req);
        if (!errors.isEmpty()) {
            return res.status(400).render('error.twig', {
                success: false,
                message: "Validation failed",
                errors: errors.array()
            });
        }

        const cardId = req.query.cardId;

        res.render('card-viewer.twig', {
            config: config,
            cardId: cardId,
            email: req.query.email,
        });
    });

app.get('/preview-card', [
    query('cardId')
        .notEmpty()
        .withMessage('Query param cardId is required')
        .custom(isValidCardId)
        .withMessage('Invalid cardId'),
], async (req, res) => {

    const errors = validationResult(req);
    if (!errors.isEmpty()) {
        return res.status(400).render('error.twig', {
            config: config,
            success: false,
            message: "Validation failed",
            errors: errors.array()
        });
    }

    const cardId = req.query.cardId;

    const filePath = resolveCardFilename(cardId, "html");

    try {

        // Read the file as text
        const fileContent = fs.readFileSync(filePath, 'utf-8');

        // Set the Content-Type header to indicate that the response contains HTML
        res.setHeader('Content-Type', 'text/html');
        // Send the file content as the response body
        res.send(fileContent);
    } catch (error) {
        res.status(500).render('error.twig', {
            config: config,
            message: "Internal Server Error",
            code: 500,
        });
    }
});

app.get('/download-card', [
    query('cardId')
        .notEmpty()
        .withMessage('Query param cardId is required')
        .custom(isValidCardId)
        .withMessage('Invalid cardId'),
    query('email')
        .notEmpty()
        .withMessage('Email address is required')
        .isEmail()
        .withMessage('Invalid email address'),
], async (req, res) => {

    const errors = validationResult(req);
    if (!errors.isEmpty()) {
        return res.status(400).render('error.twig', {
            config: config,
            success: false,
            message: "Validation failed",
            errors: errors.array()
        })
    }

    const cardId = req.query.cardId;
    const email = req.query.email;
    const cardPdf = resolveCardFilename(cardId, "pdf");

    if (!fs.existsSync(cardPdf)) {
        try {
            const url = `http://localhost:${parseInt(config.saveToPdfPort)}/make-card-pdf?data=${email}%20${cardId}`;
            const response = await axios.get(url, {responseType: "arraybuffer"});
            fs.writeFileSync(cardPdf, response.data);
        } catch (error) {
            console.error(error.message)
            res.status(500).render('error.twig', {
                config: config,
                message: "Error generating PDF",
                code: 500,
            });
            return;
        }
    }
    res.sendFile(cardPdf);
});

// ---------------------------------------------------------------------------------------------------------------------
// Starting app
// ---------------------------------------------------------------------------------------------------------------------

const PORT = 2025;

app.listen(PORT, () => {
    console.log(`cardjacker is listening at http://localhost:${PORT}`);
});