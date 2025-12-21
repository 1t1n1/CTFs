const fastify = require('fastify')({ logger: true });
const path = require('path');

async function build() {
    await fastify.register(require('@fastify/static'), {
        root: path.join(__dirname, 'public'),
        prefix: '/public/'
    });

    await fastify.register(require('@fastify/formbody'));
    await fastify.register(require('@fastify/cookie'));
    
    await fastify.register(require('@fastify/session'), {
        cookieName: 'sessionId',
        secret: crypto.randomUUID(),
        cookie: { 
            secure: false,
            httpOnly: false
        },
        maxAge: 1000 * 60 * 60 * 24
    });

    await fastify.register(require('@fastify/view'), {
        engine: {
            ejs: require('ejs')
        },
        root: path.join(__dirname, 'views')
    });

    await fastify.register(require('./routes/home'), { prefix: '/' });
    await fastify.register(require('./routes/account'), { prefix: '/account' });
    await fastify.register(require('./routes/cart'), { prefix: '/cart' });
    await fastify.register(require('./routes/admin'), { prefix: '/admin' });
    await fastify.register(require('./routes/inquiry'), { prefix: '/inquiry' });

    return fastify;
}

if (require.main === module) {
    const start = async () => {
        try {
            const app = await build();
            await app.listen({ port: 3000, host: '0.0.0.0' });
            console.log('Server running on http://localhost:3000');
        } catch (err) {
            console.error(err);
            process.exit(1);
        }
    };
    start();
}

module.exports = build;