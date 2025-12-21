const User = require('../models/user');
const Product = require('../models/product');
const CartItem = require('../models/cart-item');
const { Order, OrderItem } = require('../models/order');

function requireAuth(request, reply, done) {
    if (!request.session?.user) {
        return reply.redirect('/account/login?returnUrl=' + encodeURIComponent(request.url));
    }
    done();
}

async function cartRoutes(fastify, options) {
    fastify.addHook('preHandler', requireAuth);

    fastify.get('/', async (request, reply) => {
        const userId = request.session.user.id;

        try {
            const cartItems = await CartItem.findByUserId({ userId });
            const user = await User.findById({ id: userId });

            return reply.view('cart/index', {
                cartItems,
                user: request.session.user,
                userProfile: user,
                error: request.query.error || ''
            });
        } catch (error) {
            return reply.status(500).send({ error: 'Database error' });
        }
    });

    fastify.post('/add', async (request, reply) => {
        const { productId, quantity = 1 } = request.body;
        const userId = request.session.user.id;

        try {
            const product = await Product.findById({ id: productId });

            if (!product) {
                return reply.status(404).send('Product not found');
            }

            const existingItem = await CartItem.findByUserIdAndProductId({ userId, productId });

            let requestedQuantity = parseInt(quantity);
            if (existingItem) {
                requestedQuantity += existingItem.quantity;
            }

            if (requestedQuantity > product.stock) {
                return reply.redirect(`/product/${productId}?error=Only ${product.stock} items available in stock`);
            }

            if (existingItem) {
                await CartItem.incrementQuantity({ id: existingItem.id, amount: parseInt(quantity) });
            } else {
                await CartItem.create({
                    userId,
                    productId,
                    quantity: parseInt(quantity)
                });
            }

            return reply.redirect('/cart');
        } catch (error) {
            return reply.status(500).send({ error: 'Database error' });
        }
    });

    fastify.post('/update', async (request, reply) => {
        const { id, quantity } = request.body;
        const userId = request.session.user.id;

        try {
            const cartItem = await CartItem.findByIdAndUserId({ id, userId });

            if (!cartItem) {
                return reply.redirect('/cart');
            }

            const newQuantity = parseInt(quantity);

            if (newQuantity <= 0) {
                await CartItem.deleteById({ id, userId });
            } else if (newQuantity > cartItem.productStock) {
                return reply.redirect(`/cart?error=Only ${cartItem.productStock} items available in stock for ${cartItem.productName}`);
            } else {
                await CartItem.updateQuantity({ id, quantity: newQuantity });
            }

            return reply.redirect('/cart');
        } catch (error) {
            return reply.status(500).send({ error: 'Database error' });
        }
    });

    fastify.post('/remove', async (request, reply) => {
        const { id } = request.body;
        const userId = request.session.user.id;

        try {
            await CartItem.deleteById({ id, userId });
            return reply.redirect('/cart');
        } catch (error) {
            return reply.status(500).send({ error: 'Database error' });
        }
    });

    fastify.get('/checkout', async (request, reply) => {
        const userId = request.session.user.id;

        try {
            const cartItems = await CartItem.findByUserId({ userId });

            if (cartItems.length === 0) {
                return reply.redirect('/cart');
            }

            const user = await User.findById({ id: userId });
            const total = cartItems.reduce((sum, item) => sum + (item.productPrice * item.quantity), 0);

            return reply.view('cart/checkout', {
                cartItems,
                user: request.session.user,
                userProfile: user,
                total
            });
        } catch (error) {
            return reply.status(500).send({ error: 'Database error' });
        }
    });

    fastify.post('/placeorder', async (request, reply) => {
        const userId = request.session.user.id;

        try {
            const cartItems = await CartItem.findByUserId({ userId });

            if (cartItems.length === 0) {
                return reply.redirect('/cart');
            }

            const user = await User.findById({ id: userId });
            const totalAmount = cartItems.reduce((sum, item) => sum + (item.productPrice * item.quantity), 0);

            for (const cartItem of cartItems) {
                if (cartItem.quantity > cartItem.productStock) {
                    return reply.redirect(`/cart?error=Insufficient stock for ${cartItem.productName}. Only ${cartItem.productStock} available.`);
                }
            }

            if (user.creditBalance < totalAmount) {
                return reply.redirect(`/cart?error=Insufficient credit balance. You have $${user.creditBalance.toFixed(2)}, but the order total is $${totalAmount.toFixed(2)}.`);
            }

            const order = await Order.create({
                userId,
                totalAmount,
                shippingAddress: '',
                status: 'Confirmed'
            });

            for (const cartItem of cartItems) {
                await OrderItem.create({
                    orderId: order.id,
                    productId: cartItem.productId,
                    quantity: cartItem.quantity,
                    price: cartItem.productPrice
                });

                const product = await Product.findById({ id: cartItem.productId });
                if (product && product.name !== 'Flag') {
                    await Product.updateStock({ productId: cartItem.productId, quantity: cartItem.quantity });
                }
            }

            await User.decrementCreditBalance({ userId, amount: totalAmount });
            await CartItem.deleteByUserId({ userId });

            return reply.redirect(`/cart/orderconfirmation/${order.id}`);
        } catch (error) {
            return reply.status(500).send({ error: 'Database error' });
        }
    });

    fastify.get('/orderconfirmation/:orderId', async (request, reply) => {
        const { orderId } = request.params;

        try {
            const order = await Order.findById({ orderId });

            if (!order) {
                return reply.status(404).send('Order not found');
            }

            const orderItems = await OrderItem.findByOrderId({ orderId });

            // Check if Flag product was purchased
            const fs = require('fs');
            let flag2 = null;
            for (const item of orderItems) {
                const product = await Product.findById({ id: item.productId });
                if (product && product.name === 'Flag') {
                    try {
                        flag2 = fs.readFileSync('/flag2', 'utf8').trim();
                    } catch (error) {
                        flag2 = 'Flag not found';
                    }
                    break;
                }
            }

            return reply.view('cart/orderconfirmation', {
                order,
                orderItems,
                user: request.session?.user,
                flag2
            });
        } catch (error) {
            return reply.status(500).send({ error: 'Database error' });
        }
    });
}

module.exports = cartRoutes;