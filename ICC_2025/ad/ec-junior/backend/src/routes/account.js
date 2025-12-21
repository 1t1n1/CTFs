const bcrypt = require('bcrypt');
const User = require('../models/user');

async function accountRoutes(fastify, options) {
    fastify.get('/register', async (request, reply) => {
        return reply.view('account/register', { 
            user: request.session?.user,
            errors: []
        });
    });

    fastify.post('/register', async (request, reply) => {
        const { email, password } = request.body;
        const errors = [];

        if (!email || !password) {
            errors.push('Email and password are required');
        }

        if (errors.length > 0) {
            return reply.view('account/register', { errors, user: request.session?.user });
        }

        try {
            const existingUser = await User.findByEmail({ email });
            if (existingUser) {
                errors.push('User with this email already exists');
                return reply.view('account/register', { errors, user: request.session?.user });
            }

            const passwordHash = await bcrypt.hash(password, 10);

            const user = await User.create({
                ...request.body,
                email,
                passwordHash,
                creditBalance: 1000.00
            });

            request.session.user = {
                id: user.id,
                email: user.email,
                firstName: user.firstName,
                lastName: user.lastName,
                isAdmin: user.isAdmin
            };

            return reply.redirect('/');
        } catch (error) {
            errors.push('Registration failed');
            return reply.view('account/register', { errors, user: request.session?.user });
        }
    });

    fastify.get('/login', async (request, reply) => {
        return reply.view('account/login', { 
            user: request.session?.user,
            errors: [],
            returnUrl: request.query.returnUrl || ''
        });
    });

    fastify.post('/login', async (request, reply) => {
        const { email, password, returnUrl = '' } = request.body;
        const errors = [];

        try {
            const user = await User.findByEmail({ email });

            if (!user || !await bcrypt.compare(password, user.passwordHash)) {
                errors.push('Invalid login attempt');
                return reply.view('account/login', { errors, returnUrl, user: request.session?.user });
            }

            request.session.user = {
                id: user.id,
                email: user.email,
                firstName: user.firstName,
                lastName: user.lastName,
                isAdmin: user.isAdmin
            };

            if (returnUrl && returnUrl.startsWith('/')) {
                return reply.redirect(returnUrl);
            }
            return reply.redirect('/');
        } catch (error) {
            errors.push('Login failed');
            return reply.view('account/login', { errors, returnUrl, user: request.session?.user });
        }
    });

    fastify.get('/logout', async (request, reply) => {
        request.session.destroy();
        return reply.redirect('/');
    });

    fastify.get('/profile', async (request, reply) => {
        if (!request.session?.user) {
            return reply.redirect('/account/login');
        }

        try {
            const user = await User.findById({ id: request.session.user.id });
            const { Order } = require('../models/order');
            const orders = await Order.findByUserId({ userId: request.session.user.id });

            return reply.view('account/profile', {
                user: request.session.user,
                userProfile: user,
                orders
            });
        } catch (error) {
            return reply.status(500).send({ error: 'Database error' });
        }
    });

    fastify.get('/editprofile', async (request, reply) => {
        if (!request.session?.user) {
            return reply.redirect('/account/login');
        }

        try {
            const user = await User.findById({ id: request.session.user.id });
            return reply.view('account/editprofile', {
                user: request.session.user,
                userProfile: user,
                errors: []
            });
        } catch (error) {
            return reply.status(500).send({ error: 'Database error' });
        }
    });

    fastify.post('/editprofile', async (request, reply) => {
        if (!request.session?.user) {
            return reply.redirect('/account/login');
        }

        const { firstName, lastName, email } = request.body;
        const errors = [];

        try {
            await User.updateById({
                id: request.session.user.id,
                firstName: firstName || '',
                lastName: lastName || '',
                email
            });

            request.session.user.email = email;
            request.session.user.firstName = firstName;
            request.session.user.lastName = lastName;

            return reply.redirect('/account/profile');
        } catch (error) {
            const user = await User.findById({ id: request.session.user.id });
            errors.push('Update failed');
            return reply.view('account/editprofile', {
                user: request.session.user,
                userProfile: user,
                errors
            });
        }
    });
}

module.exports = accountRoutes;