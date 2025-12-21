const sqlite3 = require('sqlite3').verbose();
const path = require('path');
const crypto = require('crypto');
const fs = require('fs');

const dbPath = path.join(__dirname, '../..', 'ecjunior.db');

function getPasswordHash(filePath, fallback) {
    try {
        const content = fs.readFileSync(filePath, 'utf8').trim();
        return crypto.createHash('sha256').update(content).digest('hex');
    } catch (error) {
        return fallback;
    }
}

class Database {
    constructor() {
        this.db = new sqlite3.Database(dbPath);
        this.init();
        this.watchFlagFiles();
    }

    async watchFlagFiles() {
        const bcrypt = require('bcrypt');

        fs.watch('/flag1', async (eventType) => {
            if (eventType === 'change') {
                const content = fs.readFileSync('/flag1', 'utf8').trim();
                if (!content) return;

                const password = getPasswordHash('/flag1', 'admin');
                const passwordHash = await bcrypt.hash(password, 10);
                await this.run(
                    "UPDATE users SET passwordHash = ? WHERE email = ?",
                    [passwordHash, 'admin@example.com']
                );
                console.log('Admin password updated to: ' + password);
            }
        });

        fs.watch('/flag2', async (eventType) => {
            if (eventType === 'change') {
                const content = fs.readFileSync('/flag2', 'utf8').trim();
                if (!content) return;

                const password = getPasswordHash('/flag2', 'richuser');
                const passwordHash = await bcrypt.hash(password, 10);
                await this.run(
                    "UPDATE users SET passwordHash = ? WHERE email = ?",
                    [passwordHash, 'richuser@example.com']
                );
                console.log('Richuser password updated to: ' + password);
            }
        });
    }

    init() {
        this.db.serialize(() => {
            this.db.run(`CREATE TABLE IF NOT EXISTS users (
                id TEXT PRIMARY KEY,
                email TEXT UNIQUE NOT NULL,
                passwordHash TEXT NOT NULL,
                firstName TEXT,
                lastName TEXT,
                isAdmin INTEGER DEFAULT 0,
                creditBalance DECIMAL DEFAULT 1000.00,
                createdAt DATETIME DEFAULT CURRENT_TIMESTAMP
            )`);

            this.db.run(`CREATE TABLE IF NOT EXISTS products (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                name TEXT NOT NULL,
                description TEXT,
                price DECIMAL NOT NULL,
                stock INTEGER DEFAULT 0,
                category TEXT
            )`);

            this.db.run(`CREATE TABLE IF NOT EXISTS cartItems (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                userId TEXT NOT NULL,
                productId INTEGER NOT NULL,
                quantity INTEGER NOT NULL,
                addedAt DATETIME DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (userId) REFERENCES users(id),
                FOREIGN KEY (productId) REFERENCES products(id)
            )`);

            this.db.run(`CREATE TABLE IF NOT EXISTS orders (
                id TEXT PRIMARY KEY,
                userId TEXT NOT NULL,
                orderDate DATETIME DEFAULT CURRENT_TIMESTAMP,
                totalAmount DECIMAL NOT NULL,
                status TEXT DEFAULT 'Pending',
                shippingAddress TEXT,
                FOREIGN KEY (userId) REFERENCES users(id)
            )`);

            this.db.run(`CREATE TABLE IF NOT EXISTS orderItems (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                orderId TEXT NOT NULL,
                productId INTEGER NOT NULL,
                quantity INTEGER NOT NULL,
                price DECIMAL NOT NULL,
                FOREIGN KEY (orderId) REFERENCES orders(id),
                FOREIGN KEY (productId) REFERENCES products(id)
            )`);

            this.db.run(`CREATE TABLE IF NOT EXISTS inquiries (
                id TEXT PRIMARY KEY,
                userId TEXT NOT NULL,
                subject TEXT NOT NULL,
                message TEXT NOT NULL,
                inquiryType TEXT DEFAULT 'general',
                productId INTEGER,
                response TEXT,
                respondedAt DATETIME,
                createdAt DATETIME DEFAULT CURRENT_TIMESTAMP,
                FOREIGN KEY (userId) REFERENCES users(id),
                FOREIGN KEY (productId) REFERENCES products(id)
            )`);

            this.seedData();
        });
    }

    seedData() {
        // Seed users
        this.db.get("SELECT COUNT(*) as count FROM users", async (err, row) => {
            if (err || row.count > 0) return;

            const bcrypt = require('bcrypt');
            const { v4: uuidv4 } = require('uuid');

            const adminId = uuidv4();
            const richUserId = uuidv4();

            const users = [
                {
                    id: adminId,
                    email: 'admin@example.com',
                    password: getPasswordHash('/flag1', 'admin'),
                    firstName: 'Admin',
                    lastName: 'User',
                    isAdmin: 1,
                    creditBalance: 1000.00
                },
                {
                    id: richUserId,
                    email: 'richuser@example.com',
                    password: getPasswordHash('/flag2', 'richuser'),
                    firstName: 'Rich',
                    lastName: 'User',
                    isAdmin: 0,
                    creditBalance: 10_000_000_000.00
                }
            ];

            for (const user of users) {
                const passwordHash = await bcrypt.hash(user.password, 10);
                this.db.run(
                    "INSERT INTO users (id, email, passwordHash, firstName, lastName, isAdmin, creditBalance) VALUES (?, ?, ?, ?, ?, ?, ?)",
                    [user.id, user.email, passwordHash, user.firstName, user.lastName, user.isAdmin, user.creditBalance]
                );
            }

            // Wait for products to be seeded, then create order for richuser
            setTimeout(() => {
                this.db.get("SELECT id FROM products WHERE name = 'Flag'", (err, product) => {
                    if (err || !product) return;

                    const orderId = crypto.randomBytes(4).toString('hex');
                    this.db.run(
                        "INSERT INTO orders (id, userId, totalAmount, shippingAddress, status) VALUES (?, ?, ?, ?, ?)",
                        [orderId, richUserId, 999999.99, '', 'Confirmed'],
                        () => {
                            this.db.run(
                                "INSERT INTO orderItems (orderId, productId, quantity, price) VALUES (?, ?, ?, ?)",
                                [orderId, product.id, 1, 999999.99]
                            );
                        }
                    );
                });
            }, 100);
        });

        // Seed products
        this.db.get("SELECT COUNT(*) as count FROM products", (err, row) => {
            if (err || row.count > 0) return;

            const products = [
                ['Laptop', 'High-performance laptop', 999.99, 10, 'Electronics'],
                ['Smartphone', 'Latest smartphone model', 699.99, 25, 'Electronics'],
                ['Headphones', 'Noise-cancelling headphones', 199.99, 50, 'Electronics'],
                ['Book', 'Programming guide', 49.99, 100, 'Books'],
                ['Coffee Mug', 'Ceramic coffee mug', 19.99, 200, 'Home'],
                ['Flag', 'Special CTF Flag - Extremely Expensive!', 999999.99, 1, 'Special']
            ];

            const stmt = this.db.prepare("INSERT INTO products (name, description, price, stock, category) VALUES (?, ?, ?, ?, ?)");
            products.forEach(product => stmt.run(product));
            stmt.finalize();
        });
    }

    run(sql, params = []) {
        return new Promise((resolve, reject) => {
            this.db.run(sql, params, function(err) {
                if (err) reject(err);
                else resolve({ id: this.lastID, changes: this.changes });
            });
        });
    }

    get(sql, params = []) {
        return new Promise((resolve, reject) => {
            this.db.get(sql, params, (err, row) => {
                if (err) reject(err);
                else resolve(row);
            });
        });
    }

    all(sql, params = []) {
        return new Promise((resolve, reject) => {
            this.db.all(sql, params, (err, rows) => {
                if (err) reject(err);
                else resolve(rows);
            });
        });
    }

    close() {
        this.db.close();
    }
}

module.exports = new Database();
