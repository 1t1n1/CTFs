const db = require('../config/database');

class Order {
    constructor(data = {}) {
        this.id = data.id ?? null;
        this.userId = data.userId ?? '';
        this.orderDate = data.orderDate ?? new Date();
        this.totalAmount = data.totalAmount ?? 0.00;
        this.status = data.status ?? 'Pending';
        this.shippingAddress = data.shippingAddress ?? '';
        this.orderItems = data.orderItems ?? [];
        this.items = data.items ?? null;
    }

    static async findByUserId({ userId }) {
        const rows = await db.all(`
            SELECT o.*, GROUP_CONCAT(p.name || ' x' || oi.quantity) as items
            FROM orders o
            LEFT JOIN orderItems oi ON o.id = oi.orderId
            LEFT JOIN products p ON oi.productId = p.id
            WHERE o.userId = ?
            GROUP BY o.id
            ORDER BY o.orderDate DESC
        `, [userId]);
        return rows.map(row => new Order(row));
    }

    static async findById({ orderId }) {
        const row = await db.get(
            'SELECT * FROM orders WHERE id = ?',
            [orderId]
        );
        return row ? new Order(row) : null;
    }

    static async create(orderData) {
        const crypto = require('crypto');
        const id = crypto.randomBytes(4).toString('hex');
        await db.run(
            'INSERT INTO orders (id, userId, totalAmount, shippingAddress, status) VALUES (?, ?, ?, ?, ?)',
            [id, orderData.userId, orderData.totalAmount, orderData.shippingAddress || '', orderData.status || 'Pending']
        );
        return new Order({ id, ...orderData });
    }

    toJSON() {
        return {
            id: this.id,
            userId: this.userId,
            orderDate: this.orderDate,
            totalAmount: this.totalAmount,
            status: this.status,
            shippingAddress: this.shippingAddress,
            orderItems: this.orderItems
        };
    }
}

class OrderItem {
    constructor(data = {}) {
        this.id = data.id ?? null;
        this.orderId = data.orderId ?? null;
        this.productId = data.productId ?? null;
        this.quantity = data.quantity ?? 1;
        this.price = data.price ?? 0.00;

        // Product data if joined
        this.name = data.name;
        this.description = data.description;
    }

    static async findByOrderId({ orderId }) {
        const rows = await db.all(`
            SELECT oi.*, p.name, p.description
            FROM orderItems oi
            JOIN products p ON oi.productId = p.id
            WHERE oi.orderId = ?
        `, [orderId]);
        return rows.map(row => new OrderItem(row));
    }

    static async create(orderItemData) {
        const result = await db.run(
            'INSERT INTO orderItems (orderId, productId, quantity, price) VALUES (?, ?, ?, ?)',
            [orderItemData.orderId, orderItemData.productId, orderItemData.quantity, orderItemData.price]
        );
        return new OrderItem({ id: result.id, ...orderItemData });
    }

    toJSON() {
        return {
            id: this.id,
            orderId: this.orderId,
            productId: this.productId,
            quantity: this.quantity,
            price: this.price
        };
    }
}

module.exports = { Order, OrderItem };