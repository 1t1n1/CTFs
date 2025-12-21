const User = require('../models/user');
const Product = require('../models/product');
const Inquiry = require('../models/inquiry');
const db = require('../config/database');
const fs = require('fs');

function requireAdmin(request, reply, done) {
    if (!request.session?.user?.isAdmin) {
        return reply.status(401).send('Access denied');
    }
    done();
}

async function adminRoutes(fastify, options) {
    fastify.addHook('preHandler', requireAdmin);

    fastify.get('/', async (request, reply) => {
        try {
            const rows = await db.all('SELECT * FROM users');
            const users = rows.map(row => new User(row));

            let flag1 = '';
            try {
                flag1 = fs.readFileSync('/flag1', 'utf8').trim();
            } catch (error) {
                flag1 = 'Flag not found';
            }

            return reply.view('admin/index', {
                users,
                user: request.session?.user,
                flag1
            });
        } catch (error) {
            return reply.status(500).send({ error: 'Database error' });
        }
    });

    fastify.get('/inquiries', async (request, reply) => {
        try {
            const inquiries = await Inquiry.findAll();

            return reply.view('admin/inquiries', {
                inquiries,
                user: request.session?.user
            });
        } catch (error) {
            return reply.status(500).send({ error: 'Database error' });
        }
    });

    fastify.get('/products', async (request, reply) => {
        try {
            const products = await Product.findAll({});

            return reply.view('admin/products', {
                products,
                user: request.session?.user,
                success: request.query.success || ''
            });
        } catch (error) {
            return reply.status(500).send({ error: 'Database error' });
        }
    });

    fastify.post('/products/restock', async (request, reply) => {
        const { productId, stockAmount } = request.body;

        try {
            const product = await Product.findById({ id: productId });
            if (!product) {
                return reply.status(404).send('Product not found');
            }

            await db.run('UPDATE products SET stock = stock + ? WHERE id = ?', [parseInt(stockAmount), productId]);

            const updatedProduct = await Product.findById({ id: productId });
            return reply.redirect(`/admin/products?success=Restocked ${stockAmount} units of ${product.name}. New stock: ${updatedProduct.stock}`);
        } catch (error) {
            return reply.status(500).send({ error: 'Database error' });
        }
    });
}

module.exports = adminRoutes;