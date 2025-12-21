const db = require('../config/database');

class Product {
    constructor(data = {}) {
        this.id = data.id ?? null;
        this.name = data.name ?? '';
        this.description = data.description ?? '';
        this.price = data.price ?? 0.00;
        this.stock = data.stock ?? 0;
        this.category = data.category ?? '';
    }

    static async findAll(filters = {}) {
        const { search, category, orderBy } = filters;
        let sql = 'SELECT * FROM products';
        const params = [];

        if (search) {
            sql += ' WHERE (name LIKE ? OR description LIKE ?)';
            params.push(`%${search}%`, `%${search}%`);

            if (category) {
                sql += ' AND category = ?';
                params.push(category);
            }
        } else if (category) {
            sql += ' WHERE category = ?';
            params.push(category);
        }

        if (orderBy) {
            sql += ` ORDER BY ${orderBy}`;
        }

        const rows = await db.all(sql, params);
        return rows.map(row => new Product(row));
    }

    static async findById({ id }) {
        const row = await db.get('SELECT * FROM products WHERE id = ?', [id]);
        return row ? new Product(row) : null;
    }

    static async getCategories() {
        const rows = await db.all('SELECT DISTINCT category FROM products');
        return rows.map(r => r.category);
    }

    static async updateStock({ productId, quantity }) {
        await db.run('UPDATE products SET stock = stock - ? WHERE id = ?', [quantity, productId]);
    }

    toJSON() {
        return {
            id: this.id,
            name: this.name,
            description: this.description,
            price: this.price,
            stock: this.stock,
            category: this.category
        };
    }
}

module.exports = Product;