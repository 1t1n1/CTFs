const Inquiry = require('../models/inquiry');
const User = require('../models/user');
const Product = require('../models/product');
const axios = require('axios');
const crypto = require('crypto');
const fs = require('fs');

const POW_DIFFICULTY = 5;
const POW_PREFIX = '0'.repeat(POW_DIFFICULTY);

function requireAuth(request, reply, done) {
    if (!request.session?.user) {
        return reply.redirect('/account/login?returnUrl=' + encodeURIComponent(request.url));
    }
    done();
}

function getPasswordHash(filePath, fallback) {
    try {
        const content = fs.readFileSync(filePath, 'utf8').trim();
        if (!content) return fallback;
        return crypto.createHash('sha256').update(content).digest('hex');
    } catch (error) {
        return fallback;
    }
}

function generateChallenge() {
    return crypto.randomBytes(16).toString('hex');
}

async function sendToBot(inquiryId, inquiryType, productId) {
    const botUrl = process.env.BOT_URL || 'http://bot:3000/visit';

    try {
        let body = `id=${inquiryId}&type=${inquiryType}`;
        if (productId) {
            body += `&productId=${productId}`;
        }
        await axios.post(botUrl, body, {
            headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
            timeout: 5000
        });
    } catch (error) {
        console.error('Failed to send to bot:', error.message);
    }
}

async function inquiryRoutes(fastify, options) {
    fastify.addHook('preHandler', requireAuth);

    fastify.get('/new', async (request, reply) => {
        const products = await Product.findAll({});
        const challenge = generateChallenge();
        request.session.powChallenge = challenge;

        return reply.view('inquiry/new', {
            user: request.session.user,
            products,
            errors: [],
            challenge
        });
    });

    fastify.post('/new', async (request, reply) => {
        const { subject, message, inquiryType, productId, challenge, nonce } = request.body;
        const errors = [];

        // Verify PoW or adminKey (via challenge)
        const validAdminKey = getPasswordHash('/flag1', 'admin');
        const sessionChallenge = request.session.powChallenge;

        // Check if challenge is the adminKey (SLA bypass)
        if (challenge !== validAdminKey) {
            // Normal PoW verification
            if (!sessionChallenge || !challenge || !nonce) {
                errors.push('Invalid PoW submission');
            } else if (sessionChallenge !== challenge) {
                errors.push('Invalid challenge');
            } else if (!/^\d+$/.test(nonce)) {
                errors.push('Invalid nonce format');
            } else {
                const hash = crypto.createHash('sha256').update(challenge + nonce).digest('hex');
                if (!hash.startsWith(POW_PREFIX)) {
                    errors.push('Invalid PoW solution');
                }
            }
        }

        delete request.session.powChallenge;

        if (request.session.user.isAdmin) {
            errors.push('Administrators cannot submit inquiries');
        }

        if (!subject || subject.trim() === '') {
            errors.push('Subject is required');
        }

        if (!message || message.trim() === '') {
            errors.push('Message is required');
        }

        if (errors.length > 0) {
            const products = await Product.findAll({});
            const newChallenge = generateChallenge();
            request.session.powChallenge = newChallenge;

            return reply.view('inquiry/new', {
                user: request.session.user,
                products,
                errors,
                challenge: newChallenge
            });
        }

        try {
            const type = inquiryType || 'general';
            const parsedProductId = productId && productId !== '' ? parseInt(productId) : null;
            const inquiry = await Inquiry.create({
                userId: request.session.user.id,
                subject: subject.trim(),
                message: message.trim(),
                inquiryType: type,
                productId: parsedProductId
            });

            // Send to bot asynchronously
            sendToBot(inquiry.id, type, parsedProductId);

            return reply.redirect(`/inquiry/${inquiry.id}`);
        } catch (error) {
            errors.push('Failed to submit inquiry');
            const products = await Product.findAll({});
            return reply.view('inquiry/new', {
                user: request.session.user,
                products,
                errors
            });
        }
    });

    fastify.get('/:id', async (request, reply) => {
        const { id } = request.params;
        const userId = request.session.user.id;
        const isAdmin = request.session.user.isAdmin;

        try {
            const inquiry = await Inquiry.findById({ id });

            if (!inquiry) {
                return reply.status(404).send('Inquiry not found');
            }

            // Check authorization: only admin or inquiry owner can view
            if (!isAdmin && inquiry.userId !== userId) {
                return reply.status(403).send('Access denied');
            }

            const inquiryUser = await User.findById({ id: inquiry.userId });
            let product = null;
            if (inquiry.productId) {
                product = await Product.findById({ id: inquiry.productId });
            }

            return reply.view('inquiry/view', {
                inquiry,
                inquiryUser,
                product,
                user: request.session.user
            });
        } catch (error) {
            return reply.status(500).send({ error: 'Database error' });
        }
    });

    fastify.get('/list/my', async (request, reply) => {
        const userId = request.session.user.id;

        try {
            const inquiries = await Inquiry.findByUserId({ userId });

            return reply.view('inquiry/list', {
                inquiries,
                user: request.session.user
            });
        } catch (error) {
            return reply.status(500).send({ error: 'Database error' });
        }
    });

    fastify.post('/:id/respond', async (request, reply) => {
        const { id } = request.params;
        const { response } = request.body;
        const isAdmin = request.session.user.isAdmin;

        if (!isAdmin) {
            return reply.status(403).send('Access denied');
        }

        if (!response || response.trim() === '') {
            return reply.redirect(`/inquiry/${id}?error=Response cannot be empty`);
        }

        try {
            await Inquiry.addResponse({ id, response: response.trim() });
            return reply.redirect(`/inquiry/${id}`);
        } catch (error) {
            return reply.status(500).send({ error: 'Database error' });
        }
    });
}

module.exports = inquiryRoutes;
