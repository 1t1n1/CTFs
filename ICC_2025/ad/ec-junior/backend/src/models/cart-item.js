const db = require('../config/database');

class CartItem {
    constructor(data = {}) {
        this.id = data.id ?? null;
        this.userId = data.userId ?? '';
        this.productId = data.productId ?? null;
        this.quantity = data.quantity ?? 1;
        this.addedAt = data.addedAt ?? new Date();

        // Product data if joined
        this.productName = data.name ?? data.productName;
        this.productDescription = data.description ?? data.productDescription;
        this.productPrice = data.price ?? data.productPrice;
        this.productStock = data.stock ?? data.productStock;
    }

    static async findByUserId({ userId }) {
        const rows = await db.all(`
            SELECT ci.*, p.name, p.description, p.price, p.stock
            FROM cartItems ci
            JOIN products p ON ci.productId = p.id
            WHERE ci.userId = ?
        `, [userId]);
        return rows.map(row => new CartItem(row));
    }

    static async findByUserIdAndProductId({ userId, productId }) {
        const row = await db.get(
            'SELECT * FROM cartItems WHERE userId = ? AND productId = ?',
            [userId, productId]
        );
        return row ? new CartItem(row) : null;
    }

    static async findByIdAndUserId({ id, userId }) {
        const row = await db.get(`
            SELECT ci.*, p.name, p.stock
            FROM cartItems ci
            JOIN products p ON ci.productId = p.id
            WHERE ci.id = ? AND ci.userId = ?
        `, [id, userId]);
        return row ? new CartItem(row) : null;
    }

    static async create(cartItemData) {
        const result = await db.run(
            'INSERT INTO cartItems (userId, productId, quantity) VALUES (?, ?, ?)',
            [cartItemData.userId, cartItemData.productId, cartItemData.quantity]
        );
        return new CartItem({ id: result.id, ...cartItemData });
    }

    static async updateQuantity({ id, quantity }) {
        await db.run('UPDATE cartItems SET quantity = ? WHERE id = ?', [quantity, id]);
    }

    static async incrementQuantity({ id, amount }) {
        await db.run('UPDATE cartItems SET quantity = quantity + ? WHERE id = ?', [amount, id]);
    }

    static async deleteById({ id, userId }) {
        await db.run('DELETE FROM cartItems WHERE id = ? AND userId = ?', [id, userId]);
    }

    static async deleteByUserId({ userId }) {
        await db.run('DELETE FROM cartItems WHERE userId = ?', [userId]);
    }

    toJSON() {
        return {
            id: this.id,
            userId: this.userId,
            productId: this.productId,
            quantity: this.quantity,
            addedAt: this.addedAt
        };
    }
}

module.exports = CartItem;