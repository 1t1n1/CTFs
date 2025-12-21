const Product = require('../models/product');

async function homeRoutes(fastify, options) {
    fastify.get('/', async (request, reply) => {
        const { search = '', category = '', sort = '', order = 'ASC' } = request.query;

        try {
            const validatedSort = /^price|name|stock$/i.test(sort) ? sort : '';
            const validatedOrder = /^ASC|DESC$/i.test(order) ? order : 'ASC';

            let orderBy = '';
            if (validatedSort) {
                orderBy = `${validatedSort} ${validatedOrder}`;
            }

            const products = await Product.findAll({ search, category, orderBy });
            const categories = await Product.getCategories();

            return reply.view('home/index', {
                products,
                categories,
                search,
                category,
                sort: validatedSort,
                order: validatedOrder,
                user: request.session?.user
            });
        } catch (error) {
            return reply.status(500).send({ error: 'Database error' });
        }
    });

    fastify.get('/product/:id', async (request, reply) => {
        const { id } = request.params;

        try {
            const product = await Product.findById({ id });

            if (!product) {
                return reply.status(404).send('Product not found');
            }

            return reply.view('home/product', {
                product,
                user: request.session?.user
            });
        } catch (error) {
            return reply.status(500).send({ error: 'Database error' });
        }
    });
}

module.exports = homeRoutes;